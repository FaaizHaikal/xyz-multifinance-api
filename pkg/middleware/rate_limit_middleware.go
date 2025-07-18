package middleware

import (
	"fmt"
	"net/http"
	"time"

	internalredis "xyz-multifinance-api/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiterConfig struct {
	RequestsPerSecond int
	Burst             int
	Window            time.Duration
}

func RateLimitMiddleware(config RateLimiterConfig, rdbClient *redis.Client) gin.HandlerFunc { // NEW: rdbClient argument
	if config.Window == 0 {
		config.Window = time.Second // Default window to 1 second
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown" // Fallback
		}

		// Use a fixed-window key based on current time window
		windowStart := time.Now().Truncate(config.Window).UnixNano() / int64(config.Window)
		key := fmt.Sprintf("rate_limit:%s:%d", ip, windowStart)

		pipe := rdbClient.Pipeline() // Use injected rdbClient
		incr := pipe.Incr(internalredis.Ctx, key)
		pipe.Expire(internalredis.Ctx, key, config.Window) // Set/reset expiration for the window
		_, err := pipe.Exec(internalredis.Ctx)

		if err != nil {
			// Log the error but don't expose Redis details to client
			c.Error(fmt.Errorf("redis pipeline error for rate limit: %w", err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limit service unavailable"})
			return
		}

		currentRequests := incr.Val()

		// Calculate Retry-After based on remaining time in the current window
		ttl := rdbClient.TTL(internalredis.Ctx, key).Val() // Use injected rdbClient
		retryAfterSeconds := int(ttl.Seconds())
		if retryAfterSeconds <= 0 { // If TTL is already expired or negative, set to 1s
			retryAfterSeconds = 1
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerSecond))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, config.RequestsPerSecond-int(currentRequests))))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix())) // Unix timestamp of reset

		if currentRequests > int64(config.RequestsPerSecond) {
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": fmt.Sprintf("You have exceeded your request rate limit. Please try again after %d seconds.", retryAfterSeconds),
			})
			return
		}

		c.Next() // Continue to the next handler
	}
}
