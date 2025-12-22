// Package responses contains HTTP response DTOs for the realtime-api.
// Session-specific response types are in the session subpackage.
package responses

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
