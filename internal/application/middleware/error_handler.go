package middleware

import (
	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// ErrorHandler is a middleware that handles errors and returns appropriate HTTP responses
type ErrorHandler struct {
	logger logging.Logger
}

// NewErrorHandler creates a new error handler middleware
func NewErrorHandler(logger logging.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// Handle handles errors and returns appropriate HTTP responses
func (h *ErrorHandler) Handle(err error, c echo.Context) {
	// Log the error
	h.logger.Error("request error",
		logging.ErrorField("error", err),
		logging.StringField("path", c.Request().URL.Path),
		logging.StringField("method", c.Request().Method),
	)

	// Translate the error to an HTTP error
	httpErr := errors.TranslateToHTTP(err)

	// Send the error response
	if jsonErr := c.JSON(httpErr.Code, httpErr); jsonErr != nil {
		h.logger.Error("failed to send error response", logging.Error(jsonErr))
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
