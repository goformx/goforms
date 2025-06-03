package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

// Config represents the middleware configuration
type Config struct {
	Security    *appconfig.SecurityConfig
	UserService user.Service
	Config      *appconfig.Config
}

// SecurityConfig represents the security configuration
type SecurityConfig struct {
	Debug                  bool
	LogLevel               string
	CSRF                   appconfig.CSRFConfig
	CorsAllowedOrigins     appconfig.CORSOriginsDecoder
	CorsAllowedMethods     appconfig.CORSMethodsDecoder
	CorsAllowedHeaders     appconfig.CORSHeadersDecoder
	CorsMaxAge             int
	CorsAllowCredentials   bool
	RequestTimeout         time.Duration
	FormCorsAllowedOrigins appconfig.CORSOriginsDecoder
	FormCorsAllowedMethods appconfig.CORSMethodsDecoder
	FormCorsAllowedHeaders appconfig.CORSHeadersDecoder
	FormRateLimit          int
	FormRateLimitWindow    time.Duration
}

// Manager handles middleware configuration and setup
type Manager struct {
	logger logging.Logger
	config *ManagerConfig
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

// corsConfig creates a CORS configuration with the given parameters
func corsConfig(
	allowedOrigins,
	allowedMethods,
	allowedHeaders []string,
	allowCredentials bool,
	maxAge int,
) echomw.CORSConfig {
	return echomw.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     allowedMethods,
		AllowHeaders:     allowedHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
	}
}

// isStaticFile checks if the given path is a static file
func isStaticFile(path string) bool {
	// System files that should always be considered static
	if strings.HasPrefix(path, "/.well-known/") ||
		path == "/favicon.ico" ||
		path == "/robots.txt" {
		return true
	}

	// Application static files
	return strings.HasPrefix(path, "/public/") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".jpeg") ||
		strings.HasSuffix(path, ".gif") ||
		strings.HasSuffix(path, ".svg") ||
		strings.HasSuffix(path, ".woff") ||
		strings.HasSuffix(path, ".woff2") ||
		strings.HasSuffix(path, ".ttf") ||
		strings.HasSuffix(path, ".eot")
}

// setupStaticFileMiddleware creates middleware to handle static files
func setupStaticFileMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if isStaticFile(path) {
				c.Set("skip_csrf", true)
				c.Set("skip_auth", true)
			}
			return next(c)
		}
	}
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

// setupMIMETypeMiddleware creates middleware to set appropriate Content-Type headers
// This is now handled by StaticHandler and should be removed in future versions
func setupMIMETypeMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}

// Setup initializes the middleware manager with the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Info("starting middleware setup")

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
	}

	// Setup middleware in correct order
	m.setupBasicMiddleware(e)

	// Setup session middleware globally (after basic, before security)
	m.setupSessionMiddleware(e)

	m.setupSecurityMiddleware(e)

	// Add security headers middleware
	m.setupSecurityHeadersMiddleware(e)

	m.logger.Info("middleware setup complete")
}

// Setup basic middleware (recovery, request ID, secure headers, body limit, logging, MIME type, static files)
func (m *Manager) setupBasicMiddleware(e *echo.Echo) {
	m.logger.Debug("registering middleware", logging.StringField("type", "recovery"))
	e.Use(echomw.Recover())

	m.logger.Debug("registering middleware", logging.StringField("type", "request ID"))
	e.Use(echomw.RequestID())

	m.logger.Debug("registering middleware", logging.StringField("type", "secure headers"))
	e.Use(echomw.Secure())

	m.logger.Debug("registering middleware", logging.StringField("type", "body limit"))
	e.Use(echomw.BodyLimit("2M"))

	m.logger.Debug("registering middleware", logging.StringField("type", "MIME type"))
	e.Use(setupMIMETypeMiddleware())

	m.logger.Debug("registering middleware", logging.StringField("type", "static file"))
	e.Use(setupStaticFileMiddleware())
}

// Setup session middleware adds global session middleware
func (m *Manager) setupSessionMiddleware(e *echo.Echo) {
	m.logger.Debug("registering middleware", logging.StringField("type", "session"))
	e.Use(m.config.SessionManager.SessionMiddleware())
}

// Setup security middleware (secure headers, CORS, CSRF, rate limiting)
func (m *Manager) setupSecurityMiddleware(e *echo.Echo) {
	m.logger.Debug("registering middleware", logging.StringField("type", "security headers"))
	e.Use(echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            HSTSOneYear,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: strings.Join([]string{
			"default-src 'self' http://localhost:3000; ",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' http://localhost:3000 https://cdn.form.io; ",
			"script-src-elem 'self' 'unsafe-inline' 'unsafe-eval' http://localhost:3000 https://cdn.form.io; ",
			"worker-src 'self' blob:; ",
			"child-src 'self' blob:; ",
			"style-src 'self' 'unsafe-inline' http://localhost:3000; ",
			"style-src-elem 'self' 'unsafe-inline' http://localhost:3000; ",
			"img-src 'self' data: http://localhost:3000; ",
			"font-src 'self' http://localhost:3000; ",
			"connect-src 'self' http://localhost:3000 ws://localhost:3000;",
		}, ""),
		ReferrerPolicy: "strict-origin-when-cross-origin",
	}))

	m.logger.Debug("registering middleware", logging.StringField("type", "CORS"))
	e.Use(echomw.CORSWithConfig(corsConfig(
		m.config.Security.CorsAllowedOrigins,
		m.config.Security.CorsAllowedMethods,
		m.config.Security.CorsAllowedHeaders,
		m.config.Security.CorsAllowCredentials,
		m.config.Security.CorsMaxAge,
	)))

	// Form submission routes group with specific middleware
	formGroup := e.Group("/v1/forms")
	m.logger.Debug("registering middleware", logging.StringField("type", "form CORS"))
	formGroup.Use(echomw.CORSWithConfig(corsConfig(
		m.config.Security.FormCorsAllowedOrigins,
		m.config.Security.FormCorsAllowedMethods,
		m.config.Security.FormCorsAllowedHeaders,
		false,
		m.config.Security.CorsMaxAge,
	)))

	m.logger.Debug("registering middleware", logging.StringField("type", "rate limiter"))
	formGroup.Use(setupRateLimiter(m.config.Security))

	if m.config.Security.CSRF.Enabled {
		m.logger.Debug("registering middleware", logging.StringField("type", "CSRF"))
		e.Use(setupCSRF(m.config.Security.Debug))
	}
}

// Setup security headers middleware (adds extra security headers to all responses)
func (m *Manager) setupSecurityHeadersMiddleware(e *echo.Echo) {
	m.logger.Info("setting up security headers middleware")
	m.logger.Debug("registering middleware", logging.StringField("type", "security_headers"))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add security headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			return next(c)
		}
	})
}
