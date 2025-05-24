package middleware

import (
	"net/http"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// ErrorLogger is middleware that logs errors
type ErrorLogger struct {
	logger logging.Logger
}

// NewErrorLogger creates a new error logger middleware
func NewErrorLogger(logger logging.Logger) *ErrorLogger {
	return &ErrorLogger{
		logger: logger,
	}
}

// Handle logs errors and returns appropriate responses
func (m *ErrorLogger) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		// Get the domain error if it exists
		domainErr := errors.GetDomainError(err)
		if domainErr != nil {
			// Log the error with context
			m.logger.Error("request error",
				logging.StringField("error_code", string(domainErr.Code)),
				logging.StringField("error_message", domainErr.Message),
				logging.ErrorField("error", domainErr.Err),
				logging.StringField("path", c.Request().URL.Path),
				logging.StringField("method", c.Request().Method),
			)

			// Map domain errors to HTTP status codes
			status := mapErrorToStatus(domainErr.Code)
			return c.JSON(status, map[string]interface{}{
				"error": domainErr.Message,
				"code":  domainErr.Code,
			})
		}

		// Handle echo errors
		if he, ok := err.(*echo.HTTPError); ok {
			m.logger.Error("http error",
				logging.IntField("status", he.Code),
				logging.StringField("message", he.Message.(string)),
				logging.StringField("path", c.Request().URL.Path),
				logging.StringField("method", c.Request().Method),
			)
			return c.JSON(he.Code, map[string]interface{}{
				"error": he.Message,
			})
		}

		// Handle unknown errors
		m.logger.Error("unknown error",
			logging.ErrorField("error", err),
			logging.StringField("path", c.Request().URL.Path),
			logging.StringField("method", c.Request().Method),
		)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Internal server error",
		})
	}
}

// mapErrorToStatus maps domain error codes to HTTP status codes
func mapErrorToStatus(code errors.ErrorCode) int {
	switch code {
	case errors.ErrCodeValidation, errors.ErrCodeRequired, errors.ErrCodeInvalid,
		errors.ErrCodeInvalidFormat, errors.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case errors.ErrCodeUnauthorized, errors.ErrCodeInvalidToken,
		errors.ErrCodeAuthentication:
		return http.StatusUnauthorized
	case errors.ErrCodeForbidden, errors.ErrCodeInsufficientRole:
		return http.StatusForbidden
	case errors.ErrCodeNotFound:
		return http.StatusNotFound
	case errors.ErrCodeConflict, errors.ErrCodeAlreadyExists:
		return http.StatusConflict
	case errors.ErrCodeBadRequest:
		return http.StatusBadRequest
	case errors.ErrCodeServerError, errors.ErrCodeDatabase,
		errors.ErrCodeStartup, errors.ErrCodeShutdown, errors.ErrCodeConfig:
		return http.StatusInternalServerError
	case errors.ErrCodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
