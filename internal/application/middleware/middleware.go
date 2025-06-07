package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/time/rate"

	"github.com/goformx/goforms/internal/domain/user"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// NonceSize is the size of the nonce in bytes (32 bytes = 256 bits)
	NonceSize = 32
	// HSTSOneYear is the number of seconds in one year
	HSTSOneYear = 31536000
	// DefaultTokenLength is the default length for generated tokens
	DefaultTokenLength = 32
	// RateLimitBurst is the number of requests allowed in a burst
	RateLimitBurst = 5
	// DefaultRateLimit is the default number of requests allowed per second
	DefaultRateLimit = 20
	// CookieMaxAge is the maximum age of cookies in seconds (24 hours)
	CookieMaxAge = 86400
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger logging.Logger
	config *ManagerConfig
}

// GetSessionManager returns the session manager
func (m *Manager) GetSessionManager() *SessionManager {
	return m.config.SessionManager
}

// ManagerConfig represents the configuration for the middleware manager
type ManagerConfig struct {
	Logger         logging.Logger
	Security       *appconfig.SecurityConfig
	UserService    user.Service
	Config         *appconfig.Config
	SessionManager *SessionManager
}

// New creates a new middleware manager
func New(cfg *ManagerConfig) *Manager {
	if cfg.Logger == nil {
		panic("logger is required for Manager")
	}
	if cfg.Security == nil {
		panic("security configuration is required for Manager")
	}
	if cfg.UserService == nil {
		panic("user service is required for Manager")
	}
	if cfg.SessionManager == nil {
		panic("session manager is required for Manager")
	}

	return &Manager{
		logger: cfg.Logger,
		config: cfg,
	}
}

// Setup initializes the middleware manager with the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Info("middleware setup: starting")

	// Enable debug mode and set log level
	e.Debug = m.config.Security.Debug
	if l, ok := e.Logger.(*log.Logger); ok {
		level := log.INFO
		switch strings.ToLower(m.config.Security.LogLevel) {
		case "debug":
			level = log.DEBUG
		case "info":
			level = log.INFO
		case "warn":
			level = log.WARN
		case "error":
			level = log.ERROR
		}
		l.SetLevel(level)
		l.SetHeader("${time_rfc3339} ${level} ${prefix} ${short_file}:${line}")
		m.logger.Debug("middleware setup: echo log level set", logging.StringField("level", m.config.Security.LogLevel))
	}

	// Register basic middleware first
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     m.config.Security.CorsAllowedOrigins,
		AllowMethods:     m.config.Security.CorsAllowedMethods,
		AllowHeaders:     m.config.Security.CorsAllowedHeaders,
		AllowCredentials: m.config.Security.CorsAllowCredentials,
		MaxAge:           m.config.Security.CorsMaxAge,
	}))

	// Development mode specific setup
	if m.config.Config.App.Env == "development" {
		m.logger.Info("middleware setup: development mode enabled")
		e.Use(echomw.Logger())
	}

	// Register security middleware
	e.Use(setupSecurityHeadersMiddleware())
	e.Use(setupCSRF(m.config.Config.App.Env == "development"))
	e.Use(setupRateLimiter(m.config.Security))

	// Register session middleware last
	m.logger.Info("middleware setup: registering session middleware")
	e.Use(m.config.SessionManager.SessionMiddleware())

	m.logger.Info("middleware setup: completed")
}

// setupCSRF creates and configures CSRF middleware
func setupCSRF(isDevelopment bool) echo.MiddlewareFunc {
	return echomw.CSRFWithConfig(echomw.CSRFConfig{
		TokenLength:    DefaultTokenLength,
		TokenLookup:    "header:X-Csrf-Token,form:csrf_token,cookie:_csrf",
		ContextKey:     "csrf",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSecure:   !isDevelopment,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		CookieMaxAge:   CookieMaxAge,
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			method := c.Request().Method

			if skip, ok := c.Get("skip_csrf").(bool); ok && skip {
				return true
			}

			if isStaticFile(path) {
				return true
			}

			if method == http.MethodHead || method == http.MethodOptions {
				return true
			}

			if strings.HasPrefix(path, "/api/validation/") {
				return true
			}

			if strings.HasPrefix(path, "/api/") {
				authHeader := c.Request().Header.Get("Authorization")
				if authHeader != "" {
					return true
				}
			}

			// Don't skip CSRF for login page
			if path == "/login" {
				return false
			}

			return false
		},
	})
}

// setupRateLimiter creates and configures rate limiter middleware
func setupRateLimiter(securityConfig *appconfig.SecurityConfig) echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(securityConfig.FormRateLimit),
				Burst:     RateLimitBurst,
				ExpiresIn: securityConfig.FormRateLimitWindow,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
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
			return echo.NewHTTPError(http.StatusTooManyRequests,
				"Rate limit exceeded: too many requests from the same form or origin")
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests,
				"Rate limit exceeded: please try again later")
		},
	})
}

// setupSecurityHeadersMiddleware creates and configures security headers middleware
func setupSecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set security headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			return next(c)
		}
	}
}

// isStaticFile checks if the given path is a static file
func isStaticFile(path string) bool {
	staticExtensions := []string{
		".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".ico",
		".svg", ".woff", ".woff2", ".ttf", ".eot",
	}
	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
