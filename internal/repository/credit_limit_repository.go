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

type creditLimitRepository struct {
	db *gorm.DB
}

func NewCreditLimitRepository(db *gorm.DB) domain.CreditLimitRepository {
	return &creditLimitRepository{db: db}
}

func (r *creditLimitRepository) CreateCreditLimit(creditLimit *domain.CreditLimit) error {
	creditLimit.ID = uuid.New().String()

	result := r.db.Create(creditLimit)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create credit limit: %w", result.Error)
	}

	creditLimitJSON, err := utils.MarshalJSON(creditLimit)
	if err != nil {
		return fmt.Errorf("failed to marshal credit limit for cache: %w", err)
	}

	// Cache for 1 hour
	redis.Set(fmt.Sprintf("credit_limit:%s:%d", creditLimit.CustomerID, creditLimit.TenorMonths), creditLimitJSON, time.Hour)

	return nil
}

func (r *creditLimitRepository) GetCreditLimitByCustomerAndTenor(customerID string, tenorMonths int) (*domain.CreditLimit, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("credit_limit:%s:%d", customerID, tenorMonths)
	if cachedLimitJSON, err := redis.Get(cacheKey); err == nil {
		creditLimit := &domain.CreditLimit{}
		if err := utils.UnmarshalJSON(cachedLimitJSON, creditLimit); err == nil {
			return creditLimit, nil
		}
		fmt.Printf("Warning: Failed to unmarshal cached credit limit %s:%d: %v. Fetching from DB.\n", customerID, tenorMonths, err)
	}

	creditLimit := &domain.CreditLimit{}
	result := r.db.Where("customer_id = ? AND tenor_months = ?", customerID, tenorMonths).First(creditLimit)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get credit limit from DB: %w", result.Error)
	}

	// Cache the result from DB
	creditLimitJSON, err := utils.MarshalJSON(creditLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credit limit for cache: %w", err)
	}
	redis.Set(cacheKey, creditLimitJSON, time.Hour)

	return creditLimit, nil
}

func (r *creditLimitRepository) UpdateCreditLimit(creditLimit *domain.CreditLimit) error {
	result := r.db.Save(creditLimit)
	if result.Error != nil {
		return fmt.Errorf("failed to update credit limit: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}

	// Delete cache after update
	redis.Del(fmt.Sprintf("credit_limit:%s:%d", creditLimit.CustomerID, creditLimit.TenorMonths))

	return nil
}

func (r *creditLimitRepository) GetCreditLimitsByCustomerID(customerID string) ([]domain.CreditLimit, error) {
	var creditLimits []domain.CreditLimit
	result := r.db.Where("customer_id = ?", customerID).Find(&creditLimits)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get credit limits by customer ID: %w", result.Error)
	}
	return creditLimits, nil
}
