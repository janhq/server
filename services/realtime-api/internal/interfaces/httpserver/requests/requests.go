// Package requests contains HTTP request DTOs for the realtime-api.
package requests

// CreateSessionRequest represents the request body for creating a session.
// Currently empty as session creation doesn't require parameters.
type CreateSessionRequest struct {
	// Model is an optional model identifier for future use.
	Model string `json:"model,omitempty"`
	// Voice is an optional voice identifier for future use.
	Voice string `json:"voice,omitempty"`
}
