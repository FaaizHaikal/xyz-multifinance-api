package usecase_test

import (
	"errors"
	"testing"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/usecase"
	"xyz-multifinance-api/test/mock"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestCreditLimitUseCase_SetCustomerCreditLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCreditLimitRepo := mock.NewMockCreditLimitRepository(ctrl)
	mockCustomerRepo := mock.NewMockCustomerRepository(ctrl)

	creditLimitUseCase := usecase.NewCreditLimitUseCase(mockCreditLimitRepo, mockCustomerRepo)

	testCustomerID := uuid.New().String()
	testCustomer := &domain.Customer{ID: testCustomerID, NIK: "1234567890123456"}

	// Test case 1: Successfully create a new credit limit
	t.Run("success_create_new_limit", func(t *testing.T) {
		req := &model.SetCreditLimitRequest{
			CustomerID:  testCustomerID,
			TenorMonths: 1,
			LimitAmount: 1000000,
		}

		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(testCustomerID, req.TenorMonths).Return(nil, domain.ErrNotFound).Times(1)
		mockCreditLimitRepo.EXPECT().CreateCreditLimit(gomock.Any()).Return(nil).Times(1)

		res, err := creditLimitUseCase.SetCustomerCreditLimit(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected response, got nil")
		}
		if res.CustomerID != testCustomerID || res.TenorMonths != req.TenorMonths || res.LimitAmount != req.LimitAmount {
			t.Errorf("Mismatch in response data: %+v", res)
		}
	})

	// Test case 2: Successfully update an existing credit limit
	t.Run("success_update_existing_limit", func(t *testing.T) {
		req := &model.SetCreditLimitRequest{
			CustomerID:  testCustomerID,
			TenorMonths: 1,
			LimitAmount: 1500000, // New amount
		}
		existingLimit := &domain.CreditLimit{
			ID:          "existing-limit-id",
			CustomerID:  testCustomerID,
			TenorMonths: 1,
			LimitAmount: 1000000,
		}

		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(testCustomerID, req.TenorMonths).Return(existingLimit, nil).Times(1)
		mockCreditLimitRepo.EXPECT().UpdateCreditLimit(gomock.Any()).Return(nil).Times(1)

		res, err := creditLimitUseCase.SetCustomerCreditLimit(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected response, got nil")
		}
		if res.LimitAmount != req.LimitAmount {
			t.Errorf("Expected updated limit %f, got %f", req.LimitAmount, res.LimitAmount)
		}
	})

	// Test case 3: Customer not found
	t.Run("customer_not_found", func(t *testing.T) {
		req := &model.SetCreditLimitRequest{
			CustomerID:  uuid.New().String(),
			TenorMonths: 1,
			LimitAmount: 1000000,
		}

		mockCustomerRepo.EXPECT().FindByID(req.CustomerID).Return(nil, domain.ErrNotFound).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(gomock.Any(), gomock.Any()).Times(0)

		_, err := creditLimitUseCase.SetCustomerCreditLimit(req)

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test case 4: Invalid input
	t.Run("invalid_input_tenor", func(t *testing.T) {
		req := &model.SetCreditLimitRequest{
			CustomerID:  testCustomerID,
			TenorMonths: 5, // Invalid tenor
			LimitAmount: 1000000,
		}

		mockCustomerRepo.EXPECT().FindByID(gomock.Any()).Times(0)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(gomock.Any(), gomock.Any()).Times(0)

		_, err := creditLimitUseCase.SetCustomerCreditLimit(req)

		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Expected ErrInvalidInput, got %v", err)
		}
	})

	// Test case 5: Repository error during create
	t.Run("repo_error_on_create", func(t *testing.T) {
		req := &model.SetCreditLimitRequest{
			CustomerID:  testCustomerID,
			TenorMonths: 2,
			LimitAmount: 2000000,
		}

		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(testCustomerID, req.TenorMonths).Return(nil, domain.ErrNotFound).Times(1)
		mockCreditLimitRepo.EXPECT().CreateCreditLimit(gomock.Any()).Return(errors.New("db write error")).Times(1)

		_, err := creditLimitUseCase.SetCustomerCreditLimit(req)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError, got %v", err)
		}
	})

	// Test case 6: Repository error during update
	t.Run("repo_error_on_update", func(t *testing.T) {
		req := &model.SetCreditLimitRequest{
			CustomerID:  testCustomerID,
			TenorMonths: 1,
			LimitAmount: 1500000,
		}
		existingLimit := &domain.CreditLimit{
			ID:          "existing-limit-id",
			CustomerID:  testCustomerID,
			TenorMonths: 1,
			LimitAmount: 1000000,
		}

		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(testCustomerID, req.TenorMonths).Return(existingLimit, nil).Times(1)
		mockCreditLimitRepo.EXPECT().UpdateCreditLimit(gomock.Any()).Return(errors.New("db update error")).Times(1)

		_, err := creditLimitUseCase.SetCustomerCreditLimit(req)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError, got %v", err)
		}
	})
}

func TestCreditLimitUseCase_GetCustomerCreditLimits(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCreditLimitRepo := mock.NewMockCreditLimitRepository(ctrl)
	mockCustomerRepo := mock.NewMockCustomerRepository(ctrl)

	creditLimitUseCase := usecase.NewCreditLimitUseCase(mockCreditLimitRepo, mockCustomerRepo)

	testCustomerID := "test-cust-id-get"
	testCustomer := &domain.Customer{ID: testCustomerID, NIK: "1234567890123456"}
	testLimits := []domain.CreditLimit{
		{ID: "l1", CustomerID: testCustomerID, TenorMonths: 1, LimitAmount: 100000},
		{ID: "l2", CustomerID: testCustomerID, TenorMonths: 3, LimitAmount: 500000},
	}

	// Test case 1: Successfully retrieve multiple limits
	t.Run("success_get_multiple_limits", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitsByCustomerID(testCustomerID).Return(testLimits, nil).Times(1)

		res, err := creditLimitUseCase.GetCustomerCreditLimits(testCustomerID)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected response, got nil")
		}
		if len(res) != len(testLimits) {
			t.Errorf("Expected %d limits, got %d", len(testLimits), len(res))
		}
	})

	// Test case 2: Customer not found
	t.Run("get_limits_customer_not_found", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID("non-existent").Return(nil, domain.ErrNotFound).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitsByCustomerID(gomock.Any()).Times(0)

		_, err := creditLimitUseCase.GetCustomerCreditLimits("non-existent")

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test case 3: Repository error during retrieval
	t.Run("repo_error_get_limits", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitsByCustomerID(testCustomerID).Return(nil, errors.New("db error")).Times(1)

		_, err := creditLimitUseCase.GetCustomerCreditLimits(testCustomerID)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError, got %v", err)
		}
	})
}

func TestCreditLimitUseCase_GetCustomerCreditLimitByTenor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCreditLimitRepo := mock.NewMockCreditLimitRepository(ctrl)
	mockCustomerRepo := mock.NewMockCustomerRepository(ctrl)

	creditLimitUseCase := usecase.NewCreditLimitUseCase(mockCreditLimitRepo, mockCustomerRepo)

	testCustomerID := "test-cust-id-tenor"
	testTenor := 6
	testCustomer := &domain.Customer{ID: testCustomerID, NIK: "1234567890123456"}
	testLimit := &domain.CreditLimit{
		ID: "l3", CustomerID: testCustomerID, TenorMonths: testTenor, LimitAmount: 700000,
	}

	// Test case 1: Successfully retrieve specific limit
	t.Run("success_get_specific_limit", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(testCustomerID, testTenor).Return(testLimit, nil).Times(1)

		res, err := creditLimitUseCase.GetCustomerCreditLimitByTenor(testCustomerID, testTenor)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected response, got nil")
		}
		if res.CustomerID != testCustomerID || res.TenorMonths != testTenor {
			t.Errorf("Mismatch in response data: %+v", res)
		}
	})

	// Test case 2: Customer not found
	t.Run("get_specific_limit_customer_not_found", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID("non-existent").Return(nil, domain.ErrNotFound).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(gomock.Any(), gomock.Any()).Times(0)

		_, err := creditLimitUseCase.GetCustomerCreditLimitByTenor("non-existent", testTenor)

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test case 3: Specific limit not found
	t.Run("specific_limit_not_found", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(testCustomerID, testTenor).Return(nil, domain.ErrNotFound).Times(1)

		_, err := creditLimitUseCase.GetCustomerCreditLimitByTenor(testCustomerID, testTenor)

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Expected ErrNotFound, got %v", err)
		}
	})

	// Test case 4: Repository error during retrieval
	t.Run("repo_error_get_specific_limit", func(t *testing.T) {
		mockCustomerRepo.EXPECT().FindByID(testCustomerID).Return(testCustomer, nil).Times(1)
		mockCreditLimitRepo.EXPECT().GetCreditLimitByCustomerAndTenor(testCustomerID, testTenor).Return(nil, errors.New("db error")).Times(1)

		_, err := creditLimitUseCase.GetCustomerCreditLimitByTenor(testCustomerID, testTenor)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError, got %v", err)
		}
	})
}
