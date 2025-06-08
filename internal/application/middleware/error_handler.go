package middleware

import (
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
	if domainErr, ok := err.(*errors.DomainError); ok {
		h.logger.Error("domain error",
			logging.ErrorField("error", domainErr),
			logging.StringField("code", string(domainErr.Code)),
		)

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
		}

		response := map[string]any{
			"error": map[string]any{
				"code":    domainErr.Code,
				"message": domainErr.Message,
			},
		}

		if err := c.JSON(statusCode, response); err != nil {
			h.logger.Error("failed to send error response",
				logging.ErrorField("error", err),
			)
		}

		return
	}

	// Handle HTTP errors
	if httpErr, ok := err.(*echo.HTTPError); ok {
		h.logger.Error("http error",
			logging.ErrorField("error", httpErr),
			logging.IntField("code", httpErr.Code),
		)

		response := map[string]any{
			"error": map[string]any{
				"code":    "HTTP_ERROR",
				"message": httpErr.Message,
			},
		}

		if err := c.JSON(httpErr.Code, response); err != nil {
			h.logger.Error("failed to send error response",
				logging.ErrorField("error", err),
			)
		}

		return
	}

	// Handle unknown errors
	h.logger.Error("unknown error",
		logging.ErrorField("error", err),
	)

	response := map[string]any{
		"error": map[string]any{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	}

	if err := c.JSON(http.StatusInternalServerError, response); err != nil {
		h.logger.Error("failed to send error response",
			logging.ErrorField("error", err),
		)
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
