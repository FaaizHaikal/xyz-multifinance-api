package usecase_test

import (
	"errors"
	"testing"
	"time"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/usecase"
	"xyz-multifinance-api/test/mock"

	"go.uber.org/mock/gomock"
)

func TestCustomerUseCase_GetCustomerProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCustomerRepo := mock.NewMockCustomerRepository(ctrl)
	customerUseCase := usecase.NewCustomerUseCase(mockCustomerRepo)

	testCustomerID := "test-customer-id-123"
	testCustomer := &domain.Customer{
		ID:        testCustomerID,
		NIK:       "1234567890123456",
		FullName:  "John Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test case 1: Successful retrieval
	t.Run("success_get_customer_by_id", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)

		res, err := customerUseCase.GetCustomerProfileByID(testCustomerID)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected a response, got nil")
		}
		if res.ID != testCustomerID {
			t.Errorf("Expected customer ID %s, got %s", testCustomerID, res.ID)
		}
		// Add more assertions for other fields
	})

	// Test case 2: Customer not found
	t.Run("customer_not_found_by_id", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID("non-existent-id").Return(nil, domain.ErrNotFound).Times(1)

		_, err := customerUseCase.GetCustomerProfileByID("non-existent-id")

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test case 3: Repository error
	t.Run("repository_error_get_customer_by_id", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(nil, errors.New("db connection failed")).Times(1)

		_, err := customerUseCase.GetCustomerProfileByID(testCustomerID)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError, got %v", err)
		}
	})
}

func TestCustomerUseCase_GetCustomerProfileByNIK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCustomerRepo := mock.NewMockCustomerRepository(ctrl)
	customerUseCase := usecase.NewCustomerUseCase(mockCustomerRepo)

	testNIK := "1234567890123456"
	testCustomer := &domain.Customer{
		ID:        "some-id",
		NIK:       testNIK,
		FullName:  "Jane Doe",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test case 1: Successful retrieval by NIK
	t.Run("success_get_customer_by_nik", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByNIK(testNIK).Return(testCustomer, nil).Times(1)

		res, err := customerUseCase.GetCustomerProfileByNIK(testNIK)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected a response, got nil")
		}
		if res.NIK != testNIK {
			t.Errorf("Expected NIK %s, got %s", testNIK, res.NIK)
		}
	})

	// Test case 2: Customer not found by NIK
	t.Run("customer_not_found_by_nik", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByNIK("non-existent-nik").Return(nil, domain.ErrNotFound).Times(1)

		_, err := customerUseCase.GetCustomerProfileByNIK("non-existent-nik")

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test case 3: Repository error for NIK
	t.Run("repository_error_get_customer_by_nik", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByNIK(testNIK).Return(nil, errors.New("db error")).Times(1)

		_, err := customerUseCase.GetCustomerProfileByNIK(testNIK)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError, got %v", err)
		}
	})
}
