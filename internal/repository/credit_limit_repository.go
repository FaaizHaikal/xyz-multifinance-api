package repository

import (
	"errors"
	"fmt"
	"xyz-multifinance-api/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreditLimitRepository struct {
	db *gorm.DB
}

func NewCreditLimitRepository(db *gorm.DB) *CreditLimitRepository {
	return &CreditLimitRepository{db: db}
}

func (r *CreditLimitRepository) CreateCreditLimit(creditLimit *domain.CreditLimit) error {
	creditLimit.ID = uuid.New().String()

	result := r.db.Create(creditLimit)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create credit limit: %w", result.Error)
	}
	return nil
}

func (r *CreditLimitRepository) GetCreditLimitsByCustomerID(customerID string) ([]domain.CreditLimit, error) {
	var creditLimits []domain.CreditLimit
	result := r.db.Where("customer_id = ?", customerID).Find(&creditLimits)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get credit limits by customer ID: %w", result.Error)
	}
	return creditLimits, nil
}

func (r *CreditLimitRepository) GetCreditLimitByCustomerAndTenor(customerID string, tenorMonths int) (*domain.CreditLimit, error) {
	creditLimit := &domain.CreditLimit{}
	result := r.db.Where("customer_id = ? AND tenor_months = ?", customerID, tenorMonths).First(creditLimit)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get credit limit: %w", result.Error)
	}
	return creditLimit, nil
}

func (r *CreditLimitRepository) UpdateCreditLimit(creditLimit *domain.CreditLimit) error {
	result := r.db.Save(creditLimit)
	if result.Error != nil {
		return fmt.Errorf("failed to update credit limit: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
