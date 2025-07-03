package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// ErrorHandlerMiddleware returns a middleware that uses the unified error handler
func ErrorHandlerMiddleware(logger logging.Logger, sanitizer sanitization.ServiceInterface) echo.MiddlewareFunc {
	handler := response.NewErrorHandler(logger, sanitizer)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				return handler.HandleError(err, c, "An unexpected error occurred")
			}

			return nil
		}
	}
}
