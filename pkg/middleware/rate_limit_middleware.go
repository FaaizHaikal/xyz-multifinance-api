package middleware

import (
	"fmt"
	"net/http"
	"time"
	"xyz-multifinance-api/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
)

type RateLimiterConfig struct {
	RequestsPerSecond int
	Burst             int
	Window            time.Duration
}

func RateLimitMiddleware(config RateLimiterConfig) gin.HandlerFunc {
	if config.Window == 0 {
		config.Window = time.Second // Default window to 1 second
	}

	return func(ctx *gin.Context) {
		ip := ctx.ClientIP()
		if ip == "" {
			ip = "unknown" // Fallback
		}

		key := fmt.Sprintf("rate_limit:%s:%d", ip, time.Now().UnixNano()/int64(config.Window))

		pipe := redis.RDB.Pipeline()
		incr := pipe.Incr(redis.Ctx, key)
		pipe.Expire(redis.Ctx, key, config.Window)
		_, err := pipe.Exec(redis.Ctx)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limit service unavailable"})
			ctx.Error(fmt.Errorf("redis pipeline error for rate limit: %w", err))
			return
		}

		currentRequests := incr.Val()

		if currentRequests > int64(config.RequestsPerSecond) {
			ttl := redis.RDB.TTL(redis.Ctx, key).Val()
			retryAfterSeconds := int(ttl.Seconds())
			if retryAfterSeconds <= 0 { // If TTL is already expired or negative, set to 1s
				retryAfterSeconds = 1
			}

			ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerSecond))
			ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", 0)) // No requests remaining
			ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))
			ctx.Header("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))

			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests",
				"message": fmt.Sprintf("You have exceeded your request rate limit. Please try again after %d seconds.", retryAfterSeconds),
			})
			return
		}

		// If within limit, set remaining count (approximate)
		ctx.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerSecond))
		ctx.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", config.RequestsPerSecond-int(currentRequests)))
		ttl := redis.RDB.TTL(redis.Ctx, key).Val()
		ctx.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

		ctx.Next()
	}
}
