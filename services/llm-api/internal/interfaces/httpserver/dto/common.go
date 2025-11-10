// Package dto provides data transfer objects for HTTP requests/responses
package dto

// Pagination holds pagination parameters
type Pagination struct {
	Page     int   `json:"page" form:"page"`
	PageSize int   `json:"page_size" form:"page_size"`
	Total    int64 `json:"total"`
}

// Response is a generic API response wrapper
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo holds error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
