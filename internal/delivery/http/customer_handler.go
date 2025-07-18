package http

import (
	"net/http"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/usecase"
	"xyz-multifinance-api/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	useCase *usecase.CustomerUseCase
}

func NewCustomerHandler(router *gin.RouterGroup, useCase *usecase.CustomerUseCase) {
	handler := &CustomerHandler{useCase: useCase}

	router.POST("/customers", handler.RegisterCustomer)
	router.GET("/customers/:customer_id", handler.GetCustomerByID)
	router.GET("/customers/nik/:nik", handler.GetCustomerByNIK)
}

func (h *CustomerHandler) RegisterCustomer(ctx *gin.Context) {
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

func (h *CustomerHandler) GetCustomerByID(ctx *gin.Context) {
	id := ctx.Param("customer_id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "customer ID is required"})
		return
	}

	customerRes, err := h.useCase.GetCustomerProfileByID(id)
	if err != nil {
		if err == domain.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		} else {
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, customerRes)
}

func (h *CustomerHandler) GetCustomerByNIK(ctx *gin.Context) {
	nik := ctx.Param("nik")
	if nik == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "NIK is required"})
		return
	}

	customerRes, err := h.useCase.GetCustomerProfileByNIK(nik)
	if err != nil {
		if err == domain.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		} else {
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, customerRes)
}

func (h *CustomerHandler) GetMyCustomerProfile(ctx *gin.Context) {
	customerID, exists := middleware.GetCustomerIDFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Customer ID not found in token."})
		return
	}

	customerRes, err := h.useCase.GetCustomerProfileByID(customerID)
	if err != nil {
		if err == domain.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "customer not found (from token)"})
		} else {
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, customerRes)
}
