package middleware

import (
	"net/http"

	"errors"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// Recovery returns a middleware that recovers from panics
func Recovery(logger logging.Logger) echo.MiddlewareFunc {
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
func handlePanic(r interface{}) error {
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
	var domainErr *domainerrors.DomainError
	if errors.As(err, &domainErr) {
		statusCode := getStatusCode(domainErr.Code)
		if err := c.JSON(statusCode, domainErr); err != nil {
			logger.Error("failed to send error response", logging.Error(err))
		}
		return
	}

	// Handle unknown errors
	if err := c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Internal Server Error",
	}); err != nil {
		logger.Error("failed to send error response", logging.Error(err))
	}
}

// getStatusCode returns the appropriate HTTP status code for an error code
func getStatusCode(code domainerrors.ErrorCode) int {
	switch code {
	case domainerrors.ErrCodeNotFound:
		return http.StatusNotFound
	case domainerrors.ErrCodeInvalid,
		domainerrors.ErrCodeInvalidFormat,
		domainerrors.ErrCodeInvalidInput,
		domainerrors.ErrCodeBadRequest:
		return http.StatusBadRequest
	case domainerrors.ErrCodeInvalidToken,
		domainerrors.ErrCodeAuthentication:
		return http.StatusUnauthorized
	case domainerrors.ErrCodeInsufficientRole:
		return http.StatusForbidden
	case domainerrors.ErrCodeConflict,
		domainerrors.ErrCodeAlreadyExists:
		return http.StatusConflict
	case domainerrors.ErrCodeStartup,
		domainerrors.ErrCodeShutdown:
		return http.StatusServiceUnavailable
	case domainerrors.ErrCodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
