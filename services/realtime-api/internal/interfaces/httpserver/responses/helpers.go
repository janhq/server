package responses

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"jan-server/services/realtime-api/internal/infrastructure/store"
	"jan-server/services/realtime-api/internal/utils/platformerrors"
)

// HandleError handles errors and writes appropriate HTTP responses.
// It maps store-specific errors and platform errors to HTTP status codes.
func HandleError(c *gin.Context, err error, message string) {
	logger := log.With().Str("path", c.Request.URL.Path).Logger()

	// Check for store-specific errors first
	if errors.Is(err, store.ErrSessionNotFound) {
		platformerrors.WriteNotFound(c, message)
		return
	}
	if errors.Is(err, store.ErrSessionAlreadyExists) || errors.Is(err, store.ErrRoomAlreadyExists) {
		platformerrors.WriteConflict(c, message)
		return
	}

	// Use platform error handler for everything else
	platformerrors.WriteError(c, err, logger)
}

// HandleErrorWithStatus handles errors with a custom HTTP status code.
func HandleErrorWithStatus(c *gin.Context, statusCode int, err error, message string) {
	c.JSON(statusCode, platformerrors.HTTPErrorResponse{
		Error: &platformerrors.HTTPErrorDetail{
			Message: message,
			Type:    statusToErrorType(statusCode),
		},
	})
}

// HandleNewError creates and writes a new typed error response.
// Use this for route-level errors like validation or authorization failures.
func HandleNewError(c *gin.Context, errorType platformerrors.ErrorType, message string) {
	status := platformerrors.ErrorTypeToHTTPStatus(errorType)
	c.JSON(status, platformerrors.HTTPErrorResponse{
		Error: &platformerrors.HTTPErrorDetail{
			Message: message,
			Type:    errorTypeToString(errorType),
		},
	})
}

// statusToErrorType converts HTTP status code to error type string.
func statusToErrorType(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "validation_error"
	case http.StatusUnauthorized:
		return "unauthorized_error"
	case http.StatusForbidden:
		return "forbidden_error"
	case http.StatusNotFound:
		return "not_found_error"
	case http.StatusConflict:
		return "conflict_error"
	case http.StatusTooManyRequests:
		return "rate_limited_error"
	default:
		return "internal_error"
	}
}

// errorTypeToString converts an ErrorType to a snake_case string for API responses.
func errorTypeToString(t platformerrors.ErrorType) string {
	switch t {
	case platformerrors.ErrorTypeNotFound:
		return "not_found_error"
	case platformerrors.ErrorTypeValidation:
		return "validation_error"
	case platformerrors.ErrorTypeConflict:
		return "conflict_error"
	case platformerrors.ErrorTypeUnauthorized:
		return "unauthorized_error"
	case platformerrors.ErrorTypeForbidden:
		return "forbidden_error"
	case platformerrors.ErrorTypeNotImplemented:
		return "not_implemented_error"
	case platformerrors.ErrorTypeExpired:
		return "expired_error"
	case platformerrors.ErrorTypeRateLimited:
		return "rate_limited_error"
	case platformerrors.ErrorTypeTimeout:
		return "timeout_error"
	case platformerrors.ErrorTypeExternal:
		return "external_error"
	case platformerrors.ErrorTypeTooManyRecords:
		return "too_many_records_error"
	case platformerrors.ErrorTypeInternal:
		fallthrough
	default:
		return "internal_error"
	}
}
