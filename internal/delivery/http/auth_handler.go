package http

import (
	"net/http"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	useCase *usecase.AuthUseCase
}

func NewAuthHandler(router *gin.Engine, authUseCase *usecase.AuthUseCase) {
	handler := &AuthHandler{useCase: authUseCase}

	router.POST("/api/v1/auth/login", handler.Login)
	router.POST("/api/v1/auth/refresh", handler.RefreshToken)
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req model.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	res, err := h.useCase.Login(&req)
	if err != nil {
		if err == domain.ErrInvalidInput {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid NIK or password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	var req model.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	res, err := h.useCase.RefreshToken(&req)
	if err != nil {
		if err == domain.ErrInvalidInput {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, res)
}
