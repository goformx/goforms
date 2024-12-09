package middleware

import (
	"fmt"
	"net/http"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Middleware struct {
	logger *zap.Logger
	config *config.Config
}

func New(logger *zap.Logger, cfg *config.Config) *Middleware {
	return &Middleware{
		logger: logger,
		config: cfg,
	}
}

func (m *Middleware) Setup(e *echo.Echo) {
	// Debug logging first
	e.Use(m.requestLogger())

	// CORS middleware
	e.Use(m.corsMiddleware())

	// Recovery middleware
	e.Use(echomw.Recover())

	// Request ID
	e.Use(echomw.RequestID())

	// Security headers
	e.Use(m.securityHeaders())

	// Rate limiting last
	e.Use(m.rateLimiter())

	// Custom error handler
	e.HTTPErrorHandler = m.errorHandler()

	m.logConfig()
}

func (m *Middleware) requestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			m.logger.Debug("incoming request",
				zap.String("origin", c.Request().Header.Get(echo.HeaderOrigin)),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.Any("headers", c.Request().Header),
			)
			return next(c)
		}
	}
}

func (m *Middleware) corsMiddleware() echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     m.config.Security.CorsAllowedOrigins,
		AllowMethods:     m.config.Security.CorsAllowedMethods,
		AllowHeaders:     m.config.Security.CorsAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           m.config.Security.CorsMaxAge,
		ExposeHeaders:    []string{"X-Request-Id"},
	})
}

func (m *Middleware) securityHeaders() echo.MiddlewareFunc {
	return echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
	})
}

func (m *Middleware) rateLimiter() echo.MiddlewareFunc {
	return echomw.RateLimiter(echomw.NewRateLimiterMemoryStore(
		rate.Limit(m.config.RateLimit.Rate)))
}

func (m *Middleware) errorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := "Internal Server Error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = fmt.Sprintf("%v", he.Message)
		}

		m.logger.Error("request error",
			zap.Int("status", code),
			zap.String("message", message),
			zap.Error(err),
			zap.String("path", c.Request().URL.Path),
			zap.String("method", c.Request().Method),
			zap.String("origin", c.Request().Header.Get(echo.HeaderOrigin)),
		)

		if !c.Response().Committed {
			if err := c.JSON(code, map[string]interface{}{
				"error":  message,
				"code":   code,
				"path":   c.Request().URL.Path,
				"method": c.Request().Method,
			}); err != nil {
				m.logger.Error("error sending error response", zap.Error(err))
			}
		}
	}
}

func (m *Middleware) logConfig() {
	m.logger.Info("CORS configuration",
		zap.Strings("allowed_origins", m.config.Security.CorsAllowedOrigins),
		zap.Strings("allowed_methods", m.config.Security.CorsAllowedMethods),
		zap.Strings("allowed_headers", m.config.Security.CorsAllowedHeaders),
		zap.Int("max_age", m.config.Security.CorsMaxAge),
	)
}

// Individual middleware methods...
