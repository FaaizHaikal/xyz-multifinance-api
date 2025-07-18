package http

import (
	"errors"
	"net/http"
	"xyz-multifinance-api/internal/domain"
	"xyz-multifinance-api/internal/model"
	"xyz-multifinance-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	useCase *usecase.TransactionUseCase
}

func NewTransactionHandler(router *gin.Engine, transactionUseCase *usecase.TransactionUseCase) {
	handler := &TransactionHandler{useCase: transactionUseCase}

	v1 := router.Group("/api/v1")
	{
		v1.POST("/transactions", handler.CreateTransaction)
		v1.GET("/transactions/contract/:contract_number", handler.GetTransactionByContractNumber)
		v1.GET("/customers/:customer_id/transactions", handler.GetTransactionsByCustomerID)
	}
}

func (h *TransactionHandler) CreateTransaction(ctx *gin.Context) {
	req := new(model.CreateTransactionRequest)
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	transactionRes, err := h.useCase.CreateTransaction(req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInput):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input provided", "details": err.Error()})
		case errors.Is(err, domain.ErrNotFound): // Customer or credit limit does not exist
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrInsufficientCredit): // Credit limit reached
			ctx.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()}) // 402 Payment Required
		case errors.Is(err, domain.ErrAlreadyExists): // Contract number already exist
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusCreated, transactionRes)
}

func (h *TransactionHandler) GetTransactionByContractNumber(ctx *gin.Context) {
	contractNumber := ctx.Param("contract_number")
	if contractNumber == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "contract number is required"})
		return
	}

	transactionRes, err := h.useCase.GetTransactionByContractNumber(contractNumber)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		} else {
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, transactionRes)
}

func (h *TransactionHandler) GetTransactionsByCustomerID(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	if customerID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "customer ID is required"})
		return
	}

	transactionsRes, err := h.useCase.GetTransactionsByCustomerID(customerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) { // Customer does not exist
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			ctx.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	if len(transactionsRes) == 0 {
		ctx.JSON(http.StatusOK, []interface{}{}) // No transactions found
		return
	}
	ctx.JSON(http.StatusOK, transactionsRes)
}
