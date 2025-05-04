package middleware

import (
	"errors"
	"fmt"
	"strings"
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
				status := c.Response().Status
				path := c.Request().URL.Path
				isStatic404 := status == 404 && (strings.HasPrefix(path, "/node_modules/") ||
					strings.HasPrefix(path, "/dist/") ||
					strings.HasPrefix(path, "/public/") ||
					strings.HasSuffix(path, ".woff2") ||
					strings.HasSuffix(path, ".woff") ||
					strings.HasSuffix(path, ".ttf") ||
					strings.HasSuffix(path, ".eot") ||
					strings.HasSuffix(path, ".svg") ||
					strings.HasSuffix(path, ".png") ||
					strings.HasSuffix(path, ".jpg") ||
					strings.HasSuffix(path, ".jpeg") ||
					strings.HasSuffix(path, ".gif"))
				if isStatic404 {
					log.Info("static asset not found", fields...)
				} else {
					log.Error("request error", append(fields, logging.Error(err))...)
				}
			} else {
				log.Info("request", fields...)
			}

			return err
		}
	}
}
