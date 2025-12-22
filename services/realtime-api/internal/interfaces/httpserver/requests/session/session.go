// Package session contains HTTP request DTOs for session endpoints.
package session

// CreateSessionRequest represents the request body for creating a session.
// Currently minimal as session creation doesn't require many parameters.
type CreateSessionRequest struct {
	// Model is an optional model identifier for future use.
	Model string `json:"model,omitempty"`
	// Voice is an optional voice identifier for future use.
	Voice string `json:"voice,omitempty"`
}
