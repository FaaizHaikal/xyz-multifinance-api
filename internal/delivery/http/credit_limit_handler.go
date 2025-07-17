package http

import (
	"errors"
	"net/http"
	"strconv"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CreditLimitHandler struct {
	useCase usecase.CreditLimitUseCase
}

func NewCreditLimitHandler(router *gin.Engine, creditLimitUseCase usecase.CreditLimitUseCase) {
	handler := &CreditLimitHandler{useCase: creditLimitUseCase}

	v1 := router.Group("/api/v1")
	{
		v1.POST("/credit-limits", handler.SetCustomerCreditLimit)
		v1.GET("/customers/:customer_id/credit-limits", handler.GetCustomerCreditLimits)
		v1.GET("/customers/:customer_id/credit-limits/:tenor_months", handler.GetCustomerCreditLimitByTenor)
	}
}

func (h *CreditLimitHandler) SetCustomerCreditLimit(ctx *gin.Context) {
	req := new(model.SetCreditLimitRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	creditLimitRes, err := h.useCase.SetCustomerCreditLimit(req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input provided", "details": err.Error()})
		case errors.Is(err, domain.ErrNotFound): // Customer not found
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrAlreadyExists):
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, creditLimitRes)
}

func (h *CreditLimitHandler) GetCustomerCreditLimits(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	if customerID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "customer ID is required"})
		return
	}

	creditLimitsRes, err := h.useCase.GetCustomerCreditLimits(customerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) { // Customer not found
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, creditLimitsRes)
}

func (h *CreditLimitHandler) GetCustomerCreditLimitByTenor(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	tenorMonthsStr := ctx.Param("tenor_months")

	if customerID == "" || tenorMonthsStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "customer ID and tenor months are required"})
		return
	}

	tenorMonths, err := strconv.Atoi(tenorMonthsStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenor months format"})
		return
	}

	creditLimitRes, err := h.useCase.GetCustomerCreditLimitByTenor(customerID, tenorMonths)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, creditLimitRes)
}
