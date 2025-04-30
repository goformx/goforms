package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4/middleware"
)

// Config holds middleware configuration
type Config struct {
	Logger      logging.Logger
	UserService user.Service
	Security    *config.SecurityConfig
}

// Setup configures all middleware for an Echo instance
func Setup(e *echo.Echo, cfg *Config) {
	cfg.Logger.Debug("starting middleware setup")

	// Logging must be first to capture all requests
	cfg.Logger.Debug("adding logging middleware")
	e.Use(LoggingMiddleware(cfg.Logger))

	// Basic middleware
	cfg.Logger.Debug("adding basic middleware")
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Secure())
	e.Use(middleware.BodyLimit("2M"))

	// Security middleware with comprehensive configuration
	cfg.Logger.Debug("adding security headers middleware")
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            HSTSOneYear,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data:; " +
			"font-src 'self'; " +
			"connect-src 'self'",
		ReferrerPolicy: "strict-origin-when-cross-origin",
	}))

	// CORS
	cfg.Logger.Debug("adding CORS middleware")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.Security.CorsAllowedOrigins,
		AllowMethods:     cfg.Security.CorsAllowedMethods,
		AllowHeaders:     cfg.Security.CorsAllowedHeaders,
		AllowCredentials: cfg.Security.CorsAllowCredentials,
		MaxAge:           cfg.Security.CorsMaxAge,
	}))

	// CSRF if enabled
	if cfg.Security.CSRF.Enabled {
		cfg.Logger.Debug("adding CSRF middleware", 
			logging.Bool("config_enabled", cfg.Security.CSRF.Enabled))
		
		e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLength:    DefaultTokenLength,
			TokenLookup:    "header:X-CSRF-Token",
			ContextKey:     "csrf",
			CookieName:     "csrf_token",
			CookiePath:     "/",
			CookieSecure:   true,
			CookieHTTPOnly: true,
			CookieSameSite: http.SameSiteStrictMode,
			ErrorHandler: func(err error, c echo.Context) error {
				cfg.Logger.Error("CSRF token validation failed", 
					logging.Error(err),
					logging.String("path", c.Request().URL.Path),
					logging.String("method", c.Request().Method))
				return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
			},
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				method := c.Request().Method

				cfg.Logger.Debug("CSRF middleware evaluating request", 
					logging.String("path", path),
					logging.String("method", method))

				// Skip for static content
				if strings.HasPrefix(path, "/static/") || 
				   strings.HasPrefix(path, "/favicon.ico") ||
				   strings.HasPrefix(path, "/robots.txt") {
					cfg.Logger.Debug("CSRF skipped: static content", 
						logging.String("path", path))
					return true
				}

				// Skip for API routes that use proper authentication
				if strings.HasPrefix(path, "/api/") {
					// Check if the route requires authentication
					authHeader := c.Request().Header.Get("Authorization")
					if authHeader != "" {
						cfg.Logger.Debug("CSRF skipped: authenticated API route", 
							logging.String("path", path))
						return true
					}
				}

				// Always generate tokens for pages with forms
				if strings.HasPrefix(path, "/login") || 
				   strings.HasPrefix(path, "/signup") || 
				   strings.HasPrefix(path, "/forgot-password") ||
				   strings.HasPrefix(path, "/contact") ||
				   strings.HasPrefix(path, "/demo") {
					cfg.Logger.Debug("CSRF not skipped: page with form", 
						logging.String("path", path),
						logging.String("method", method))
					return false
				}

				// Generate tokens for all other pages by default
				// This ensures AJAX requests will have tokens available
				cfg.Logger.Debug("CSRF not skipped: default case", 
					logging.String("path", path),
					logging.String("method", method))
				return false
			},
		}))
	} else {
		cfg.Logger.Debug("CSRF middleware is disabled", 
			logging.Bool("config_enabled", cfg.Security.CSRF.Enabled))
	}

	// Auth if secret provided
	if cfg.UserService != nil {
		cfg.Logger.Debug("setting up JWT middleware")
		middleware, err := NewJWTMiddleware(cfg.UserService, cfg.Security.JWTSecret)
		if err != nil {
			cfg.Logger.Error("failed to create JWT middleware", logging.Error(err))
			return
		}
		e.Use(middleware)
	}

	cfg.Logger.Debug("middleware setup complete")
}
