package responses

import (
	"encoding/json"
	"net/http"

	"github.com/janhq/jan-server/services/memory-tools/internal/interfaces/httpserver/middleware"
	"github.com/rs/zerolog/log"
)

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// JSON writes a JSON response and propagates the request ID header.
func JSON(w http.ResponseWriter, r *http.Request, status int, payload interface{}) {
	requestID := middleware.GetRequestID(r.Context())
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Ctx(r.Context()).Error().Err(err).Msg("failed to write response")
	}
}

// Error writes a structured error response with request ID included.
func Error(w http.ResponseWriter, r *http.Request, status int, message string) {
	resp := errorResponse{
		Error:     message,
		RequestID: middleware.GetRequestID(r.Context()),
	}
	JSON(w, r, status, resp)
}
