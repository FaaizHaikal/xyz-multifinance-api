package repository

import (
	"errors"
	"fmt"
	"time"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/infrastructure/redis"
	"xyz-multifinance-api/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) domain.CustomerRepository {
	return &customerRepository{db}
}

func (r *customerRepository) Create(customer *domain.Customer) error {
	customer.ID = uuid.New().String()

	if err := r.db.Create(customer).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create customer: %w", err)
	}

	customerJSON, err := utils.MarshalJSON(customer)
	if err != nil {
		return fmt.Errorf("failed to marshal customer for cache: %w", err)
	}

	// Cache for 1 hour
	redis.Set(fmt.Sprintf("customer:%s", customer.ID), customerJSON, time.Hour)
	redis.Set(fmt.Sprintf("customer_nik:%s", customer.NIK), customerJSON, time.Hour)

	return nil
}

func (r *customerRepository) FindByID(id string) (*domain.Customer, error) {
	cacheKey := fmt.Sprintf("customer:%s", id)
	if cachedCustomerJSON, err := redis.Get(cacheKey); err == nil {
		customer := &domain.Customer{}
		if err := utils.UnmarshalJSON(cachedCustomerJSON, customer); err == nil {
			return customer, nil // Cache hit
		}
		fmt.Printf("Warning: Failed to unmarshal cached customer %s: %v. Fetching from DB.\n", id, err)
	}

	customer := &domain.Customer{}
	result := r.db.First(customer, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by ID from DB: %w", result.Error)
	}

	customerJSON, err := utils.MarshalJSON(customer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal customer for cache: %w", err)
	}
	redis.Set(cacheKey, customerJSON, time.Hour)
	redis.Set(fmt.Sprintf("customer_nik:%s", customer.NIK), customerJSON, time.Hour) // Also cache by NIK

	return customer, nil
}

func (r *customerRepository) FindByNIK(nik string) (*domain.Customer, error) {
	cacheKey := fmt.Sprintf("customer_nik:%s", nik)
	if cachedCustomerJSON, err := redis.Get(cacheKey); err == nil {
		customer := &domain.Customer{}
		if err := utils.UnmarshalJSON(cachedCustomerJSON, customer); err == nil {
			return customer, nil // Cache hit
		}
		fmt.Printf("Warning: Failed to unmarshal cached customer by NIK %s: %v. Fetching from DB.\n", nik, err)
	}

	customer := &domain.Customer{}
	result := r.db.First(customer, "nik = ?", nik)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by NIK from DB: %w", result.Error)
	}

	customerJSON, err := utils.MarshalJSON(customer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal customer for cache: %w", err)
	}
	redis.Set(cacheKey, customerJSON, time.Hour)
	redis.Set(fmt.Sprintf("customer:%s", customer.ID), customerJSON, time.Hour) // Also cache by ID

	return customer, nil
}
