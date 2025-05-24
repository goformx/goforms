package middleware

import (
	stderrors "errors"
	"net/http"

	derr "github.com/goformx/goforms/internal/domain/common/errors"
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
		domainErr := derr.GetDomainError(err)
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
			return c.JSON(status, map[string]any{
				"error": domainErr.Message,
				"code":  domainErr.Code,
			})
		}

		// Handle echo errors
		var he *echo.HTTPError
		if stderrors.As(err, &he) {
			msg, ok := he.Message.(string)
			if !ok {
				msg = "Internal server error"
			}
			m.logger.Error("http error",
				logging.IntField("status", he.Code),
				logging.StringField("message", msg),
				logging.StringField("path", c.Request().URL.Path),
				logging.StringField("method", c.Request().Method),
			)
			return c.JSON(he.Code, map[string]any{
				"error": msg,
				"code":  he.Code,
			})
		}

		// Handle unknown errors
		m.logger.Error("unknown error",
			logging.ErrorField("error", err),
			logging.StringField("path", c.Request().URL.Path),
			logging.StringField("method", c.Request().Method),
		)
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Internal server error",
			"code":  http.StatusInternalServerError,
		})
	}
}

// mapErrorToStatus maps domain error codes to HTTP status codes
func mapErrorToStatus(code derr.ErrorCode) int {
	switch code {
	case derr.ErrCodeValidation, derr.ErrCodeRequired, derr.ErrCodeInvalid,
		derr.ErrCodeInvalidFormat, derr.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case derr.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case derr.ErrCodeForbidden:
		return http.StatusForbidden
	case derr.ErrCodeNotFound:
		return http.StatusNotFound
	case derr.ErrCodeConflict, derr.ErrCodeAlreadyExists:
		return http.StatusConflict
	case derr.ErrCodeServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
