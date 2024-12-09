package middleware

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
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
	// Log middleware configuration
	m.logger.Info("middleware configuration",
		zap.Any("cors", map[string]interface{}{
			"allowed_origins":   m.config.Security.CorsAllowedOrigins,
			"allowed_methods":   m.config.Security.CorsAllowedMethods,
			"allowed_headers":   m.config.Security.CorsAllowedHeaders,
			"allow_credentials": m.config.Security.CorsAllowCredentials,
			"max_age":           m.config.Security.CorsMaxAge,
		}),
		zap.Any("rate_limit", m.config.RateLimit),
		zap.String("request_timeout", m.config.Security.RequestTimeout.String()),
	)

	// Configure CORS
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     m.config.Security.CorsAllowedOrigins,
		AllowMethods:     m.config.Security.CorsAllowedMethods,
		AllowHeaders:     m.config.Security.CorsAllowedHeaders,
		AllowCredentials: m.config.Security.CorsAllowCredentials,
		MaxAge:           m.config.Security.CorsMaxAge,
		ExposeHeaders:    []string{echo.HeaderXRequestID}, // Expose Request ID
	}))

	// Add request ID middleware
	e.Use(echomw.RequestIDWithConfig(echomw.RequestIDConfig{
		Generator: func() string {
			return uuid.New().String()
		},
	}))

	// Add request logger
	e.Use(m.requestLogger())

	// Add rate limiter if enabled
	if m.config.RateLimit.Enabled {
		e.Use(m.rateLimiter())
	}

	// Add timeout middleware
	e.Use(m.timeoutMiddleware())

	// Add panic recovery
	e.Use(echomw.Recover())

	// Set custom error handler
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
		var code int
		var message string

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = fmt.Sprintf("%v", he.Message)
		} else {
			code = http.StatusInternalServerError
			message = "Internal Server Error"
		}

		// Log the error
		m.logger.Error("request error",
			zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			zap.Int("status", code),
			zap.String("message", message),
			zap.Error(err),
			zap.String("path", c.Request().URL.Path),
			zap.String("method", c.Request().Method),
		)

		// Send error response
		if !c.Response().Committed {
			if err := c.JSON(code, map[string]interface{}{
				"error":      message,
				"code":       code,
				"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
			}); err != nil {
				m.logger.Error("failed to send error response", zap.Error(err))
			}
		}
	}
}
