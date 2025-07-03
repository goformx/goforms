package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// Recovery returns a middleware that recovers from panics
func Recovery(logger logging.Logger, sanitizer sanitization.ServiceInterface) echo.MiddlewareFunc {
	logger = logger.WithComponent("recovery")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err := handlePanic(r)
					handleError(c, err, logger, sanitizer)
				}
			}()

			return next(c)
		}
	}
}

// sanitizePath sanitizes a URL path for safe logging
func sanitizePath(path string, sanitizer sanitization.ServiceInterface) string {
	if path == "" {
		return ""
	}

	// Use the sanitization service to clean the path
	return sanitizer.SingleLine(path)
}

// handlePanic converts a panic value to an error
func handlePanic(r any) error {
	switch x := r.(type) {
	case string:
		return errors.New(x)
	case error:
		return x
	default:
		return errors.New("unknown panic")
	}
}

// handleError sends an appropriate error response
func handleError(c echo.Context, err error, logger logging.Logger, sanitizer sanitization.ServiceInterface) {
	// Create a logger with request context
	logger = logger.With(
		"request_id", c.Request().Header.Get("X-Request-ID"),
		"method", c.Request().Method,
		"path", sanitizePath(c.Request().URL.Path, sanitizer),
		"remote_addr", c.Request().RemoteAddr,
	)

	var domainErr *domainerrors.DomainError
	if errors.As(err, &domainErr) {
		logger.Error("recovered from panic with domain error",
			"error", err,
			"error_code", domainErr.Code,
			"error_message", domainErr.Message,
			"error_type", "panic_domain_error",
		)

		statusCode := domainerrors.GetHTTPStatus(domainErr.Code)
		if jsonErr := response.ErrorResponse(c, statusCode, domainErr.Message); jsonErr != nil {
			logger.Error("failed to send error response",
				"error", jsonErr,
				"error_type", "response_error",
				"original_error", err,
			)
		}

		return
	}

	// Handle unknown errors
	logger.Error("recovered from panic with unknown error",
		"error", err,
		"error_type", "panic_unknown_error",
	)

	if jsonErr := response.ErrorResponse(c, http.StatusInternalServerError, "Internal server error"); jsonErr != nil {
		logger.Error("failed to send error response",
			"error", jsonErr,
			"error_type", "response_error",
			"original_error", err,
		)
	}
}
