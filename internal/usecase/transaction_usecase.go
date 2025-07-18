package usecase

import (
	"errors"
	"fmt"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/repository"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionUseCase struct {
	db              *gorm.DB // DB instance for transaction management
	transactionRepo *repository.TransactionRepository
	customerRepo    *repository.CustomerRepository
	creditLimitRepo *repository.CreditLimitRepository
	validator       *validator.Validate
}

func NewTransactionUseCase(
	db *gorm.DB,
	transactionRepo *repository.TransactionRepository,
	customerRepo *repository.CustomerRepository,
	creditLimitRepo *repository.CreditLimitRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		db:              db,
		transactionRepo: transactionRepo,
		customerRepo:    customerRepo,
		creditLimitRepo: creditLimitRepo,
		validator:       validator.New(),
	}
}

func (uc *TransactionUseCase) CreateTransaction(req *model.CreateTransactionRequest) (*model.TransactionResponse, error) {
	if err := uc.validator.Struct(req); err != nil {
		return nil, domain.ErrInvalidInput
	}

	var createdTransaction *domain.Transaction // The transaction created in GORM

	err := uc.db.Transaction(func(tx *gorm.DB) error {
		txCustomerRepo := repository.NewCustomerRepository(tx)
		txCreditLimitRepo := repository.NewCreditLimitRepository(tx)
		txTransactionRepo := repository.NewTransactionRepository(tx)

		_, err := txCustomerRepo.FindByID(req.CustomerID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return fmt.Errorf("%w: customer with ID %s not found", domain.ErrNotFound, req.CustomerID)
			}
			return fmt.Errorf("failed to verify customer existence: %w", err)
		}

		// Retrieve Credit Limit with Pessimistic Locking
		creditLimit := &domain.CreditLimit{}
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("customer_id = ? AND tenor_months = ?", req.CustomerID, req.TenorMonths).
			First(creditLimit).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: credit limit for tenor %d not found for customer %s", domain.ErrNotFound, req.TenorMonths, req.CustomerID)
			}
			return fmt.Errorf("failed to retrieve credit limit with lock: %w", err)
		}

		totalTransactionCost := req.OTRAmount + req.AdminFee + req.InterestAmount

		// Credit Limit is reached
		if creditLimit.LimitAmount < totalTransactionCost {
			return domain.ErrInsufficientCredit
		}

		// Update Credit Limit
		creditLimit.LimitAmount -= totalTransactionCost
		err = txCreditLimitRepo.UpdateCreditLimit(creditLimit)
		if err != nil {
			return fmt.Errorf("failed to deduct credit limit: %w", err)
		}

		transaction := &domain.Transaction{
			CustomerID:        req.CustomerID,
			ContractNumber:    req.ContractNumber,
			OTRAmount:         req.OTRAmount,
			AdminFee:          req.AdminFee,
			InstallmentAmount: req.InstallmentAmount,
			InterestAmount:    req.InterestAmount,
			AssetName:         req.AssetName,
		}

		err = txTransactionRepo.CreateTransaction(transaction)
		if err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				return fmt.Errorf("%w: transaction with contract number %s already exists", domain.ErrAlreadyExists, req.ContractNumber)
			}
			return fmt.Errorf("failed to create transaction record: %w", err)
		}
		createdTransaction = transaction // Store for the outer scope

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: transaction process failed: %v", domain.ErrInternalServerError, err)
	}

	return &model.TransactionResponse{
		ID:                createdTransaction.ID,
		CustomerID:        createdTransaction.CustomerID,
		ContractNumber:    createdTransaction.ContractNumber,
		OTRAmount:         createdTransaction.OTRAmount,
		AdminFee:          createdTransaction.AdminFee,
		InstallmentAmount: createdTransaction.InstallmentAmount,
		InterestAmount:    createdTransaction.InterestAmount,
		AssetName:         createdTransaction.AssetName,
	}, nil
}

func (uc *TransactionUseCase) GetTransactionByContractNumber(contractNumber string) (*model.TransactionResponse, error) {
	transaction, err := uc.transactionRepo.GetTransactionByContractNumber(contractNumber)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: failed to get transaction by contract number: %v", domain.ErrInternalServerError, err)
	}

	return &model.TransactionResponse{
		ID:                transaction.ID,
		CustomerID:        transaction.CustomerID,
		ContractNumber:    transaction.ContractNumber,
		OTRAmount:         transaction.OTRAmount,
		AdminFee:          transaction.AdminFee,
		InstallmentAmount: transaction.InstallmentAmount,
		InterestAmount:    transaction.InterestAmount,
		AssetName:         transaction.AssetName,
	}, nil
}

func (uc *TransactionUseCase) GetTransactionsByCustomerID(customerID string) ([]model.TransactionResponse, error) {
	transactions, err := uc.transactionRepo.GetTransactionsByCustomerID(customerID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve transactions: %v", domain.ErrInternalServerError, err)
	}

	var responses []model.TransactionResponse
	for _, transaction := range transactions {
		responses = append(responses, model.TransactionResponse{
			ID:                transaction.ID,
			CustomerID:        transaction.CustomerID,
			ContractNumber:    transaction.ContractNumber,
			OTRAmount:         transaction.OTRAmount,
			AdminFee:          transaction.AdminFee,
			InstallmentAmount: transaction.InstallmentAmount,
			InterestAmount:    transaction.InterestAmount,
			AssetName:         transaction.AssetName,
		})
	}
	return responses, nil
}
