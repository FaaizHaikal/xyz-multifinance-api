package usecase

import (
	"errors"
	"fmt"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/repository"

	"github.com/go-playground/validator/v10"
)

type CreditLimitUseCase struct {
	creditLimitRepo repository.CreditLimitRepository
	customerRepo    repository.CustomerRepository
	validator       *validator.Validate
}

func NewCreditLimitUseCase(creditLimitRepo repository.CreditLimitRepository, customerRepo repository.CustomerRepository) *CreditLimitUseCase {
	return &CreditLimitUseCase{
		creditLimitRepo: creditLimitRepo,
		customerRepo:    customerRepo,
		validator:       validator.New(),
	}
}

func (uc *CreditLimitUseCase) SetCustomerCreditLimit(req *model.SetCreditLimitRequest) (*model.CreditLimitResponse, error) {
	if err := uc.validator.Struct(req); err != nil {
		return nil, domain.ErrInvalidInput
	}

	// Verify customer exist
	_, err := uc.customerRepo.FindByID(req.CustomerID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, fmt.Errorf("%w: customer with ID %s not found", domain.ErrNotFound, req.CustomerID)
		}
		return nil, fmt.Errorf("%w: failed to verify customer existence: %v", domain.ErrInternalServerError, err)
	}

	existingLimit, err := uc.creditLimitRepo.GetCreditLimitByCustomerAndTenor(req.CustomerID, req.TenorMonths)

	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("%w: failed to check existing credit limit: %v", domain.ErrInternalServerError, err)
	}

	creditLimit := &domain.CreditLimit{
		CustomerID:  req.CustomerID,
		TenorMonths: req.TenorMonths,
		LimitAmount: req.LimitAmount,
	}

	if existingLimit != nil {
		creditLimit.ID = existingLimit.ID // Update existing row
		err = uc.creditLimitRepo.UpdateCreditLimit(creditLimit)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to update credit limit: %v", domain.ErrInternalServerError, err)
		}
	} else {
		// Create if not exist
		err = uc.creditLimitRepo.CreateCreditLimit(creditLimit)
		if err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				return nil, fmt.Errorf("%w: credit limit for this customer and tenor already exists", domain.ErrAlreadyExists)
			}
			return nil, fmt.Errorf("%w: failed to create credit limit: %v", domain.ErrInternalServerError, err)
		}
	}

	return &model.CreditLimitResponse{
		ID:          creditLimit.ID,
		CustomerID:  creditLimit.CustomerID,
		TenorMonths: creditLimit.TenorMonths,
		LimitAmount: creditLimit.LimitAmount,
	}, nil
}

func (uc *CreditLimitUseCase) GetCustomerCreditLimits(customerID string) ([]model.CreditLimitResponse, error) {
	// Verify customer exist
	_, err := uc.customerRepo.FindByID(customerID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, fmt.Errorf("%w: customer with ID %s not found", domain.ErrNotFound, customerID)
		}
		return nil, fmt.Errorf("%w: failed to verify customer existence: %v", domain.ErrInternalServerError, err)
	}

	limits, err := uc.creditLimitRepo.GetCreditLimitsByCustomerID(customerID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve credit limits: %v", domain.ErrInternalServerError, err)
	}

	var responses []model.CreditLimitResponse
	for _, limit := range limits {
		responses = append(responses, model.CreditLimitResponse{
			ID:          limit.ID,
			CustomerID:  limit.CustomerID,
			TenorMonths: limit.TenorMonths,
			LimitAmount: limit.LimitAmount,
		})
	}
	return responses, nil
}

func (uc *CreditLimitUseCase) GetCustomerCreditLimitByTenor(customerID string, tenorMonths int) (*model.CreditLimitResponse, error) {
	// Verify customer exist
	_, err := uc.customerRepo.FindByID(customerID)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, fmt.Errorf("%w: customer with ID %s not found", domain.ErrNotFound, customerID)
		}
		return nil, fmt.Errorf("%w: failed to verify customer existence: %v", domain.ErrInternalServerError, err)
	}

	limit, err := uc.creditLimitRepo.GetCreditLimitByCustomerAndTenor(customerID, tenorMonths)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, fmt.Errorf("%w: credit limit for tenor %d not found for customer %s", domain.ErrNotFound, tenorMonths, customerID)
		}
		return nil, fmt.Errorf("%w: failed to retrieve credit limit: %v", domain.ErrInternalServerError, err)
	}

	return &model.CreditLimitResponse{
		ID:          limit.ID,
		CustomerID:  limit.CustomerID,
		TenorMonths: limit.TenorMonths,
		LimitAmount: limit.LimitAmount,
	}, nil
}
