package platformerrors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// HTTPErrorResponse represents the standard error response format.
type HTTPErrorResponse struct {
	Error *HTTPErrorDetail `json:"error"`
}

// HTTPErrorDetail contains error details for HTTP responses.
type HTTPErrorDetail struct {
	Message   string `json:"message"`
	Type      string `json:"type"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteHTTPError writes a PlatformError as an HTTP response.
// It maps the error type to an appropriate HTTP status code and formats the response.
func WriteHTTPError(c *gin.Context, err *PlatformError, log zerolog.Logger) {
	if err == nil {
		c.JSON(http.StatusInternalServerError, HTTPErrorResponse{
			Error: &HTTPErrorDetail{
				Message: "unknown error",
				Type:    "internal_error",
			},
		})
		return
	}

	// Log the error
	LogError(log, err)

	status := ErrorTypeToHTTPStatus(err.Type)
	response := HTTPErrorResponse{
		Error: &HTTPErrorDetail{
			Message:   err.Message,
			Type:      errorTypeToString(err.Type),
			Code:      err.UUID,
			RequestID: err.RequestID,
		},
	}

	c.JSON(status, response)
}

// WriteError writes a generic error as an HTTP response.
// If the error is a PlatformError, it will be handled appropriately.
// Otherwise, it will be treated as an internal error.
func WriteError(c *gin.Context, err error, log zerolog.Logger) {
	if err == nil {
		c.JSON(http.StatusInternalServerError, HTTPErrorResponse{
			Error: &HTTPErrorDetail{
				Message: "unknown error",
				Type:    "internal_error",
			},
		})
		return
	}

	platformErr := GetPlatformError(err)
	if platformErr != nil {
		WriteHTTPError(c, platformErr, log)
		return
	}

	// Generic error - treat as internal
	c.JSON(http.StatusInternalServerError, HTTPErrorResponse{
		Error: &HTTPErrorDetail{
			Message: err.Error(),
			Type:    "internal_error",
		},
	})
}

// WriteNotFound writes a 404 Not Found response.
func WriteNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, HTTPErrorResponse{
		Error: &HTTPErrorDetail{
			Message: message,
			Type:    "not_found_error",
		},
	})
}

// WriteValidationError writes a 400 Bad Request response.
func WriteValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, HTTPErrorResponse{
		Error: &HTTPErrorDetail{
			Message: message,
			Type:    "validation_error",
		},
	})
}

// WriteUnauthorized writes a 401 Unauthorized response.
func WriteUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, HTTPErrorResponse{
		Error: &HTTPErrorDetail{
			Message: message,
			Type:    "unauthorized_error",
		},
	})
}

// WriteForbidden writes a 403 Forbidden response.
func WriteForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, HTTPErrorResponse{
		Error: &HTTPErrorDetail{
			Message: message,
			Type:    "forbidden_error",
		},
	})
}

// WriteConflict writes a 409 Conflict response.
func WriteConflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, HTTPErrorResponse{
		Error: &HTTPErrorDetail{
			Message: message,
			Type:    "conflict_error",
		},
	})
}

// WriteInternalError writes a 500 Internal Server Error response.
func WriteInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, HTTPErrorResponse{
		Error: &HTTPErrorDetail{
			Message: message,
			Type:    "internal_error",
		},
	})
}

// errorTypeToString converts an ErrorType to a snake_case string for API responses.
func errorTypeToString(t ErrorType) string {
	switch t {
	case ErrorTypeNotFound:
		return "not_found_error"
	case ErrorTypeValidation:
		return "validation_error"
	case ErrorTypeConflict:
		return "conflict_error"
	case ErrorTypeUnauthorized:
		return "unauthorized_error"
	case ErrorTypeForbidden:
		return "forbidden_error"
	case ErrorTypeNotImplemented:
		return "not_implemented_error"
	case ErrorTypeExpired:
		return "expired_error"
	case ErrorTypeRateLimited:
		return "rate_limited_error"
	case ErrorTypeTimeout:
		return "timeout_error"
	case ErrorTypeExternal:
		return "external_error"
	case ErrorTypeTooManyRecords:
		return "too_many_records_error"
	case ErrorTypeInternal:
		fallthrough
	default:
		return "internal_error"
	}
}
