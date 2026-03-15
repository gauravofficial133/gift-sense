package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/giftsense/backend/internal/delivery/dto"
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
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
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
