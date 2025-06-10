package middleware

import (
	"errors"
	"net/http"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// Recovery returns a middleware that recovers from panics
func Recovery(logger logging.Logger) echo.MiddlewareFunc {
	logger = logger.WithComponent("recovery")
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err := handlePanic(r)
					handleError(c, err, logger)
				}
			}()
			return next(c)
		}
	}
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
func handleError(c echo.Context, err error, logger logging.Logger) {
	// Create a logger with request context
	logger = logger.With(
		"request_id", c.Request().Header.Get("X-Request-ID"),
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
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

		statusCode := getStatusCode(domainErr.Code)
		if jsonErr := c.JSON(statusCode, domainErr); jsonErr != nil {
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

	if jsonErr := c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Internal Server Error",
	}); jsonErr != nil {
		logger.Error("failed to send error response",
			"error", jsonErr,
			"error_type", "response_error",
			"original_error", err,
		)
	}
}

// getStatusCode returns the appropriate HTTP status code for an error code
func getStatusCode(code domainerrors.ErrorCode) int {
	switch code {
	case domainerrors.ErrCodeValidation:
		return http.StatusBadRequest
	case domainerrors.ErrCodeNotFound:
		return http.StatusNotFound
	case domainerrors.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case domainerrors.ErrCodeForbidden:
		return http.StatusForbidden
	case domainerrors.ErrCodeRequired, domainerrors.ErrCodeInvalid, domainerrors.ErrCodeInvalidFormat, domainerrors.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case domainerrors.ErrCodeInvalidToken, domainerrors.ErrCodeAuthentication:
		return http.StatusUnauthorized
	case domainerrors.ErrCodeInsufficientRole:
		return http.StatusForbidden
	case domainerrors.ErrCodeConflict, domainerrors.ErrCodeAlreadyExists:
		return http.StatusConflict
	case domainerrors.ErrCodeBadRequest:
		return http.StatusBadRequest
	case domainerrors.ErrCodeServerError, domainerrors.ErrCodeDatabase, domainerrors.ErrCodeTimeout:
		return http.StatusInternalServerError
	case domainerrors.ErrCodeStartup, domainerrors.ErrCodeShutdown, domainerrors.ErrCodeConfig:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
