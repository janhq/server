package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type requestIDKey struct{}

// RequestIDMiddleware ensures every request has an ID and logger bound to the context.
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.NewString()
			}

			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
			logger := log.With().Str("request_id", requestID).Logger()
			ctx = logger.WithContext(ctx)

			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestID extracts the request ID from context.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}

	return ""
}
