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
		logger: logger.WithComponent("error_handler"),
	}
}

// handleDomainError handles domain-specific errors
func (h *ErrorHandler) handleDomainError(err error, c echo.Context, logger logging.Logger) {
	var domainErr *errors.DomainError
	if stderrors.As(err, &domainErr) {
		logger.Error("domain error",
			"error", err,
			"error_code", domainErr.Code,
			"error_message", domainErr.Message,
			"error_type", "domain_error",
		)

		statusCode := errors.GetHTTPStatus(domainErr.Code)

		response := map[string]any{
			"error": map[string]any{
				"code":    domainErr.Code,
				"message": domainErr.Message,
			},
		}

		if jsonErr := c.JSON(statusCode, response); jsonErr != nil {
			logger.Error("failed to send error response",
				"error", jsonErr,
				"error_type", "response_error",
				"original_error", err,
			)
		}
	}
}

// handleHTTPError handles HTTP-specific errors
func (h *ErrorHandler) handleHTTPError(err error, c echo.Context, logger logging.Logger) {
	var httpErr *echo.HTTPError
	if stderrors.As(err, &httpErr) {
		logger.Error("http error",
			"error", err,
			"error_code", httpErr.Code,
			"error_message", httpErr.Message,
			"error_type", "http_error",
		)

		response := map[string]any{
			"error": map[string]any{
				"code":    "HTTP_ERROR",
				"message": httpErr.Message,
			},
		}

		if jsonErr := c.JSON(httpErr.Code, response); jsonErr != nil {
			logger.Error("failed to send error response",
				"error", jsonErr,
				"error_type", "response_error",
				"original_error", err,
			)
		}
	}
}

// handleUnknownError handles unknown errors
func (h *ErrorHandler) handleUnknownError(err error, c echo.Context, logger logging.Logger) {
	logger.Error("unknown error",
		"error", err,
		"error_type", "unknown_error",
	)

	response := map[string]any{
		"error": map[string]any{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	}

	if jsonErr := c.JSON(http.StatusInternalServerError, response); jsonErr != nil {
		logger.Error("failed to send error response",
			"error", jsonErr,
			"error_type", "response_error",
			"original_error", err,
		)
	}
}

// Handle handles errors and returns appropriate HTTP responses
func (h *ErrorHandler) Handle(err error, c echo.Context) {
	if err == nil {
		return
	}

	// Create a logger with request context
	logger := h.logger.With(
		"request_id", c.Request().Header.Get("X-Request-ID"),
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"remote_addr", c.Request().RemoteAddr,
	)

	// Handle domain errors
	h.handleDomainError(err, c, logger)

	// Handle HTTP errors
	h.handleHTTPError(err, c, logger)

	// Handle unknown errors
	h.handleUnknownError(err, c, logger)
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

// handleDomainError handles domain errors and returns appropriate HTTP status code
func handleDomainError(c echo.Context, err *errors.DomainError) error {
	statusCode := errors.GetHTTPStatus(err.Code)
	return c.JSON(statusCode, err)
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
