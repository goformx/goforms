package middleware

import (
	"fmt"
	"net/http"
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
	config *config.Config
}

// New creates a new middleware manager
func New(logger *zap.Logger, config *config.Config) *Manager {
	return &Manager{
		logger: logger,
		config: config,
	}
}

// Setup configures all middleware for the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	// Recovery middleware should be first for safety
	e.Use(echomw.Recover())

	// Request ID for tracing
	e.Use(echomw.RequestID())

	// Debug logging for detailed request information
	e.Use(m.requestLogger())

	// Structured logging for machine-readable logs
	e.Use(m.structuredLogger())

	// Log initial configuration
	m.logConfig()

	// CORS configuration
	e.Use(m.corsMiddleware())

	// Security headers
	e.Use(m.securityHeaders())

	// Timeout middleware
	e.Use(m.timeoutMiddleware())

	// Rate limiting should be last
	e.Use(m.rateLimiter())

	// Custom error handler
	e.HTTPErrorHandler = m.errorHandler()
}

// requestLogger provides detailed debug logging for requests
func (m *Manager) requestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			m.logger.Debug("incoming request",
				zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
				zap.String("origin", req.Header.Get(echo.HeaderOrigin)),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("remote_addr", c.RealIP()),
				zap.Any("headers", req.Header),
			)
			return next(c)
		}
	}
}

// structuredLogger provides machine-readable JSON logs
func (m *Manager) structuredLogger() echo.MiddlewareFunc {
	return echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","request_id":"${id}","remote_ip":"${remote_ip}",` +
			`"method":"${method}","uri":"${uri}","status":${status},` +
			`"latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		Output: os.Stdout,
	})
}

// corsMiddleware handles CORS configuration
func (m *Manager) corsMiddleware() echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     m.config.Security.CorsAllowedOrigins,
		AllowMethods:     m.config.Security.CorsAllowedMethods,
		AllowHeaders:     m.config.Security.CorsAllowedHeaders,
		AllowCredentials: m.config.Security.CorsAllowCredentials,
		MaxAge:           m.config.Security.CorsMaxAge,
		ExposeHeaders:    []string{echo.HeaderXRequestID}, // Expose Request ID
	})
}

// securityHeaders adds security-related HTTP headers
func (m *Manager) securityHeaders() echo.MiddlewareFunc {
	return echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		HSTSMaxAge:         31536000,
		HSTSPreloadEnabled: true,
	})
}

// timeoutMiddleware adds request timeout handling
func (m *Manager) timeoutMiddleware() echo.MiddlewareFunc {
	return echomw.TimeoutWithConfig(echomw.TimeoutConfig{
		Timeout: m.config.Security.RequestTimeout,
	})
}

// rateLimiter implements rate limiting
func (m *Manager) rateLimiter() echo.MiddlewareFunc {
	return echomw.RateLimiter(echomw.NewRateLimiterMemoryStore(
		rate.Limit(m.config.RateLimit.Rate),
	))
}

// errorHandler provides consistent error responses
func (m *Manager) errorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := "Internal Server Error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = fmt.Sprintf("%v", he.Message)
		}

		// Log the error with context
		m.logger.Error("request error",
			zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			zap.Int("status", code),
			zap.String("message", message),
			zap.Error(err),
			zap.String("path", c.Request().URL.Path),
			zap.String("method", c.Request().Method),
			zap.String("origin", c.Request().Header.Get(echo.HeaderOrigin)),
			zap.String("remote_addr", c.RealIP()),
		)

		// Send error response if not already sent
		if !c.Response().Committed {
			if err := c.JSON(code, map[string]interface{}{
				"error":      message,
				"code":       code,
				"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
				"path":       c.Request().URL.Path,
				"method":     c.Request().Method,
			}); err != nil {
				m.logger.Error("error sending error response", zap.Error(err))
			}
		}
	}
}

// logConfig logs the middleware configuration for debugging
func (m *Manager) logConfig() {
	m.logger.Info("middleware configuration",
		zap.Any("cors", map[string]interface{}{
			"allowed_origins":   m.config.Security.CorsAllowedOrigins,
			"allowed_methods":   m.config.Security.CorsAllowedMethods,
			"allowed_headers":   m.config.Security.CorsAllowedHeaders,
			"max_age":           m.config.Security.CorsMaxAge,
			"allow_credentials": m.config.Security.CorsAllowCredentials,
		}),
		zap.Any("rate_limit", m.config.RateLimit),
		zap.Duration("request_timeout", m.config.Security.RequestTimeout),
	)
}
