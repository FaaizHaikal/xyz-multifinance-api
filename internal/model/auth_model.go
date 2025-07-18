package model

import "github.com/golang-jwt/jwt/v5"

type LoginRequest struct {
	NIK      string `json:"nik" validate:"required,len=16"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Claims struct {
	CustomerID string `json:"customer_id"`
	NIK        string `json:"nik"`
	jwt.RegisteredClaims
}
