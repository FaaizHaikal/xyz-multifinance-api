package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"xyz-multifinance-api/config"
	"xyz-multifinance-api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		tokenString := parts[1]

		claims := &model.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired or not active yet"})
				return
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			return
		}

		if !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		ctx.Set("customerID", claims.CustomerID)
		ctx.Set("customerNIK", claims.NIK)

		ctx.Next()
	}
}

func GetCustomerIDFromContext(ctx *gin.Context) (string, bool) {
	customerID, exists := ctx.Get("customerID")
	if !exists {
		return "", false
	}
	return customerID.(string), true
}

func GetCustomerNIKFromContext(ctx *gin.Context) (string, bool) {
	customerNIK, exists := ctx.Get("customerNIK")
	if !exists {
		return "", false
	}
	return customerNIK.(string), true
}
