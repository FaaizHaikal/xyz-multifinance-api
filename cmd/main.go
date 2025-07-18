package main

import (
	"fmt"
	"log"
	"net/http"
	"xyz-multifinance-api/config"
	"xyz-multifinance-api/internal/infrastructure/database"
	"xyz-multifinance-api/internal/repository"
	"xyz-multifinance-api/internal/usecase"

	"github.com/gin-gonic/gin"

	apphttp "xyz-multifinance-api/internal/delivery/http"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gormDB, err := database.InitGORM(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	router := gin.Default()

	customerRepo := repository.NewCustomerRepository(gormDB)
	creditLimitRepo := repository.NewCreditLimitRepository(gormDB)
	transactionRepo := repository.NewTransactionRepository(gormDB)

	customerUseCase := usecase.NewCustomerUseCase(customerRepo)
	creditLimitUseCase := usecase.NewCreditLimitUseCase(creditLimitRepo, customerRepo)
	transactionUseCase := usecase.NewTransactionUseCase(gormDB, transactionRepo, customerRepo, creditLimitRepo)

	apphttp.NewCustomerHandler(router, customerUseCase)
	apphttp.NewCreditLimitHandler(router, creditLimitUseCase)
	apphttp.NewTransactionHandler(router, transactionUseCase)

	serverAddress := fmt.Sprintf(":%s", cfg.APIPort)
	log.Printf("Starting server on %s...", serverAddress)
	if err := router.Run(serverAddress); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
