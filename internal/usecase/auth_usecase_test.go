package usecase_test

import (
	"errors"
	"testing"
	"time"
	"xyz-multifinance-api/config"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/usecase"
	"xyz-multifinance-api/test/mock"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCustomerRepo := mock.NewMockCustomerRepository(ctrl)
	cfg := &config.Config{}

	authUseCase := usecase.NewAuthUseCase(mockCustomerRepo, cfg)

	// Test case 1: Successful registration
	t.Run("success_registration", func(t *testing.T) {
		req := &model.RegisterCustomerRequest{
			NIK:         "1234567890123456",
			FullName:    "Test User",
			Password:    "password123",
			LegalName:   "Test User Legal",
			BirthPlace:  "Jakarta",
			BirthDate:   "1990-01-01",
			Salary:      5000000,
			KTPPhoto:    "http://example.com/ktp.jpg",
			SelfiePhoto: "http://example.com/selfie.jpg",
		}

		mockCustomerRepo.EXPECT().FindByNIK(req.NIK).Return(nil, domain.ErrNotFound).Times(1)
		mockCustomerRepo.EXPECT().Create(gomock.Any()).Return(nil).Times(1)

		res, err := authUseCase.Register(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected a response, got nil")
		}
		if res.NIK != req.NIK {
			t.Errorf("Expected NIK %s, got %s", req.NIK, res.NIK)
		}
		if res.ID == "" {
			t.Error("Expected customer ID to be generated, got empty")
		}
	})

	// Test case 2: NIK already exists
	t.Run("nik_already_exists", func(t *testing.T) {
		req := &model.RegisterCustomerRequest{
			NIK:         "existingNIK12345",
			FullName:    "Existing User",
			Password:    "password123",
			LegalName:   "Existing Legal",
			BirthPlace:  "Bandung",
			BirthDate:   "1985-03-20",
			Salary:      6000000,
			KTPPhoto:    "http://example.com/ktp_exist.jpg",
			SelfiePhoto: "http://example.com/selfie_exist.jpg",
		}

		mockCustomerRepo.EXPECT().FindByNIK(req.NIK).Return(&domain.Customer{NIK: req.NIK}, nil).Times(1)
		mockCustomerRepo.EXPECT().Create(gomock.Any()).Times(0)

		_, err := authUseCase.Register(req)

		if !errors.Is(err, domain.ErrAlreadyExists) {
			t.Fatalf("Expected ErrAlreadyExists, got %v", err)
		}
	})

	// Test case 3: Invalid input
	t.Run("invalid_input", func(t *testing.T) {
		req := &model.RegisterCustomerRequest{
			NIK:         "short", // Invalid NIK length
			FullName:    "Invalid User",
			Password:    "password123",
			LegalName:   "Invalid Legal",
			BirthPlace:  "Bogor",
			BirthDate:   "1999-11-11",
			Salary:      1000000,
			KTPPhoto:    "http://example.com/ktp_invalid.jpg",
			SelfiePhoto: "http://example.com/selfie_invalid.jpg",
		}

		// No repository calls expected for invalid input
		mockCustomerRepo.EXPECT().FindByNIK(gomock.Any()).Times(0)
		mockCustomerRepo.EXPECT().Create(gomock.Any()).Times(0)

		_, err := authUseCase.Register(req)

		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Expected ErrInvalidInput, got %v", err)
		}
	})

	// Test case 4: Repository creation fails
	t.Run("repository_create_failure", func(t *testing.T) {
		req := &model.RegisterCustomerRequest{
			NIK:         "createfailNIK123",
			FullName:    "Fail User",
			Password:    "password123",
			LegalName:   "Fail Legal",
			BirthPlace:  "Depok",
			BirthDate:   "2000-05-05",
			Salary:      3000000,
			KTPPhoto:    "http://example.com/ktp_fail.jpg",
			SelfiePhoto: "http://example.com/selfie_fail.jpg",
		}

		mockCustomerRepo.EXPECT().FindByNIK(req.NIK).Return(nil, domain.ErrNotFound).Times(1)
		mockCustomerRepo.EXPECT().Create(gomock.Any()).Return(errors.New("db error")).Times(1) // Simulate DB error

		_, err := authUseCase.Register(req)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError, got %v", err)
		}
	})
}

func TestAuthUseCase_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCustomerRepo := mock.NewMockCustomerRepository(ctrl)
	cfg := &config.Config{
		JWTSecret:          "test-jwt-secret",
		AccessTokenExpiry:  time.Minute * 15,
		RefreshTokenExpiry: time.Hour * 24 * 7,
	}
	authUseCase := usecase.NewAuthUseCase(mockCustomerRepo, cfg)

	// Prepare a customer with a hashed password
	password := "testpassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testCustomer := &domain.Customer{
		ID:       "customer-id-123",
		NIK:      "1111111111111111",
		FullName: "Login Test User",
		Password: string(hashedPassword),
	}

	// Test case 1: Successful login
	t.Run("success_login", func(t *testing.T) {
		req := &model.LoginRequest{
			NIK:      testCustomer.NIK,
			Password: password,
		}

		mockCustomerRepo.EXPECT().FindByNIK(req.NIK).Return(testCustomer, nil).Times(1)

		res, err := authUseCase.Login(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected a response, got nil")
		}
		if res.AccessToken == "" || res.RefreshToken == "" {
			t.Error("Expected access and refresh tokens, got empty")
		}
		if res.ExpiresAt == 0 {
			t.Error("Expected expires_at to be set, got 0")
		}

		// Optional: Verify JWT token (decode and check claims)
		token, _ := jwt.ParseWithClaims(res.AccessToken, &model.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
		claims, ok := token.Claims.(*model.Claims)
		if !ok || claims.CustomerID != testCustomer.ID {
			t.Errorf("Invalid claims in access token")
		}
	})

	// Test case 2: Invalid password
	t.Run("invalid_password", func(t *testing.T) {
		req := &model.LoginRequest{
			NIK:      testCustomer.NIK,
			Password: "wrongpassword",
		}

		mockCustomerRepo.EXPECT().FindByNIK(req.NIK).Return(testCustomer, nil).Times(1)

		_, err := authUseCase.Login(req)

		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Expected ErrInvalidInput, got %v", err)
		}
	})

	// Test case 3: NIK not found
	t.Run("nik_not_found", func(t *testing.T) {
		req := &model.LoginRequest{
			NIK:      "nonexistentNIK",
			Password: "anypassword",
		}

		mockCustomerRepo.EXPECT().FindByNIK(req.NIK).Return(nil, domain.ErrNotFound).Times(1)

		_, err := authUseCase.Login(req)

		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Expected ErrInvalidInput, got %v", err)
		}
	})

	// Test case 4: Repository error
	t.Run("repository_error", func(t *testing.T) {
		req := &model.LoginRequest{
			NIK:      "repoerrorNIK",
			Password: "password123",
		}

		mockCustomerRepo.EXPECT().FindByNIK(req.NIK).Return(nil, errors.New("db error")).Times(1)

		_, err := authUseCase.Login(req)

		if !errors.Is(err, domain.ErrInternalServerError) {
			t.Fatalf("Expected ErrInternalServerError, got %v", err)
		}
	})
}

func TestAuthUseCase_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCustomerRepo := mock.NewMockCustomerRepository(ctrl)
	cfg := &config.Config{
		JWTSecret:          "test-jwt-secret",
		AccessTokenExpiry:  time.Minute * 15,
		RefreshTokenExpiry: time.Hour * 24 * 7,
	}
	authUseCase := usecase.NewAuthUseCase(mockCustomerRepo, cfg)

	testCustomerID := "refresh-cust-id-123"
	testNIK := "2222222222222222"

	// **GENERATE A VALID REFRESH TOKEN DIRECTLY IN THE TEST**
	// TODO: Move token generation to utils
	refreshTokenClaims := &model.Claims{
		CustomerID: testCustomerID,
		NIK:        testNIK,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	refreshTokenTestToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	validRefreshToken, err := refreshTokenTestToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("Failed to sign test refresh token: %v", err)
	}

	// Test case 1: Successful refresh
	t.Run("success_refresh_token", func(t *testing.T) {
		req := &model.RefreshTokenRequest{
			RefreshToken: validRefreshToken,
		}

		res, err := authUseCase.RefreshToken(req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if res == nil {
			t.Fatal("Expected a response, got nil")
		}
		if res.AccessToken == "" || res.RefreshToken == "" {
			t.Error("Expected new access and refresh tokens, got empty")
		}
		if res.ExpiresAt == 0 {
			t.Error("Expected expires_at to be set, got 0")
		}

		// Verify new access token claims
		newAccessTokenClaims := &model.Claims{}
		token, parseErr := jwt.ParseWithClaims(res.AccessToken, newAccessTokenClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if parseErr != nil || !token.Valid || newAccessTokenClaims.CustomerID != testCustomerID {
			t.Errorf("Newly generated access token is invalid or has wrong claims: %v", parseErr)
		}
	})

	// Test case 2: Invalid refresh token string
	t.Run("invalid_refresh_token_string", func(t *testing.T) {
		req := &model.RefreshTokenRequest{
			RefreshToken: "invalid.token.string",
		}

		_, err := authUseCase.RefreshToken(req)

		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Expected ErrInvalidInput, got %v", err)
		}
	})

	// Test case 3: Expired refresh token
	t.Run("expired_refresh_token", func(t *testing.T) {
		expiredRefreshTokenClaims := &model.Claims{
			CustomerID: testCustomerID,
			NIK:        testNIK,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
			},
		}
		expiredRefreshTokenTestToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredRefreshTokenClaims)
		expiredRefreshToken, err := expiredRefreshTokenTestToken.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			t.Fatalf("Failed to sign expired test refresh token: %v", err)
		}

		req := &model.RefreshTokenRequest{
			RefreshToken: expiredRefreshToken,
		}

		_, err = authUseCase.RefreshToken(req)

		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("Expected ErrInvalidInput for expired token, got %v", err)
		}
	})
}
