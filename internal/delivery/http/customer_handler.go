package http

import (
	"net/http"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	useCase *usecase.CustomerUseCase
}

func NewCustomerHandler(router *gin.Engine, useCase *usecase.CustomerUseCase) {
	handler := &CustomerHandler{useCase: useCase}

	v1 := router.Group("/api/v1")
	{
		v1.POST("/customers", handler.RegisterCustomer)
		v1.GET("/customers/:id", handler.GetCustomerByID)
		v1.GET("/customers/nik/:nik", handler.GetCustomerByNIK)
	}
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
	id := ctx.Param("id")
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
