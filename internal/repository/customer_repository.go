package repository

import (
	"errors"
	"fmt"
	"xyz-multifinance-api/internal/domain"

	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db}
}

func (r *CustomerRepository) Create(customer *domain.Customer) error {
	return r.db.Create(customer).Error
}

func (r *CustomerRepository) FindByID(id string) (*domain.Customer, error) {
	customer := &domain.Customer{}

	if err := r.db.First(&customer, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by ID: %w", err)
	}

	return customer, nil
}

func (r *CustomerRepository) FindByNIK(nik string) (*domain.Customer, error) {
	customer := &domain.Customer{}

	if err := r.db.First(&customer, "nik = ?", nik).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by NIK: %w", err)
	}

	return customer, nil
}
