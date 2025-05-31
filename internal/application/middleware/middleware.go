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

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
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
	Config      *config.Config
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
	// Skip TypeScript files in development mode
	if strings.HasSuffix(path, ".ts") {
		return false
	}

	// TODO: Use config.Static.DistDir here if it becomes dynamic at runtime
	if strings.HasPrefix(path, "/dist/") {
		return false
	}

	return strings.HasPrefix(path, "/public/") ||
		path == StaticFileFavicon ||
		path == StaticFileRobots ||
		strings.HasPrefix(path, "/@vite/") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".css")
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

			// Don't skip CSRF for login page
			if path == "/login" {
				return false
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
			case strings.HasSuffix(path, ".ts"):
				c.Response().Header().Set("Content-Type", "application/javascript")
			case strings.HasSuffix(path, ".mjs"):
				c.Response().Header().Set("Content-Type", "application/javascript")
			case path == StaticFileFavicon:
				c.Response().Header().Set("Content-Type", "image/x-icon")
			case path == StaticFileRobots:
				c.Response().Header().Set("Content-Type", "text/plain")
			case strings.HasPrefix(path, "/@vite/"):
				c.Response().Header().Set("Content-Type", "application/javascript")
			}
			return next(c)
		}
	}
}

// Helper to log and apply middleware
func (m *Manager) useWithLog(e *echo.Echo, middlewareType string, mw echo.MiddlewareFunc) {
	m.logger.Debug("middleware registered",
		logging.StringField("type", middlewareType))
	e.Use(mw)
}

// Helper to log and apply group middleware
func (m *Manager) useGroupWithLog(g *echo.Group, middlewareType string, mw echo.MiddlewareFunc) {
	m.logger.Debug("middleware registered",
		logging.StringField("type", middlewareType))
	g.Use(mw)
}

// Setup basic middleware (recovery, request ID, secure headers, body limit, logging, MIME type, static files)
func (m *Manager) setupBasicMiddleware(e *echo.Echo) {
	m.useWithLog(e, "recovery", echomw.Recover())
	m.useWithLog(e, "request ID", echomw.RequestID())
	m.useWithLog(e, "secure headers", echomw.Secure())
	m.useWithLog(e, "body limit", echomw.BodyLimit("2M"))
	m.useWithLog(e, "request logging", LoggingMiddleware(m.logger))
	m.useWithLog(e, "MIME type", setupMIMETypeMiddleware())
	m.useWithLog(e, "static file", setupStaticFileMiddleware())
}

// Setup security middleware (secure headers, CORS, CSRF, rate limiting)
func (m *Manager) setupSecurityMiddleware(e *echo.Echo) {
	m.useWithLog(e, "security headers", echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            HSTSOneYear,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: strings.Join([]string{
			"default-src 'self' http://localhost:3000; ",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' " +
				"http://localhost:3000 https://cdn.form.io https://cdn.jsdelivr.net; ",
			"worker-src 'self' blob:; ",
			"style-src 'self' 'unsafe-inline' " +
				"http://localhost:3000 https://fonts.googleapis.com " +
				"https://cdn.form.io https://cdn.jsdelivr.net; ",
			"img-src 'self' data: http://localhost:3000; ",
			"font-src 'self' http://localhost:3000 https://fonts.googleapis.com https://fonts.gstatic.com; ",
			"connect-src 'self' http://localhost:3000 ws://localhost:3000;",
		}, ""),
		ReferrerPolicy: "strict-origin-when-cross-origin",
	}))
	m.useWithLog(e, "CORS", echomw.CORSWithConfig(corsConfig(
		m.config.Security.CorsAllowedOrigins,
		m.config.Security.CorsAllowedMethods,
		m.config.Security.CorsAllowedHeaders,
		m.config.Security.CorsAllowCredentials,
		m.config.Security.CorsMaxAge,
	)))

	// Form submission routes group with specific middleware
	formGroup := e.Group("/v1/forms")
	m.useGroupWithLog(formGroup, "form CORS", echomw.CORSWithConfig(corsConfig(
		m.config.Security.FormCorsAllowedOrigins,
		m.config.Security.FormCorsAllowedMethods,
		m.config.Security.FormCorsAllowedHeaders,
		false,
		m.config.Security.CorsMaxAge,
	)))
	m.useGroupWithLog(formGroup, "rate limiter", setupRateLimiter(m.config.Security))

	if m.config.Security.CSRF.Enabled {
		m.useWithLog(e, "CSRF", setupCSRF())
	}
}

// Setup authentication middleware (cookie auth, JWT auth, protected/admin groups)
func (m *Manager) setupAuthMiddleware(e *echo.Echo) {
	if m.config.UserService != nil {
		// Create cookie auth middleware for dashboard/admin routes
		cookieAuth := NewCookieAuthMiddleware(m.config.UserService, m.logger)

		// Create JWT middleware for API routes
		jwtMiddleware, err := NewJWTMiddleware(m.config.UserService, m.config.Security.JWTSecret, m.logger, m.config.Config)
		if err != nil {
			m.logger.Error("failed to create JWT middleware", logging.ErrorField("error", err))
			return
		}

		// Create protected API routes group (with JWT middleware)
		protected := e.Group("/api/v1")
		protected.Use(jwtMiddleware)

		// Create admin/dashboard routes group (with cookie auth)
		admin := e.Group("/dashboard")
		admin.Use(cookieAuth.RequireAuth)
	}
}

// Setup initializes the middleware manager with the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Info("starting middleware setup")

	// Enable debug mode and set log level
	e.Debug = true
	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetLevel(log.DEBUG)
		l.SetHeader("${time_rfc3339} ${level} ${prefix} ${short_file}:${line}")
	}

	m.setupBasicMiddleware(e)
	m.setupSecurityMiddleware(e)
	m.setupAuthMiddleware(e)

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
