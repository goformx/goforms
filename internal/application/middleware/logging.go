package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"runtime/debug"

	"path/filepath"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var staticFileExtensions = map[string]bool{
	".woff2": true, ".woff": true, ".ttf": true, ".eot": true,
	".svg": true, ".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
}

func handlePanic(c echo.Context, log logging.Logger, start time.Time, r any) {
	err := fmt.Errorf("panic: %v", r)
	c.Response().Status = http.StatusInternalServerError
	log.Error("request panic",
		logging.String("method", c.Request().Method),
		logging.String("path", c.Request().URL.Path),
		logging.Int("status", c.Response().Status),
		logging.String("latency", time.Since(start).String()),
		logging.String("ip", c.RealIP()),
		logging.String("user_agent", c.Request().UserAgent()),
		logging.String("stack", string(debug.Stack())),
		logging.Any("panic", r),
	)
	c.Error(err)
}

// extractFields builds the log fields for a request
func extractFields(c echo.Context, start time.Time) []any {
	return []any{
		logging.String("method", c.Request().Method),
		logging.String("path", c.Request().URL.Path),
		logging.Int("status", c.Response().Status),
		logging.String("latency", time.Since(start).String()),
		logging.String("remote_ip", c.RealIP()),
		logging.String("user_agent", c.Request().UserAgent()),
	}
}

// handleErrorStatus sets the response status based on error
func handleErrorStatus(c echo.Context, err error) {
	if httpErr, ok := err.(*echo.HTTPError); ok {
		c.Response().Status = httpErr.Code
	} else if err != nil {
		c.Response().Status = http.StatusInternalServerError
	}
}

// isStatic404 checks if a 404 error is for a static file
func isStatic404(path string, status int) bool {
	if status != http.StatusNotFound {
		return false
	}
	if strings.HasPrefix(path, "/node_modules/") || strings.HasPrefix(path, "/dist/") || strings.HasPrefix(path, "/public/") {
		return true
	}
	ext := filepath.Ext(path)
	return staticFileExtensions[ext]
}

// LoggingMiddleware creates a middleware that logs HTTP requests and responses
func LoggingMiddleware(log logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			var err error

			defer func() {
				if r := recover(); r != nil {
					handlePanic(c, log, start, r)
				}
			}()

			// Call the next handler
			err = next(c)

			// Set status based on error
			handleErrorStatus(c, err)

			// Log the request details using structured logging
			fields := extractFields(c, start)

			if err != nil {
				status := c.Response().Status
				path := c.Request().URL.Path
				if isStatic404(path, status) {
					log.Info("static asset not found", fields...)
				} else {
					log.Error("request error", append(fields, logging.Error(err))...)
				}
			} else {
				log.Info("request completed", fields...)
			}

			return err
		}
	}
}
