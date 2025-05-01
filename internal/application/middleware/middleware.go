package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
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
	m.logger.Info("starting middleware setup")

	// Enable debug mode and set log level
	e.Debug = true
	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetLevel(log.DEBUG)
		l.SetHeader("${time_rfc3339} ${level} ${prefix} ${short_file}:${line}")
	}

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
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
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

	m.logger.Info("CSRF middleware enabled", logging.Bool("enabled", m.config.Security.CSRF.Enabled))

	// CSRF if enabled
	if m.config.Security.CSRF.Enabled {
		m.logger.Info("initializing CSRF middleware",
			logging.Bool("config_enabled", m.config.Security.CSRF.Enabled),
			logging.String("secret_length", fmt.Sprintf("%d", len(m.config.Security.CSRF.Secret))),
			logging.String("token_lookup", "header:X-CSRF-Token,form:csrf_token,cookie:_csrf"),
			logging.String("cookie_name", "_csrf"),
			logging.String("cookie_path", "/"),
			logging.Bool("cookie_secure", true),
			logging.Bool("cookie_http_only", true),
			logging.String("cookie_same_site", "Strict"),
			logging.Int("cookie_max_age", 86400))

		// Create CSRF middleware with logging
		csrfMiddleware := echomw.CSRFWithConfig(echomw.CSRFConfig{
			TokenLength:    DefaultTokenLength,
			TokenLookup:    "header:X-CSRF-Token,form:csrf_token,cookie:_csrf",
			ContextKey:     "csrf",
			CookieName:     "_csrf",
			CookiePath:     "/",
			CookieDomain:   "",
			CookieSecure:   true,
			CookieHTTPOnly: true,
			CookieSameSite: http.SameSiteStrictMode,
			CookieMaxAge:   86400,
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				method := c.Request().Method

				// Check if CSRF should be skipped
				if skip, ok := c.Get("skip_csrf").(bool); ok && skip {
					m.logger.Debug("CSRF skipped: skip_csrf flag set",
						logging.String("path", path),
						logging.String("reason", "skip_csrf flag"))
					return true
				}

				// Skip for static content
				if strings.HasPrefix(path, "/static/") ||
					path == "/favicon.ico" ||
					path == "/robots.txt" {
					m.logger.Debug("CSRF skipped: static content",
						logging.String("path", path),
						logging.String("reason", "static content path"))
					return true
				}

				// Skip for safe HTTP methods
				if method == http.MethodHead || method == http.MethodOptions {
					m.logger.Debug("CSRF skipped: safe HTTP method",
						logging.String("path", path),
						logging.String("method", method),
						logging.String("reason", "safe HTTP method"))
					return true
				}

				// Skip for authenticated API routes
				if strings.HasPrefix(path, "/api/") {
					authHeader := c.Request().Header.Get("Authorization")
					if authHeader != "" {
						m.logger.Debug("CSRF skipped: authenticated API route",
							logging.String("path", path),
							logging.String("reason", "authenticated API route"))
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
						logging.String("method", method),
						logging.String("reason", "page contains form"))
					return false
				}

				// Generate tokens for all other pages by default
				m.logger.Debug("CSRF not skipped: default case",
					logging.String("path", path),
					logging.String("method", method),
					logging.String("reason", "default case - generating token"))
				return false
			},
		})

		m.logger.Debug("CSRF middleware created, adding to Echo instance")
		e.Use(csrfMiddleware)

		// Add logging middleware after CSRF
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Skip logging for static files
				path := c.Request().URL.Path
				if strings.HasPrefix(path, "/static/") ||
					path == "/favicon.ico" ||
					path == "/robots.txt" {
					return next(c)
				}

				headers := c.Request().Header
				// Log request details
				m.logger.Debug("CSRF middleware processing request",
					logging.String("path", path),
					logging.String("method", c.Request().Method),
					logging.String("content_type", headers.Get("Content-Type")),
					logging.String("user_agent", headers.Get("User-Agent")),
					logging.String("referer", headers.Get("Referer")),
					logging.String("origin", headers.Get("Origin")),
					logging.String("x_csrf_token", headers.Get("X-CSRF-Token")),
					logging.String("x_xsrf_token", headers.Get("X-XSRF-TOKEN")),
					logging.String("form_csrf_token", c.FormValue("csrf_token")))

				if cookie, err := c.Cookie("_csrf"); err == nil {
					m.logger.Debug("CSRF middleware processing request - cookie value",
						logging.String("cookie_value", cookie.Value))
				}

				// Get the token from the context before calling next middleware
				if token := c.Get("csrf"); token != nil {
					if tokenStr, ok := token.(string); ok && tokenStr != "" {
						// Set the token in the context for templates using the same key
						c.Set("csrf", tokenStr)
						m.logger.Debug("CSRF token set in context",
							logging.String("path", path),
							logging.String("method", c.Request().Method),
							logging.String("token_prefix", tokenStr[:8]),
							logging.String("token_length", fmt.Sprintf("%d", len(tokenStr))))
					}
				}

				// Call next middleware
				return next(c)
			}
		})

		// Add logging middleware before CSRF to track token generation
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Skip logging for static files
				path := c.Request().URL.Path
				if strings.HasPrefix(path, "/static/") ||
					path == "/favicon.ico" ||
					path == "/robots.txt" {
					return next(c)
				}

				headers := c.Request().Header
				// Log request details before CSRF middleware
				m.logger.Debug("Before CSRF middleware",
					logging.String("path", path),
					logging.String("method", c.Request().Method),
					logging.String("content_type", headers.Get("Content-Type")),
					logging.String("user_agent", headers.Get("User-Agent")),
					logging.String("referer", headers.Get("Referer")),
					logging.String("origin", headers.Get("Origin")),
					logging.String("x_csrf_token", headers.Get("X-CSRF-Token")),
					logging.String("x_xsrf_token", headers.Get("X-XSRF-TOKEN")),
					logging.String("form_csrf_token", c.FormValue("csrf_token")))

				if cookie, err := c.Cookie("_csrf"); err == nil {
					m.logger.Debug("Before CSRF middleware - cookie value",
						logging.String("cookie_value", cookie.Value))
				}

				// Call next middleware
				err := next(c)

				// Get the token from the context after CSRF middleware
				if token := c.Get("csrf"); token != nil {
					if tokenStr, ok := token.(string); ok && tokenStr != "" {
						m.logger.Debug("CSRF token generated",
							logging.String("path", path),
							logging.String("method", c.Request().Method),
							logging.String("token_prefix", tokenStr[:8]),
							logging.String("token_length", fmt.Sprintf("%d", len(tokenStr))))
					}
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

// ValidateCSRFToken validates the CSRF token in the request
func ValidateCSRFToken(c echo.Context) error {
	token := c.Get("csrf")
	if token == nil {
		return echo.NewHTTPError(http.StatusForbidden, "CSRF token not found")
	}

	// Get token from request
	reqToken := c.Request().Header.Get(echo.HeaderXCSRFToken)
	if reqToken == "" {
		reqToken = c.FormValue("_csrf")
	}
	if reqToken == "" {
		return echo.NewHTTPError(http.StatusForbidden, "CSRF token not provided")
	}

	// Compare tokens
	if reqToken != token.(string) {
		return echo.NewHTTPError(http.StatusForbidden, "CSRF token mismatch")
	}

	return nil
}
