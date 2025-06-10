package middleware

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// ErrorHandler handles errors and returns appropriate HTTP responses
type ErrorHandler struct {
	logger logging.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger logging.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// Handle handles errors and returns appropriate HTTP responses
func (h *ErrorHandler) Handle(err error, c echo.Context) {
	if err == nil {
		return
	}

	// Handle domain errors
	var domainErr *errors.DomainError
	if stderrors.As(err, &domainErr) {
		h.logger.Error("error handling request", "error", err, "path", c.Request().URL.Path)

		statusCode := http.StatusInternalServerError
		switch domainErr.Code {
		case errors.ErrCodeValidation:
			statusCode = http.StatusBadRequest
		case errors.ErrCodeNotFound:
			statusCode = http.StatusNotFound
		case errors.ErrCodeUnauthorized:
			statusCode = http.StatusUnauthorized
		case errors.ErrCodeForbidden:
			statusCode = http.StatusForbidden
		case errors.ErrCodeRequired, errors.ErrCodeInvalid, errors.ErrCodeInvalidFormat, errors.ErrCodeInvalidInput:
			statusCode = http.StatusBadRequest
		case errors.ErrCodeInvalidToken, errors.ErrCodeAuthentication:
			statusCode = http.StatusUnauthorized
		case errors.ErrCodeInsufficientRole:
			statusCode = http.StatusForbidden
		case errors.ErrCodeConflict, errors.ErrCodeAlreadyExists:
			statusCode = http.StatusConflict
		case errors.ErrCodeBadRequest:
			statusCode = http.StatusBadRequest
		case errors.ErrCodeServerError, errors.ErrCodeDatabase, errors.ErrCodeTimeout:
			statusCode = http.StatusInternalServerError
		case errors.ErrCodeStartup, errors.ErrCodeShutdown, errors.ErrCodeConfig:
			statusCode = http.StatusServiceUnavailable
		}

		response := map[string]any{
			"error": map[string]any{
				"code":    domainErr.Code,
				"message": domainErr.Message,
			},
		}

		if jsonErr := c.JSON(statusCode, response); jsonErr != nil {
			h.logger.Error("error handling request", "error", jsonErr)
		}

		return
	}

	// Handle HTTP errors
	var httpErr *echo.HTTPError
	if stderrors.As(err, &httpErr) {
		h.logger.Error("error handling request", "error", err)

		response := map[string]any{
			"error": map[string]any{
				"code":    "HTTP_ERROR",
				"message": httpErr.Message,
			},
		}

		if jsonErr := c.JSON(httpErr.Code, response); jsonErr != nil {
			h.logger.Error("error handling request", "error", jsonErr)
		}

		return
	}

	// Handle unknown errors
	h.logger.Error("error handling request", "error", err)

	response := map[string]any{
		"error": map[string]any{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	}

	if jsonErr := c.JSON(http.StatusInternalServerError, response); jsonErr != nil {
		h.logger.Error("error handling request", "error", jsonErr)
	}
}

// Middleware returns the error handler middleware function
func (h *ErrorHandler) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				h.Handle(err, c)
			}
			return nil
		}
	}
}

// getStatusCodeForDomainError returns the appropriate HTTP status code for a domain error
func getStatusCodeForDomainError(code errors.ErrorCode) int {
	switch code {
	case errors.ErrCodeValidation:
		return http.StatusBadRequest
	case errors.ErrCodeNotFound:
		return http.StatusNotFound
	case errors.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case errors.ErrCodeForbidden:
		return http.StatusForbidden
	case errors.ErrCodeRequired, errors.ErrCodeInvalid, errors.ErrCodeInvalidFormat, errors.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case errors.ErrCodeInvalidToken, errors.ErrCodeAuthentication:
		return http.StatusUnauthorized
	case errors.ErrCodeInsufficientRole:
		return http.StatusForbidden
	case errors.ErrCodeConflict, errors.ErrCodeAlreadyExists:
		return http.StatusConflict
	case errors.ErrCodeBadRequest:
		return http.StatusBadRequest
	case errors.ErrCodeServerError, errors.ErrCodeDatabase, errors.ErrCodeTimeout:
		return http.StatusInternalServerError
	case errors.ErrCodeStartup, errors.ErrCodeShutdown, errors.ErrCodeConfig:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// handleDomainError handles domain errors and returns appropriate HTTP responses
func handleDomainError(c echo.Context, domainErr *errors.DomainError) error {
	statusCode := getStatusCodeForDomainError(domainErr.Code)
	response := map[string]any{
		"error": domainErr.Error(),
		"code":  domainErr.Code,
	}

	if jsonErr := c.JSON(statusCode, response); jsonErr != nil {
		return fmt.Errorf("failed to send error response: %w", jsonErr)
	}
	return nil
}

// handleHTTPError handles HTTP errors and returns appropriate HTTP responses
func handleHTTPError(c echo.Context, httpErr *echo.HTTPError) error {
	response := map[string]any{
		"error": httpErr.Message,
		"code":  httpErr.Code,
	}

	if jsonErr := c.JSON(httpErr.Code, response); jsonErr != nil {
		return fmt.Errorf("failed to send error response: %w", jsonErr)
	}
	return nil
}

// handleUnknownError handles unknown errors and returns appropriate HTTP responses
func handleUnknownError(c echo.Context) error {
	response := map[string]any{
		"error": "Internal Server Error",
		"code":  http.StatusInternalServerError,
	}

	if jsonErr := c.JSON(http.StatusInternalServerError, response); jsonErr != nil {
		return fmt.Errorf("failed to send error response: %w", jsonErr)
	}
	return nil
}

// ErrorHandlerMiddleware is a middleware that handles errors
func ErrorHandlerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			var domainErr *errors.DomainError
			if stderrors.As(err, &domainErr) {
				return handleDomainError(c, domainErr)
			}

			var httpErr *echo.HTTPError
			if stderrors.As(err, &httpErr) {
				return handleHTTPError(c, httpErr)
			}

			return handleUnknownError(c)
		}
	}
}
