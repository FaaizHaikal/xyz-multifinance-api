package repository

import (
	"errors"
	"fmt"
	"xyz-multifinance-api/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) CreateTransaction(transaction *domain.Transaction) error {
	transaction.ID = uuid.New().String()

	result := r.db.Create(transaction)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrAlreadyExists
		}

		return fmt.Errorf("failed to create transaction: %w", result.Error)
	}

	return nil
}

func (r *TransactionRepository) GetTransactionByContractNumber(contractNumber string) (*domain.Transaction, error) {
	transaction := &domain.Transaction{}

	result := r.db.Where("contract_number = ?", contractNumber).First(transaction)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get transaction by contract number: %w", result.Error)
	}

	return transaction, nil
}

func (r *TransactionRepository) GetTransactionsByCustomerID(customerID string) ([]domain.Transaction, error) {
	var transactions []domain.Transaction

	result := r.db.Where("customer_id = ?", customerID).Find(&transactions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get transactions by customer ID: %w", result.Error)
	}

	return transactions, nil
}
