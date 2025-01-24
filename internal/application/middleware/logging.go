package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/logger"
)

func LoggingMiddleware(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Call the next handler
			err := next(c)

			// Log the request details
			log.Info("http request",
				logger.String("method", c.Request().Method),
				logger.String("path", c.Request().URL.Path),
				logger.Int("status", c.Response().Status),
				logger.Duration("latency", time.Since(start)),
				logger.String("ip", c.RealIP()),
			)

			return err
		}
	}
}
