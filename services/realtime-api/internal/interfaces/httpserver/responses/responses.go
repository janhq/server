// Package responses contains HTTP response DTOs for the realtime-api.
package responses

// SessionResponse represents a session in API responses.
type SessionResponse struct {
	ID           string              `json:"id"`
	Object       string              `json:"object"`
	ClientSecret *ClientSecretDetail `json:"client_secret,omitempty"`
	WsURL        string              `json:"ws_url,omitempty"`
	RoomID       string              `json:"room_id,omitempty"`
	UserID       string              `json:"user_id,omitempty"`
	Status       string              `json:"status,omitempty"`
}

// ClientSecretDetail contains the client secret for a session.
type ClientSecretDetail struct {
	Value     string `json:"value"`
	ExpiresAt int64  `json:"expires_at"`
}

// ListSessionsResponse represents the response for listing sessions.
type ListSessionsResponse struct {
	Object string             `json:"object"`
	Data   []*SessionResponse `json:"data"`
}

// DeleteSessionResponse represents the response for deleting a session.
type DeleteSessionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error *ErrorDetail `json:"error"`
}

// ErrorDetail contains error details.
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}
