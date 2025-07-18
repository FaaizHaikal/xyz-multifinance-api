package main

import (
	"fmt"
	"log"
	"net/http"
	"xyz-multifinance-api/config"
	"xyz-multifinance-api/internal/infrastructure/database"
	internalredis "xyz-multifinance-api/internal/infrastructure/redis"
	"xyz-multifinance-api/internal/repository"
	"xyz-multifinance-api/internal/usecase"
	"xyz-multifinance-api/pkg/middleware"

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

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB from GORM: %v", err)
	}
	defer sqlDB.Close()

	redisClient, err := internalredis.InitRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	cacheStore := internalredis.NewRedisCacheStore(redisClient) // NEW: Create CacheStore instance
	defer cacheStore.Close()

	router := gin.Default()

	customerRepo := repository.NewCustomerRepository(gormDB, cacheStore)
	creditLimitRepo := repository.NewCreditLimitRepository(gormDB, cacheStore)
	transactionRepo := repository.NewTransactionRepository(gormDB)

	authUseCase := usecase.NewAuthUseCase(customerRepo, cfg)
	customerUseCase := usecase.NewCustomerUseCase(customerRepo)
	creditLimitUseCase := usecase.NewCreditLimitUseCase(creditLimitRepo, customerRepo)
	transactionUseCase := usecase.NewTransactionUseCase(gormDB, transactionRepo, customerRepo, creditLimitRepo, cacheStore)

	apphttp.NewAuthHandler(router, authUseCase)

	protectedV1 := router.Group("/api/v1")
	protectedV1.Use(
		middleware.JWTAuthMiddleware(cfg),
		middleware.RateLimitMiddleware(middleware.RateLimiterConfig{
			RequestsPerSecond: cfg.RateLimitPerSecond,
			Burst:             cfg.RateLimitBurst,
		}, redisClient),
	)
	{
		apphttp.NewCustomerHandler(protectedV1, customerUseCase)
		apphttp.NewCreditLimitHandler(protectedV1, creditLimitUseCase)
		apphttp.NewTransactionHandler(protectedV1, transactionUseCase)
	}

	serverAddress := fmt.Sprintf(":%s", cfg.APIPort)
	log.Printf("Starting server on %s...", serverAddress)
	if err := router.Run(serverAddress); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
