package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

func LoggingMiddleware(log logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Call the next handler
			err := next(c)

			// Log the request details
			log.Info("http request",
				logging.String("method", c.Request().Method),
				logging.String("path", c.Request().URL.Path),
				logging.Int("status", c.Response().Status),
				logging.Duration("latency", time.Since(start)),
				logging.String("ip", c.RealIP()),
			)

			return err
		}
	}
}
