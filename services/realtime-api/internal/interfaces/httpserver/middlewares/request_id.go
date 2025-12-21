package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header key for request ID.
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID.
	RequestIDKey = "request_id"
)

// RequestID middleware generates or propagates a unique request ID.
// If the incoming request has an X-Request-ID header, it uses that value.
// Otherwise, it generates a new UUID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set in context for downstream use
		c.Set(RequestIDKey, requestID)

		// Set in response header
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}
