package middlewares

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RequestLogger logs HTTP requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Msg("incoming request")

		c.Next()

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Error().
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Int("status", c.Writer.Status()).
					Err(e.Err).
					Msg("request error")
			}
		}

		logEvent := log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status())

		if c.Writer.Status() >= 400 {
			logEvent = log.Warn().
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int("status", c.Writer.Status())
		}

		logEvent.Msg("request completed")
	}
}

// CORS adds CORS headers
func CORS(allowedOrigins []string) gin.HandlerFunc {
	normalized := map[string]struct{}{}
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		normalized[trimmed] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Add("Vary", "Origin")
			if len(normalized) == 0 {
				// No origins configured; allow only same-origin requests.
				if sameOrigin(origin, c.Request.Host) {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key, Idempotency-Key, X-Request-Id, Mcp-Session-Id, mcp-protocol-version")
					c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")
					c.Writer.Header().Set("Access-Control-Max-Age", "3600")
				} else {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
			if _, ok := normalized[origin]; !ok && len(normalized) > 0 {
				// Reject disallowed cross-origin requests
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		if origin != "" {
			if _, ok := normalized[origin]; ok && len(normalized) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Key, Idempotency-Key, X-Request-Id, Mcp-Session-Id, mcp-protocol-version")
				c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")
				c.Writer.Header().Set("Access-Control-Max-Age", "3600")
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func sameOrigin(origin string, requestHost string) bool {
	if origin == "" || requestHost == "" {
		return false
	}
	o := origin
	if !strings.Contains(origin, "://") {
		o = "https://" + origin
	}
	parsed, err := url.Parse(o)
	if err != nil {
		return false
	}
	return strings.EqualFold(parsed.Host, requestHost)
}
