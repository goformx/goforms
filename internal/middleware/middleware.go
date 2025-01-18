package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/labstack/echo/v4"
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger logger.Logger
	config *config.Config
}

// New creates a new middleware manager
func New(log logger.Logger, cfg *config.Config) *Manager {
	return &Manager{
		logger: log,
		config: cfg,
	}
}

// Setup configures middleware for the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Info("middleware configuration",
		logger.String("cors", formatConfig(m.config.Security.CorsAllowedOrigins)),
		logger.String("rate_limit", formatConfig(m.config.RateLimit)),
		logger.String("request_timeout", m.config.Security.RequestTimeout.String()),
	)

	// Add request ID middleware
	e.Use(m.requestID())
}

// requestID adds a unique request ID to each request
func (m *Manager) requestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := uuid.New().String()
			c.Set("request_id", requestID)

			m.logger.Debug("incoming request",
				logger.String("request_id", requestID),
				logger.String("origin", c.Request().Header.Get("Origin")),
				logger.String("method", c.Request().Method),
				logger.String("path", c.Request().URL.Path),
				logger.String("remote_addr", c.Request().RemoteAddr),
				logger.String("headers", formatHeaders(c.Request().Header)),
			)

			return next(c)
		}
	}
}

// formatConfig converts a config value to a string representation
func formatConfig(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// formatHeaders converts HTTP headers to a string representation
func formatHeaders(h http.Header) string {
	b, err := json.Marshal(h)
	if err != nil {
		return ""
	}
	return string(b)
}
