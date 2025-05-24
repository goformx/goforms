package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Logging is middleware that logs requests
type Logging struct {
	logger logging.Logger
}

// NewLogging creates a new logging middleware
func NewLogging(logger logging.Logger) *Logging {
	return &Logging{
		logger: logger,
	}
}

// Handle logs requests and responses
func (m *Logging) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Skip logging for static assets
		path := c.Request().URL.Path
		if strings.HasPrefix(path, "/node_modules/") ||
			strings.HasPrefix(path, "/dist/") ||
			strings.HasPrefix(path, "/public/") {
			return next(c)
		}

		// Log request
		m.logger.Info("request started",
			logging.StringField("method", c.Request().Method),
			logging.StringField("path", path),
			logging.StringField("remote_addr", c.Request().RemoteAddr),
		)

		// Process request
		err := next(c)

		// Log response
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			m.logger.Info("request completed",
				logging.StringField("method", c.Request().Method),
				logging.StringField("path", path),
				logging.IntField("status", httpErr.Code),
				logging.DurationField("duration", time.Since(start)),
			)
			return err
		}

		if err != nil {
			m.logger.Error("request failed",
				logging.StringField("method", c.Request().Method),
				logging.StringField("path", path),
				logging.ErrorField("error", err),
				logging.DurationField("duration", time.Since(start)),
			)
			return err
		}

		m.logger.Info("request completed",
			logging.StringField("method", c.Request().Method),
			logging.StringField("path", path),
			logging.IntField("status", c.Response().Status),
			logging.DurationField("duration", time.Since(start)),
		)

		return nil
	}
}
