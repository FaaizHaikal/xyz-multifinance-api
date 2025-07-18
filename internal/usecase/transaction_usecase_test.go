package usecase_test

import (
	"errors"
	"log"
	"testing"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/repository"
	"xyz-multifinance-api/internal/usecase"
	"xyz-multifinance-api/test/mock"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Helper function to initialize an in-memory SQLite DB for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to in-memory SQLite: %v", err)
	}

	err = db.AutoMigrate(&domain.Customer{}, &domain.CreditLimit{}, &domain.Transaction{})
	if err != nil {
		t.Fatalf("Failed to auto migrate SQLite DB: %v", err)
	}

	return db
}

func TestTransactionUseCase_CreateTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := setupTestDB(t) // DB initialized once for the whole test function

	mockCacheStore := mock.NewMockCacheStore(ctrl)
	mockCacheStore.EXPECT().Get(gomock.Any()).Return("", errors.New("not found in cache")).AnyTimes()
	mockCacheStore.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockCacheStore.EXPECT().Del(gomock.Any()).Return(nil).AnyTimes()
	mockCacheStore.EXPECT().Close().Return(nil).AnyTimes()

	customerRepo := repository.NewCustomerRepository(db, mockCacheStore)
	creditLimitRepo := repository.NewCreditLimitRepository(db, mockCacheStore)
	transactionRepo := repository.NewTransactionRepository(db)

	transactionUseCase := usecase.NewTransactionUseCase(
		db,
		transactionRepo,
		customerRepo,
		creditLimitRepo,
		mockCacheStore,
	)

	contractNumberPrefix := "TRX-TEST-01"

	// Test case 1: Successful transaction
	t.Run("success_create_transaction", func(t *testing.T) {
		testCustomerID := uuid.New().String()
		testNIK := "1111111111111101"
		testTenor := 3
		initialLimit := 5000000.0

		req := model.CreateTransactionRequest{
			CustomerID:        testCustomerID,
			TenorMonths:       testTenor,
			OTRAmount:         4000000.0,
			AdminFee:          100000.0,
			InstallmentAmount: 1400000.0,
			InterestAmount:    100000.0,
			AssetName:         "Test Asset",
			ContractNumber:    contractNumberPrefix + "001",
		}
		totalCost := req.OTRAmount + req.AdminFee + req.InterestAmount

		db.Exec("DELETE FROM `transactions`")
		db.Exec("DELETE FROM `credit_limits`")
		db.Exec("DELETE FROM `customers`")

		// Pre-insert customer and  credit limit into DB
		customer := &domain.Customer{ID: testCustomerID, NIK: testNIK, FullName: "Transaction Test User"}
		creditLimit := &domain.CreditLimit{
			ID: uuid.New().String(), CustomerID: testCustomerID, TenorMonths: testTenor, LimitAmount: initialLimit,
		}
		if err := db.Create(customer).Error; err != nil {
			t.Fatalf("Failed to pre-create customer in SQLite: %v", err)
		}
		if err := db.Create(creditLimit).Error; err != nil {
			t.Fatalf("Failed to pre-create credit limit in SQLite: %v", err)
		}

		res, err := transactionUseCase.CreateTransaction(&req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected a response, got nil")
		}
		if res.ContractNumber != req.ContractNumber {
			t.Errorf("Expected contract number %s, got %s", req.ContractNumber, res.ContractNumber)
		}

		var updatedCreditLimit domain.CreditLimit
		db.First(&updatedCreditLimit, "customer_id = ? AND tenor_months = ?", testCustomerID, testTenor)
		expectedRemaining := initialLimit - totalCost
		if updatedCreditLimit.LimitAmount != expectedRemaining {
			t.Errorf("Expected remaining limit %f, got %f (deduction failed)", expectedRemaining, updatedCreditLimit.LimitAmount)
		}

		var createdTransaction domain.Transaction
		if db.First(&createdTransaction, "contract_number = ?", req.ContractNumber).Error != nil {
			t.Error("Transaction was not created in DB")
		}
		if createdTransaction.ID == "" {
			t.Error("Created transaction ID is empty")
		}
	})

	// Test case 2: Insufficient credit limit
	t.Run("insufficient_credit", func(t *testing.T) {
		testCustomerID := uuid.New().String()
		testNIK := "1111111111111102"
		testTenor := 3

		req := model.CreateTransactionRequest{
			CustomerID:        testCustomerID,
			TenorMonths:       testTenor,
			OTRAmount:         4000000.0,
			AdminFee:          100000.0,
			InstallmentAmount: 1400000.0,
			InterestAmount:    100000.0,
			AssetName:         "Test Asset",
			ContractNumber:    contractNumberPrefix + "002",
		}
		// 4.200.000
		// totalCost := req.OTRAmount + req.AdminFee + req.InterestAmount

		db.Exec("DELETE FROM `transactions`")
		db.Exec("DELETE FROM `credit_limits`")
		db.Exec("DELETE FROM `customers`")

		customer := &domain.Customer{ID: testCustomerID, NIK: testNIK, FullName: "Low Limit User"}
		lowLimit := &domain.CreditLimit{
			ID: uuid.New().String(), CustomerID: testCustomerID, TenorMonths: testTenor, LimitAmount: 1000000.0, // Lower than totalCost
		}
		if err := db.Create(customer).Error; err != nil {
			t.Fatalf("Failed to pre-create customer in SQLite: %v", err)
		}
		if err := db.Create(lowLimit).Error; err != nil {
			t.Fatalf("Failed to pre-create low limit in SQLite: %v", err)
		}

		_, err := transactionUseCase.CreateTransaction(&req)

		if !errors.Is(err, domain.ErrInsufficientCredit) {
			t.Fatalf("Expected ErrInsufficientCredit, got %v", err)
		}

		var unchangedCreditLimit domain.CreditLimit
		db.First(&unchangedCreditLimit, "customer_id = ? AND tenor_months = ?", testCustomerID, testTenor)
		if unchangedCreditLimit.LimitAmount != lowLimit.LimitAmount {
			t.Errorf("Expected limit to remain %f, got %f (rollback failed)", lowLimit.LimitAmount, unchangedCreditLimit.LimitAmount)
		}

		var noTransaction domain.Transaction
		if db.First(&noTransaction, "contract_number = ?", req.ContractNumber).Error == nil {
			t.Error("Transaction should not have been created on insufficient credit (rollback failed)")
		}
	})

	// Test case 3: Duplicate contract number
	t.Run("duplicate_contract_number_causes_rollback", func(t *testing.T) {
		// Generate UNIQUE IDs for THIS sub-test
		testCustomerID := uuid.New().String()
		testNIK := "1111111111111103"
		testTenor := 3
		initialLimit := 5000000.0

		db.Exec("DELETE FROM `transactions`")
		db.Exec("DELETE FROM `credit_limits`")
		db.Exec("DELETE FROM `customers`")

		customer := &domain.Customer{ID: testCustomerID, NIK: testNIK, FullName: "Duplicate Test User"}
		creditLimit := &domain.CreditLimit{
			ID: uuid.New().String(), CustomerID: testCustomerID, TenorMonths: testTenor, LimitAmount: initialLimit,
		}
		if err := db.Create(customer).Error; err != nil {
			t.Fatalf("Failed to pre-create customer in SQLite: %v", err)
		}
		if err := db.Create(creditLimit).Error; err != nil {
			t.Fatalf("Failed to pre-create credit limit in SQLite: %v", err)
		}

		conflictingTransaction := &domain.Transaction{
			ID:             uuid.New().String(),
			CustomerID:     uuid.New().String(),
			ContractNumber: contractNumberPrefix + "003",
			OTRAmount:      1.0, AdminFee: 0, InstallmentAmount: 1, InterestAmount: 0, AssetName: "dummy",
		}
		if err := db.Create(conflictingTransaction).Error; err != nil {
			t.Fatalf("Failed to pre-create conflicting transaction: %v", err)
		}

		req := model.CreateTransactionRequest{
			CustomerID:        testCustomerID,
			TenorMonths:       testTenor,
			OTRAmount:         4000000.0,
			AdminFee:          100000.0,
			InstallmentAmount: 1400000.0,
			InterestAmount:    100000.0,
			AssetName:         "Test Asset",
			ContractNumber:    contractNumberPrefix + "003", // Unique constraint violation
		}

		// totalCost = req.OTRAmount + req.AdminFee + req.InterestAmount

		_, err := transactionUseCase.CreateTransaction(&req)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError (due to duplicate key), got %v", err)
		}

		var originalCreditLimit domain.CreditLimit
		db.First(&originalCreditLimit, "customer_id = ? AND tenor_months = ?", testCustomerID, testTenor)
		if originalCreditLimit.LimitAmount != initialLimit {
			t.Errorf("Expected limit to roll back to %f, got %f (rollback failed on duplicate contract)", initialLimit, originalCreditLimit.LimitAmount)
		}

		var count int64
		db.Model(&domain.Transaction{}).Where("contract_number = ?", req.ContractNumber).Count(&count)
		if count != 1 {
			t.Errorf("Expected only the pre-existing transaction with contract number %s, got %d (rollback failed)", req.ContractNumber, count)
		}
	})

	// Test case 4: Customer not exist
	t.Run("customer_not_found_creation_failure", func(t *testing.T) {
		testCustomerID_NotFound := uuid.New().String()
		testTenor := 3

		db.Exec("DELETE FROM `transactions`")
		db.Exec("DELETE FROM `credit_limits`")
		db.Exec("DELETE FROM `customers`")

		req := model.CreateTransactionRequest{
			CustomerID:        testCustomerID_NotFound,
			TenorMonths:       testTenor,
			OTRAmount:         1000.0,
			AdminFee:          100.0,
			InstallmentAmount: 100.0,
			InterestAmount:    10.0,
			AssetName:         "Item",
			ContractNumber:    contractNumberPrefix + "004",
		}

		_, err := transactionUseCase.CreateTransaction(&req)
		log.Println(err)

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Expected ErrNotFound for customer, got %v", err)
		}
		var count int64
		db.Model(&domain.Transaction{}).Count(&count)
		if count != 0 {
			t.Error("No transactions should be created")
		}
		db.Model(&domain.CreditLimit{}).Count(&count)
		if count != 0 {
			t.Error("No credit limits should be changed")
		}
	})

	// Test case 5: Credit Limit not exist
	t.Run("credit_limit_not_found_creation_failure", func(t *testing.T) {
		testCustomerID := uuid.New().String()
		testNIK := "1111111111111105"
		testTenor := 3

		db.Exec("DELETE FROM `transactions`")
		db.Exec("DELETE FROM `credit_limits`")
		db.Exec("DELETE FROM `customers`")

		// Pre-insert customer but NO credit limit
		customer := &domain.Customer{ID: testCustomerID, NIK: testNIK, FullName: "No Limit User"}
		if err := db.Create(customer).Error; err != nil {
			t.Fatalf("Failed to pre-create customer in SQLite: %v", err)
		}

		req := model.CreateTransactionRequest{
			CustomerID:        testCustomerID,
			TenorMonths:       testTenor,
			OTRAmount:         1000.0,
			AdminFee:          100.0,
			InstallmentAmount: 100.0,
			InterestAmount:    10.0,
			AssetName:         "Item",
			ContractNumber:    contractNumberPrefix + "005",
		}

		_, err := transactionUseCase.CreateTransaction(&req)

		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("Expected ErrNotFound for credit limit, got %v", err)
		}
		var count int64
		db.Model(&domain.Transaction{}).Count(&count)
		if count != 0 {
			t.Error("No transactions should be created")
		}
		db.Model(&domain.CreditLimit{}).Count(&count)
		if count != 0 {
			t.Error("No credit limits should be changed")
		}
	})

	// Test case 6: Invalid input
	t.Run("invalid_input_validation", func(t *testing.T) {
		invalidReq := model.CreateTransactionRequest{
			CustomerID:        uuid.New().String(),
			TenorMonths:       3,
			OTRAmount:         1000.0,
			AdminFee:          100.0,
			InstallmentAmount: 100.0,
			InterestAmount:    10.0,
			AssetName:         "Item",
			ContractNumber:    "", // invalid
		}

		db.Exec("DELETE FROM `transactions`")
		db.Exec("DELETE FROM `credit_limits`")
		db.Exec("DELETE FROM `customers`")

		_, err := transactionUseCase.CreateTransaction(&invalidReq)

		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Expected ErrInvalidInput, got %v", err)
		}
	})
}
