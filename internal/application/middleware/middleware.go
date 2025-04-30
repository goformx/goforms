package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

const (
	// NonceSize is the size of the nonce in bytes (32 bytes = 256 bits)
	NonceSize = 32
	// HSTSOneYear is the number of seconds in one year
	HSTSOneYear        = 31536000
	DefaultTokenLength = 32
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger logging.Logger
	config *ManagerConfig
}

// ManagerConfig holds middleware configuration
type ManagerConfig struct {
	Logger      logging.Logger
	UserService user.Service
	Security    *config.SecurityConfig
}

// New creates a new middleware manager
func New(cfg *ManagerConfig) *Manager {
	if cfg.Logger == nil {
		panic("logger is required for Manager")
	}

	return &Manager{
		logger: cfg.Logger,
		config: cfg,
	}
}

// Setup configures all middleware for an Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Debug("setting up middleware manager")

	// Basic middleware
	m.logger.Debug("adding basic middleware")
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.Secure())
	e.Use(echomw.BodyLimit("2M"))

	// Request logging middleware
	m.logger.Debug("adding request logging middleware")
	e.Use(LoggingMiddleware(m.logger))

	// Security middleware with comprehensive configuration
	m.logger.Debug("adding security headers middleware")
	e.Use(echomw.SecureWithConfig(echomw.SecureConfig{
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
			"connect-src 'self' *;", // Allow connections to any origin for form submissions
		ReferrerPolicy: "strict-origin-when-cross-origin",
	}))

	// CORS for admin/dashboard routes
	m.logger.Debug("adding CORS middleware")
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     m.config.Security.CorsAllowedOrigins,
		AllowMethods:     m.config.Security.CorsAllowedMethods,
		AllowHeaders:     m.config.Security.CorsAllowedHeaders,
		AllowCredentials: m.config.Security.CorsAllowCredentials,
		MaxAge:           m.config.Security.CorsMaxAge,
	}))

	// Form submission routes group with specific middleware
	formGroup := e.Group("/v1/forms")
	
	// Form-specific CORS
	formGroup.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     m.config.Security.FormCorsAllowedOrigins,
		AllowMethods:     m.config.Security.FormCorsAllowedMethods,
		AllowHeaders:     m.config.Security.FormCorsAllowedHeaders,
		AllowCredentials: false, // No credentials needed for form submissions
		MaxAge:           m.config.Security.CorsMaxAge,
	}))

	// Rate limiting for form submissions
	formGroup.Use(echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(m.config.Security.FormRateLimit),
				Burst:     5,
				ExpiresIn: m.config.Security.FormRateLimitWindow,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// Rate limit by form ID and origin
			formID := c.Param("formID")
			origin := c.Request().Header.Get("Origin")
			if formID == "" {
				formID = "unknown"
			}
			if origin == "" {
				origin = "unknown"
			}
			return fmt.Sprintf("%s:%s", formID, origin), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
		},
	}))

	// CSRF if enabled
	if m.config.Security.CSRF.Enabled {
		m.logger.Debug("adding CSRF middleware", 
			logging.Bool("config_enabled", m.config.Security.CSRF.Enabled),
			logging.String("secret_length", fmt.Sprintf("%d", len(m.config.Security.CSRF.Secret))))
		
		// Create CSRF middleware with logging
		csrfMiddleware := echomw.CSRFWithConfig(echomw.CSRFConfig{
			TokenLength:    DefaultTokenLength,
			TokenLookup:    "header:X-CSRF-Token,form:csrf_token,cookie:csrf_token",
			ContextKey:     CSRFContextKey,
			CookieName:     "csrf_token",
			CookiePath:     "/",
			CookieSecure:   true,
			CookieHTTPOnly: true,
			CookieSameSite: http.SameSiteStrictMode,
			CookieMaxAge:   86400, // 24 hours
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				method := c.Request().Method

				m.logger.Debug("CSRF middleware evaluating request", 
					logging.String("path", path),
					logging.String("method", method))

				// Skip for static content
				if strings.HasPrefix(path, "/static/") || 
				   strings.HasPrefix(path, "/favicon.ico") ||
				   strings.HasPrefix(path, "/robots.txt") {
					m.logger.Debug("CSRF skipped: static content", 
						logging.String("path", path))
					return true
				}

				// Skip for form submission endpoints
				if strings.HasPrefix(path, "/v1/forms/") {
					m.logger.Debug("CSRF skipped: form submission endpoint", 
						logging.String("path", path))
					return true
				}

				// Skip for API routes that use proper authentication
				if strings.HasPrefix(path, "/api/") {
					authHeader := c.Request().Header.Get("Authorization")
					if authHeader != "" {
						m.logger.Debug("CSRF skipped: authenticated API route", 
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
					m.logger.Debug("CSRF not skipped: page with form", 
						logging.String("path", path),
						logging.String("method", method))
					return false
				}

				// Generate tokens for all other pages by default
				m.logger.Debug("CSRF not skipped: default case", 
					logging.String("path", path),
					logging.String("method", method))
				return false
			},
			ErrorHandler: func(err error, c echo.Context) error {
				m.logger.Error("CSRF token validation failed", 
					logging.Error(err),
					logging.String("path", c.Request().URL.Path),
					logging.String("method", c.Request().Method),
					logging.String("token", c.Get(CSRFContextKey).(string)))
				return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
			},
		})

		m.logger.Debug("CSRF middleware created, adding to Echo instance")
		e.Use(csrfMiddleware)

		// Add logging middleware after CSRF
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Log request details
				m.logger.Debug("CSRF middleware processing request",
					logging.String("path", c.Request().URL.Path),
					logging.String("method", c.Request().Method),
					logging.String("content_type", c.Request().Header.Get("Content-Type")),
					logging.String("user_agent", c.Request().UserAgent()))

				// Log all context keys before CSRF processing
				keys := make([]string, 0)
				for k := range c.Get("").(map[string]interface{}) {
					keys = append(keys, k)
				}
				m.logger.Debug("Context keys before CSRF processing",
					logging.String("path", c.Request().URL.Path),
					logging.String("method", c.Request().Method),
					logging.String("keys", fmt.Sprintf("%v", keys)))

				// Call next middleware
				err := next(c)

				// Log all context keys after CSRF processing
				keys = make([]string, 0)
				for k := range c.Get("").(map[string]interface{}) {
					keys = append(keys, k)
				}
				m.logger.Debug("Context keys after CSRF processing",
					logging.String("path", c.Request().URL.Path),
					logging.String("method", c.Request().Method),
					logging.String("keys", fmt.Sprintf("%v", keys)))

				// Log the generated token if one exists
				if token := c.Get(CSRFContextKey); token != nil {
					if tokenStr, ok := token.(string); ok && tokenStr != "" {
						// Determine token source
						var tokenSource string
						if c.Request().Header.Get("X-CSRF-Token") != "" {
							tokenSource = "header"
						} else if c.Request().FormValue("csrf_token") != "" {
							tokenSource = "form"
						} else {
							tokenSource = "cookie"
						}

						m.logger.Debug("CSRF token generated", 
							logging.String("path", c.Request().URL.Path),
							logging.String("method", c.Request().Method),
							logging.String("token_prefix", tokenStr[:8]),
							logging.String("token_source", tokenSource),
							logging.String("token_length", fmt.Sprintf("%d", len(tokenStr))))

						// Set the token in the context for templates using the same key
						c.Set(CSRFContextKey, tokenStr)
					} else {
						m.logger.Debug("CSRF token exists in context but is not a string",
							logging.String("path", c.Request().URL.Path),
							logging.String("method", c.Request().Method),
							logging.String("token_type", fmt.Sprintf("%T", token)))
					}
				} else {
					m.logger.Debug("No CSRF token in context after middleware", 
						logging.String("path", c.Request().URL.Path),
						logging.String("method", c.Request().Method),
						logging.String("keys", fmt.Sprintf("%v", keys)))
				}

				return err
			}
		})
	} else {
		m.logger.Debug("CSRF middleware is disabled", 
			logging.Bool("config_enabled", m.config.Security.CSRF.Enabled),
			logging.String("reason", "CSRF disabled in config"))
	}

	// Auth if user service provided
	if m.config.UserService != nil {
		m.logger.Debug("setting up JWT middleware")
		middleware, err := NewJWTMiddleware(m.config.UserService, m.config.Security.JWTSecret)
		if err != nil {
			m.logger.Error("failed to create JWT middleware", logging.Error(err))
			return
		}
		e.Use(middleware)
	}

	m.logger.Debug("middleware setup complete")
}
