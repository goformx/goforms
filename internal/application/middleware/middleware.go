package middleware

import (
	"errors"
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
	// RateLimitBurst is the number of requests allowed in a burst
	RateLimitBurst = 5
	// CookieMaxAge is the maximum age of cookies in seconds (24 hours)
	CookieMaxAge = 86400
	// StaticFileFavicon is the path to the favicon
	StaticFileFavicon = "/favicon.ico"
	// StaticFileRobots is the path to the robots.txt file
	StaticFileRobots = "/robots.txt"
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
	if cfg.Security == nil {
		panic("security configuration is required for Manager")
	}
	if cfg.UserService == nil {
		panic("user service is required for Manager")
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

// retrieveCSRFToken gets the CSRF token from the context
func retrieveCSRFToken(c echo.Context) (string, error) {
	token := c.Get("csrf")
	if token == nil {
		return "", errors.New("CSRF token not found in context")
	}

	tokenStr, ok := token.(string)
	if !ok {
		return "", errors.New("CSRF token type is invalid")
	}

	if tokenStr == "" {
		return "", errors.New("CSRF token is empty")
	}

	return tokenStr, nil
}

// isStaticFile checks if the given path is a static file
func isStaticFile(path string) bool {
	return strings.HasPrefix(path, "/static/") ||
		path == StaticFileFavicon ||
		path == StaticFileRobots
}

// setupStaticFileMiddleware creates middleware to handle static files
func setupStaticFileMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if isStaticFile(c.Request().URL.Path) {
				c.Set("skip_csrf", true)
				c.Set("skip_auth", true)
			}
			return next(c)
		}
	}
}

// setupCSRF creates and configures CSRF middleware
func setupCSRF() echo.MiddlewareFunc {
	return echomw.CSRFWithConfig(echomw.CSRFConfig{
		TokenLength:    DefaultTokenLength,
		TokenLookup:    "header:X-Csrf-Token,form:csrf_token,cookie:_csrf",
		ContextKey:     "csrf",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieDomain:   "",
		CookieSecure:   true,
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

			if strings.HasPrefix(path, "/api/") {
				authHeader := c.Request().Header.Get("Authorization")
				if authHeader != "" {
					return true
				}
			}

			return false
		},
	})
}

// setupRateLimiter creates and configures rate limiter middleware
func setupRateLimiter(securityConfig *config.SecurityConfig) echo.MiddlewareFunc {
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
func setupMIMETypeMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			switch {
			case strings.HasSuffix(path, ".css"):
				c.Response().Header().Set("Content-Type", "text/css")
			case strings.HasSuffix(path, ".js"):
				c.Response().Header().Set("Content-Type", "application/javascript")
			case path == StaticFileFavicon:
				c.Response().Header().Set("Content-Type", "image/x-icon")
			case path == StaticFileRobots:
				c.Response().Header().Set("Content-Type", "text/plain")
			}
			return next(c)
		}
	}
}

// logMiddlewareRegistration logs middleware registration details
func logMiddlewareRegistration(logger logging.Logger, middlewareType string) {
	logger.Debug("registering middleware",
		logging.String("type", middlewareType),
	)
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

	// MIME type middleware (must be before other middleware)
	logMiddlewareRegistration(m.logger, "MIME type")
	e.Pre(setupMIMETypeMiddleware())

	// Static file middleware (must be before CSRF and auth)
	logMiddlewareRegistration(m.logger, "static file")
	e.Use(setupStaticFileMiddleware())

	// Basic middleware
	logMiddlewareRegistration(m.logger, "recovery")
	e.Use(echomw.Recover())

	logMiddlewareRegistration(m.logger, "request ID")
	e.Use(echomw.RequestID())

	logMiddlewareRegistration(m.logger, "secure headers")
	e.Use(echomw.Secure())

	logMiddlewareRegistration(m.logger, "body limit")
	e.Use(echomw.BodyLimit("2M"))

	// Request logging middleware
	logMiddlewareRegistration(m.logger, "request logging")
	e.Use(LoggingMiddleware(m.logger))

	// Security middleware with comprehensive configuration
	logMiddlewareRegistration(m.logger, "security headers")
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
			"connect-src 'self' *;",
		ReferrerPolicy: "strict-origin-when-cross-origin",
	}))

	// CORS for admin/dashboard routes
	logMiddlewareRegistration(m.logger, "CORS")
	e.Use(echomw.CORSWithConfig(corsConfig(
		m.config.Security.CorsAllowedOrigins,
		m.config.Security.CorsAllowedMethods,
		m.config.Security.CorsAllowedHeaders,
		m.config.Security.CorsAllowCredentials,
		m.config.Security.CorsMaxAge,
	)))

	// Form submission routes group with specific middleware
	formGroup := e.Group("/v1/forms")

	// Form-specific CORS
	logMiddlewareRegistration(m.logger, "form CORS")
	formGroup.Use(echomw.CORSWithConfig(corsConfig(
		m.config.Security.FormCorsAllowedOrigins,
		m.config.Security.FormCorsAllowedMethods,
		m.config.Security.FormCorsAllowedHeaders,
		false,
		m.config.Security.CorsMaxAge,
	)))

	// Rate limiting for form submissions
	logMiddlewareRegistration(m.logger, "rate limiter")
	formGroup.Use(setupRateLimiter(m.config.Security))

	// CSRF if enabled
	if m.config.Security.CSRF.Enabled {
		logMiddlewareRegistration(m.logger, "CSRF")
		e.Use(setupCSRF())
	}

	// Auth if user service provided
	if m.config.UserService != nil {
		logMiddlewareRegistration(m.logger, "JWT")
		middleware, err := NewJWTMiddleware(m.config.UserService, m.config.Security.JWTSecret)
		if err != nil {
			m.logger.Error("failed to create JWT middleware", logging.Error(err))
			return
		}
		e.Use(middleware)
	}

	m.logger.Info("middleware setup complete")
}

// ValidateCSRFToken validates the CSRF token in the request
func ValidateCSRFToken(c echo.Context) error {
	tokenStr, err := retrieveCSRFToken(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	// Get token from request
	reqToken := c.Request().Header.Get(echo.HeaderXCSRFToken)
	if reqToken == "" {
		reqToken = c.FormValue("_csrf")
	}
	if reqToken == "" {
		return echo.NewHTTPError(http.StatusForbidden, "CSRF token not provided in request")
	}

	// Compare tokens
	if reqToken != tokenStr {
		return echo.NewHTTPError(http.StatusForbidden, "CSRF token mismatch: provided token does not match expected value")
	}

	return nil
}
