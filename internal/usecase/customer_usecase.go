package usecase

import (
	"errors"
	"fmt"
	"time"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/repository"

	"github.com/go-playground/validator/v10"
)

type CustomerUseCase struct {
	repo      *repository.CustomerRepository
	validator *validator.Validate
}

func NewCustomerUseCase(customerRepo *repository.CustomerRepository) *CustomerUseCase {
	return &CustomerUseCase{
		repo:      customerRepo,
		validator: validator.New(),
	}
}

func (uc *CustomerUseCase) Register(req *model.RegisterCustomerRequest) (*model.CustomerResponse, error) {
	if err := uc.validator.Struct(req); err != nil {
		return nil, domain.ErrInvalidInput
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid birth date format, use YYYY-MM-DD", domain.ErrInvalidInput)
	}

	customer := &domain.Customer{
		NIK:        req.NIK,
		FullName:   req.FullName,
		LegalName:  req.LegalName,
		BirthPlace: req.BirthPlace,
		BirthDate:  birthDate,
	}

	err = uc.repo.Create(customer)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil, domain.ErrAlreadyExists
		}
		return nil, fmt.Errorf("%w: failed to create customer: %v", domain.ErrInternalServerError, err)
	}

	return &model.CustomerResponse{
		ID:             customer.ID,
		NIK:            customer.NIK,
		FullName:       customer.FullName,
		LegalName:      customer.LegalName,
		BirthPlace:     customer.BirthPlace,
		BirthDate:      customer.BirthDate,
		Salary:         customer.Salary,
		KTPPhotoURL:    customer.KTPPhotoURL,
		SelfiePhotoURL: customer.SelfiePhotoURL,
		CreatedAt:      customer.CreatedAt,
		UpdatedAt:      customer.UpdatedAt,
	}, nil
}

func (uc *CustomerUseCase) GetCustomerProfileByID(id string) (*model.CustomerResponse, error) {
	customer, err := uc.repo.FindByID(id)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: failed to get customer profile by ID: %v", domain.ErrInternalServerError, err)
	}

	return &model.CustomerResponse{
		ID:             customer.ID,
		NIK:            customer.NIK,
		FullName:       customer.FullName,
		LegalName:      customer.LegalName,
		BirthPlace:     customer.BirthPlace,
		BirthDate:      customer.BirthDate,
		Salary:         customer.Salary,
		KTPPhotoURL:    customer.KTPPhotoURL,
		SelfiePhotoURL: customer.SelfiePhotoURL,
		CreatedAt:      customer.CreatedAt,
		UpdatedAt:      customer.UpdatedAt,
	}, nil
}

func (uc *CustomerUseCase) GetCustomerProfileByNIK(nik string) (*model.CustomerResponse, error) {
	customer, err := uc.repo.FindByNIK(nik)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("%w: failed to get customer profile by NIK: %v", domain.ErrInternalServerError, err)
	}

	return &model.CustomerResponse{
		ID:             customer.ID,
		NIK:            customer.NIK,
		FullName:       customer.FullName,
		LegalName:      customer.LegalName,
		BirthPlace:     customer.BirthPlace,
		BirthDate:      customer.BirthDate,
		Salary:         customer.Salary,
		KTPPhotoURL:    customer.KTPPhotoURL,
		SelfiePhotoURL: customer.SelfiePhotoURL,
		CreatedAt:      customer.CreatedAt,
		UpdatedAt:      customer.UpdatedAt,
	}, nil
}
