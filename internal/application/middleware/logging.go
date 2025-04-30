package middleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

func LoggingMiddleware(log logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			var err error

			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic: %v", r)
					c.Response().Status = echo.ErrInternalServerError.Code
					log.Error("request panic",
						logging.String("method", c.Request().Method),
						logging.String("path", c.Request().URL.Path),
						logging.Int("status", c.Response().Status),
						logging.Duration("latency", time.Since(start)),
						logging.String("ip", c.RealIP()),
						logging.String("user_agent", c.Request().UserAgent()),
						logging.Any("panic", r),
					)
					c.Error(err)
				}
			}()

			// Call the next handler
			err = next(c)

			// Set status based on error
			if err != nil {
				he := &echo.HTTPError{}
				if errors.As(err, &he) {
					c.Response().Status = he.Code
				} else {
					c.Response().Status = echo.ErrInternalServerError.Code
				}
			}

			// Log the request details using structured logging
			fields := []logging.Field{
				logging.String("method", c.Request().Method),
				logging.String("path", c.Request().URL.Path),
				logging.Int("status", c.Response().Status),
				logging.Duration("latency", time.Since(start)),
				logging.String("remote_ip", c.RealIP()),
				logging.String("user_agent", c.Request().UserAgent()),
			}

			if err != nil {
				log.Error("request error", append(fields, logging.Error(err))...)
			} else {
				log.Info("request", fields...)
			}

			return err
		}
	}
}
