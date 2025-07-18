package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser             string
	DBPassword         string
	DBHost             string
	DBPort             string
	DBName             string
	APIPort            string
	RedisAddr          string
	JWTSecret          string
	JWTRefreshSecret   string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	RateLimitPerSecond int
	RateLimitBurst     int
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	accessTokenExpiryStr := getEnv("ACCESS_TOKEN_EXPIRY_MINUTES", "15")
	accessTokenExpiryMinutes, err := strconv.Atoi(accessTokenExpiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_EXPIRY_MINUTES: %w", err)
	}

	refreshTokenExpiryStr := getEnv("REFRESH_TOKEN_EXPIRY_DAYS", "7")
	refreshTokenExpiryDays, err := strconv.Atoi(refreshTokenExpiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_EXPIRY_DAYS: %w", err)
	}

	rateLimitPerSecondStr := getEnv("RATE_LIMIT_PER_SECOND", "10")
	rateLimitPerSecond, err := strconv.Atoi(rateLimitPerSecondStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_PER_SECOND: %w", err)
	}

	rateLimitBurstStr := getEnv("RATE_LIMIT_BURST", "20")
	rateLimitBurst, err := strconv.Atoi(rateLimitBurstStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_BURST: %w", err)
	}

	cfg := &Config{
		DBUser:             getEnv("DB_USER", "root"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBName:             getEnv("DB_NAME", "xyz_multifinance"),
		APIPort:            getEnv("API_PORT", "8080"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", "qwoaoiscmoaqoiwdmomocmosmc"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", "owqopkdfmvzxmcdvcpqpwo"),
		AccessTokenExpiry:  time.Duration(accessTokenExpiryMinutes) * time.Minute,
		RefreshTokenExpiry: time.Duration(refreshTokenExpiryDays) * 24 * time.Hour,
		RateLimitPerSecond: rateLimitPerSecond,
		RateLimitBurst:     rateLimitBurst,
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.APIPort == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}

	return defaultValue
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
