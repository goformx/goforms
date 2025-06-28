// Package response provides HTTP response handling utilities including
// error handling, response building, and standardized response formats.
package response

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// ErrorHandler provides unified error handling across the application
type ErrorHandler struct {
	logger    logging.Logger
	sanitizer sanitization.ServiceInterface
}

// NewErrorHandler creates a new error handler instance
func NewErrorHandler(logger logging.Logger, sanitizer sanitization.ServiceInterface) *ErrorHandler {
	return &ErrorHandler{
		logger:    logger,
		sanitizer: sanitizer,
	}
}

// HandleError handles generic errors
func (h *ErrorHandler) HandleError(_ error, c echo.Context, message string) error {
	statusCode := http.StatusInternalServerError

	// Check if this is an AJAX request
	if h.isAJAXRequest(c) {
		return ErrorResponse(c, statusCode, message)
	}

	// For web requests, return a simple error response
	return ErrorResponse(c, statusCode, message)
}

// HandleDomainError handles domain-specific errors
func (h *ErrorHandler) HandleDomainError(err *domainerrors.DomainError, c echo.Context) error {
	statusCode := h.getStatusCode(err.Code)

	// Check if this is an AJAX request
	if h.isAJAXRequest(c) {
		return ErrorResponse(c, statusCode, err.Message)
	}

	// For web requests, return a simple error response
	return ErrorResponse(c, statusCode, err.Message)
}

// HandleAuthError handles authentication errors
func (h *ErrorHandler) HandleAuthError(err error, c echo.Context) error {
	authErr := domainerrors.New(domainerrors.ErrCodeUnauthorized, "Authentication required", err)

	return h.HandleDomainError(authErr, c)
}

// HandleNotFoundError handles not found errors
func (h *ErrorHandler) HandleNotFoundError(resource string, c echo.Context) error {
	notFoundErr := domainerrors.New(domainerrors.ErrCodeNotFound, fmt.Sprintf("%s not found", resource), nil)

	return h.HandleDomainError(notFoundErr, c)
}

// getStatusCode maps error codes to HTTP status codes
func (h *ErrorHandler) getStatusCode(code domainerrors.ErrorCode) int {
	// Map of error codes to HTTP status codes
	statusCodeMap := map[domainerrors.ErrorCode]int{
		// Validation errors
		domainerrors.ErrCodeValidation:    http.StatusBadRequest,
		domainerrors.ErrCodeRequired:      http.StatusBadRequest,
		domainerrors.ErrCodeInvalid:       http.StatusBadRequest,
		domainerrors.ErrCodeInvalidFormat: http.StatusBadRequest,
		domainerrors.ErrCodeInvalidInput:  http.StatusBadRequest,
		domainerrors.ErrCodeBadRequest:    http.StatusBadRequest,

		// Authentication errors
		domainerrors.ErrCodeUnauthorized:   http.StatusUnauthorized,
		domainerrors.ErrCodeAuthentication: http.StatusUnauthorized,

		// Authorization errors
		domainerrors.ErrCodeForbidden:        http.StatusForbidden,
		domainerrors.ErrCodeInsufficientRole: http.StatusForbidden,

		// Resource errors
		domainerrors.ErrCodeNotFound:     http.StatusNotFound,
		domainerrors.ErrCodeFormNotFound: http.StatusNotFound,
		domainerrors.ErrCodeUserNotFound: http.StatusNotFound,

		// Conflict errors
		domainerrors.ErrCodeConflict:      http.StatusConflict,
		domainerrors.ErrCodeAlreadyExists: http.StatusConflict,
		domainerrors.ErrCodeUserExists:    http.StatusConflict,

		// Server errors
		domainerrors.ErrCodeServerError: http.StatusInternalServerError,
		domainerrors.ErrCodeConfig:      http.StatusInternalServerError,
		domainerrors.ErrCodeDatabase:    http.StatusInternalServerError,

		// Service errors
		domainerrors.ErrCodeStartup:  http.StatusServiceUnavailable,
		domainerrors.ErrCodeShutdown: http.StatusServiceUnavailable,
		domainerrors.ErrCodeTimeout:  http.StatusGatewayTimeout,

		// Form errors
		domainerrors.ErrCodeFormValidation:   http.StatusBadRequest,
		domainerrors.ErrCodeFormInvalid:      http.StatusBadRequest,
		domainerrors.ErrCodeFormExpired:      http.StatusBadRequest,
		domainerrors.ErrCodeFormSubmission:   http.StatusBadRequest,
		domainerrors.ErrCodeFormAccessDenied: http.StatusBadRequest,

		// User errors
		domainerrors.ErrCodeUserDisabled:     http.StatusBadRequest,
		domainerrors.ErrCodeUserInvalid:      http.StatusBadRequest,
		domainerrors.ErrCodeUserUnauthorized: http.StatusBadRequest,
	}

	if statusCode, exists := statusCodeMap[code]; exists {
		return statusCode
	}

	return http.StatusInternalServerError
}

// isAJAXRequest checks if the request is an AJAX request
func (h *ErrorHandler) isAJAXRequest(c echo.Context) bool {
	return c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		c.Request().Header.Get("Content-Type") == "application/json"
}

// ErrorHandlerInterface defines the interface for error handling
type ErrorHandlerInterface interface {
	HandleError(err error, c echo.Context, message string) error
	HandleDomainError(err *domainerrors.DomainError, c echo.Context) error
	HandleAuthError(err error, c echo.Context) error
	HandleNotFoundError(resource string, c echo.Context) error
}

// Ensure ErrorHandler implements ErrorHandlerInterface
var _ ErrorHandlerInterface = (*ErrorHandler)(nil)
