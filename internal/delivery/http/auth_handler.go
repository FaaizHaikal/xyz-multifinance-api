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

	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/refresh", handler.RefreshToken)
	}
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

func (h *AuthHandler) Register(ctx *gin.Context) {
	req := new(model.RegisterCustomerRequest)

	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	customerResp, err := h.useCase.Register(req)

	if err != nil {
		switch err {
		case domain.ErrInvalidInput:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input provided"})
		case domain.ErrAlreadyExists:
			ctx.JSON(http.StatusConflict, gin.H{"error": "customer with this NIK already exists"})
		default:
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, customerResp)
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
