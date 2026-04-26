package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/internal/delivery/dto"
	"github.com/giftsense/backend/internal/port"
)

// CORS sets Access-Control-Allow-* headers and handles preflight OPTIONS requests.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RateLimit checks per-IP request rate using the provided RateLimiter.
// Fails open: if the rate limiter returns an error, the request is allowed through.
func RateLimit(limiter port.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, err := limiter.Allow(c.Request.Context(), c.ClientIP())
		if err != nil {
			log.Printf("rate limiter error: %v", err)
			c.Next()
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error:   "rate_limited",
				Message: "Too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}

// RequestSizeLimiter aborts with 413 if Content-Length exceeds maxBytes.
func RequestSizeLimiter(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, dto.ErrorResponse{
				Error:   "file_too_large",
				Message: "Request body exceeds maximum allowed size",
			})
			return
		}
		c.Next()
	}
}
