package response

import (
	"errors"
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

// sanitizeError sanitizes an error message for safe logging
func (h *ErrorHandler) sanitizeError(err error) string {
	if err == nil {
		return ""
	}

	// Get the error message and sanitize it
	errMsg := err.Error()
	return h.sanitizer.SingleLine(errMsg)
}

// sanitizePath sanitizes a URL path for safe logging
func (h *ErrorHandler) sanitizePath(path string) string {
	if path == "" {
		return ""
	}

	// Use the sanitization service to clean the path
	return h.sanitizer.SingleLine(path)
}

// sanitizeRequestID sanitizes a request ID for safe logging
func (h *ErrorHandler) sanitizeRequestID(requestID string) string {
	if requestID == "" {
		return ""
	}

	// Use the sanitization service to clean the request ID
	return h.sanitizer.SingleLine(requestID)
}

// HandleError handles errors consistently across the application
func (h *ErrorHandler) HandleError(err error, c echo.Context, message string) error {
	requestID := h.sanitizeRequestID(c.Request().Header.Get("X-Trace-Id"))
	userID, ok := c.Get("user_id").(string)
	if !ok {
		userID = ""
	}
	if h.logger != nil {
		h.logger.Error("request error",
			"error", h.sanitizeError(err),
			"path", h.sanitizePath(c.Request().URL.Path),
			"method", c.Request().Method,
			"request_id", requestID,
			"user_id", userID,
		)
	}

	// Check if it's a domain error
	var domainErr *domainerrors.DomainError
	if errors.As(err, &domainErr) {
		return h.handleDomainError(domainErr, c)
	}

	// Handle unknown errors
	return h.handleUnknownError(err, c, message)
}

// HandleDomainError handles domain-specific errors
func (h *ErrorHandler) HandleDomainError(err *domainerrors.DomainError, c echo.Context) error {
	statusCode := h.getStatusCode(err.Code)
	requestID := h.sanitizeRequestID(c.Request().Header.Get("X-Trace-Id"))
	userID, ok := c.Get("user_id").(string)
	if !ok {
		userID = ""
	}

	// Check if this is an AJAX request
	if h.isAJAXRequest(c) {
		return c.JSON(statusCode, map[string]any{
			"error":      string(err.Code),
			"message":    err.Message,
			"details":    err.Context,
			"request_id": requestID,
			"user_id":    userID,
		})
	}

	// For regular requests, redirect with error message
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/error?code=%s&message=%s", err.Code, err.Message))
}

// HandleValidationError handles validation errors specifically
func (h *ErrorHandler) HandleValidationError(err error, c echo.Context, field string) error {
	validationErr := domainerrors.New(domainerrors.ErrCodeValidation, err.Error(), nil).
		WithContext("field", field)

	return h.HandleDomainError(validationErr, c)
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

// handleDomainError is the internal method for handling domain errors
func (h *ErrorHandler) handleDomainError(err *domainerrors.DomainError, c echo.Context) error {
	return h.HandleDomainError(err, c)
}

// handleUnknownError handles unknown errors
func (h *ErrorHandler) handleUnknownError(_ error, c echo.Context, message string) error {
	statusCode := http.StatusInternalServerError
	requestID := h.sanitizeRequestID(c.Request().Header.Get("X-Trace-Id"))
	userID, ok := c.Get("user_id").(string)
	if !ok {
		userID = ""
	}
	if h.isAJAXRequest(c) {
		return c.JSON(statusCode, map[string]any{
			"error":      "INTERNAL_ERROR",
			"message":    message,
			"request_id": requestID,
			"user_id":    userID,
		})
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/error?message=%s", message))
}

// getStatusCode maps error codes to HTTP status codes
func (h *ErrorHandler) getStatusCode(code domainerrors.ErrorCode) int {
	switch code {
	case domainerrors.ErrCodeValidation, domainerrors.ErrCodeRequired, domainerrors.ErrCodeInvalid,
		domainerrors.ErrCodeInvalidFormat, domainerrors.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case domainerrors.ErrCodeUnauthorized, domainerrors.ErrCodeAuthentication:
		return http.StatusUnauthorized
	case domainerrors.ErrCodeForbidden, domainerrors.ErrCodeInsufficientRole:
		return http.StatusForbidden
	case domainerrors.ErrCodeNotFound:
		return http.StatusNotFound
	case domainerrors.ErrCodeConflict, domainerrors.ErrCodeAlreadyExists:
		return http.StatusConflict
	case domainerrors.ErrCodeBadRequest:
		return http.StatusBadRequest
	case domainerrors.ErrCodeServerError:
		return http.StatusInternalServerError
	case domainerrors.ErrCodeStartup, domainerrors.ErrCodeShutdown:
		return http.StatusServiceUnavailable
	case domainerrors.ErrCodeConfig, domainerrors.ErrCodeDatabase:
		return http.StatusInternalServerError
	case domainerrors.ErrCodeTimeout:
		return http.StatusGatewayTimeout
	case domainerrors.ErrCodeFormValidation, domainerrors.ErrCodeFormInvalid, domainerrors.ErrCodeFormExpired:
		return http.StatusBadRequest
	case domainerrors.ErrCodeFormNotFound:
		return http.StatusNotFound
	case domainerrors.ErrCodeFormSubmission, domainerrors.ErrCodeFormAccessDenied:
		return http.StatusBadRequest
	case domainerrors.ErrCodeUserNotFound:
		return http.StatusNotFound
	case domainerrors.ErrCodeUserExists:
		return http.StatusConflict
	case domainerrors.ErrCodeUserDisabled, domainerrors.ErrCodeUserInvalid, domainerrors.ErrCodeUserUnauthorized:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
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
	HandleValidationError(err error, c echo.Context, field string) error
	HandleAuthError(err error, c echo.Context) error
	HandleNotFoundError(resource string, c echo.Context) error
}

// Ensure ErrorHandler implements ErrorHandlerInterface
var _ ErrorHandlerInterface = (*ErrorHandler)(nil)
