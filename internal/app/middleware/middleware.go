package middleware

import (
	"os"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Manager handles all middleware setup and configuration
type Manager struct {
	logger *zap.Logger
	config *config.SecurityConfig
}

// New creates a new middleware manager
func New(logger *zap.Logger, config *config.SecurityConfig) *Manager {
	return &Manager{
		logger: logger,
		config: config,
	}
}

// Setup configures all middleware for the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	// Recovery middleware should be first
	e.Use(echomw.Recover())

	// Request ID for tracing
	e.Use(echomw.RequestID())

	// Structured logging
	e.Use(m.structuredLogger())

	// Security headers
	e.Use(echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
	}))

	// CORS
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: m.config.CorsAllowedOrigins,
		AllowMethods: m.config.CorsAllowedMethods,
		AllowHeaders: m.config.CorsAllowedHeaders,
		MaxAge:       m.config.CorsMaxAge,
	}))

	// Rate limiting
	e.Use(echomw.RateLimiter(echomw.NewRateLimiterMemoryStore(rate.Limit(20))))
}

// structuredLogger returns a middleware function that logs requests in structured format
func (m *Manager) structuredLogger() echo.MiddlewareFunc {
	return echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","remote_ip":"${remote_ip}",` +
			`"method":"${method}","uri":"${uri}","status":${status},` +
			`"latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		Output: os.Stdout,
	})
}
