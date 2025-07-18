package usecase

import (
	"errors"
	"fmt"
	"time"
	"xyz-multifinance-api/config"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Login(req *model.LoginRequest) (*model.LoginResponse, error)
	Register(req *model.RegisterCustomerRequest) (*model.CustomerResponse, error)
	RefreshToken(req *model.RefreshTokenRequest) (*model.LoginResponse, error)
}

type authUseCase struct {
	customerRepo domain.CustomerRepository
	cfg          *config.Config
	validator    *validator.Validate
}

func NewAuthUseCase(customerRepo domain.CustomerRepository, cfg *config.Config) AuthUseCase {
	return &authUseCase{
		customerRepo: customerRepo,
		cfg:          cfg,
		validator:    validator.New(),
	}
}

func (uc *authUseCase) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	customer, err := uc.customerRepo.FindByNIK(req.NIK)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, fmt.Errorf("%w: invalid credentials", domain.ErrInvalidInput)
		}
		return nil, fmt.Errorf("%w: failed to retrieve customer for login: %v", domain.ErrInternalServerError, err)
	}

	// Compare password hash
	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("%w: invalid credentials", domain.ErrInvalidInput)
	}

	// Generate Access Token and Refresh Token
	accessToken, accessExpiresAt, err := uc.generateToken(customer.ID, customer.NIK, uc.cfg.AccessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate access token: %v", domain.ErrInternalServerError, err)
	}

	refreshToken, _, err := uc.generateToken(customer.ID, customer.NIK, uc.cfg.RefreshTokenExpiry) // Refresh token has no expiry check for now
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate refresh token: %v", domain.ErrInternalServerError, err)
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt.Unix(),
	}, nil
}

func (uc *authUseCase) Register(req *model.RegisterCustomerRequest) (*model.CustomerResponse, error) {
	if err := uc.validator.Struct(req); err != nil {
		return nil, domain.ErrInvalidInput
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid birth date format, use YYYY-MM-DD", domain.ErrInvalidInput)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to hash password: %v", domain.ErrInternalServerError, err)
	}

	customer := &domain.Customer{
		NIK:            req.NIK,
		FullName:       req.FullName,
		Password:       string(hashedPassword),
		LegalName:      req.LegalName,
		BirthPlace:     req.BirthPlace,
		BirthDate:      birthDate,
		Salary:         req.Salary,
		KTPPhotoURL:    req.KTPPhotoURL,
		SelfiePhotoURL: req.SelfiePhotoURL,
	}

	err = uc.customerRepo.Create(customer)
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

func (uc *authUseCase) RefreshToken(req *model.RefreshTokenRequest) (*model.LoginResponse, error) {
	// Parse and validate
	token, err := jwt.ParseWithClaims(req.RefreshToken, &model.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(uc.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, fmt.Errorf("%w: invalid or expired refresh token: %v", domain.ErrInvalidInput, err)
	}

	claims, ok := token.Claims.(*model.Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%w: invalid refresh token claims", domain.ErrInvalidInput)
	}

	// Generate a new Access Token
	newAccessToken, newAccessExpiresAt, err := uc.generateToken(claims.CustomerID, claims.NIK, uc.cfg.AccessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate new access token: %v", domain.ErrInternalServerError, err)
	}

	// New refresh token
	newRefreshToken, _, err := uc.generateToken(claims.CustomerID, claims.NIK, uc.cfg.RefreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate new refresh token: %v", domain.ErrInternalServerError, err)
	}

	return &model.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    newAccessExpiresAt.Unix(),
	}, nil
}

func (uc *authUseCase) generateToken(customerID, nik string, expiryDuration time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(expiryDuration)
	claims := &model.Claims{
		CustomerID: customerID,
		NIK:        nik,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(uc.cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}
