package usecase

import (
	"fmt"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
)

type CustomerUseCase interface {
	GetCustomerProfileByID(id string) (*model.CustomerResponse, error)
	GetCustomerProfileByNIK(nik string) (*model.CustomerResponse, error)
}

type customerUseCase struct {
	repo domain.CustomerRepository
}

func NewCustomerUseCase(customerRepo domain.CustomerRepository) *customerUseCase {
	return &customerUseCase{
		repo: customerRepo,
	}
}

func (uc *customerUseCase) GetCustomerProfileByID(id string) (*model.CustomerResponse, error) {
	customer, err := uc.repo.FindByID(id)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: failed to get customer profile by ID: %v", domain.ErrInternalServerError, err)
	}

	return &model.CustomerResponse{
		ID:          customer.ID,
		NIK:         customer.NIK,
		FullName:    customer.FullName,
		LegalName:   customer.LegalName,
		BirthPlace:  customer.BirthPlace,
		BirthDate:   customer.BirthDate,
		Salary:      customer.Salary,
		KTPPhoto:    customer.KTPPhoto,
		SelfiePhoto: customer.SelfiePhoto,
		CreatedAt:   customer.CreatedAt,
		UpdatedAt:   customer.UpdatedAt,
	}, nil
}

func (uc *customerUseCase) GetCustomerProfileByNIK(nik string) (*model.CustomerResponse, error) {
	customer, err := uc.repo.FindByNIK(nik)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: failed to get customer profile by NIK: %v", domain.ErrInternalServerError, err)
	}

	return &model.CustomerResponse{
		ID:          customer.ID,
		NIK:         customer.NIK,
		FullName:    customer.FullName,
		LegalName:   customer.LegalName,
		BirthPlace:  customer.BirthPlace,
		BirthDate:   customer.BirthDate,
		Salary:      customer.Salary,
		KTPPhoto:    customer.KTPPhoto,
		SelfiePhoto: customer.SelfiePhoto,
		CreatedAt:   customer.CreatedAt,
		UpdatedAt:   customer.UpdatedAt,
	}, nil
}
