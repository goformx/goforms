package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/time/rate"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/user"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/version"
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
	// FormCorsMaxAge is the maximum age for form-specific CORS settings
	FormCorsMaxAge = 3600
	// FieldPairSize represents the number of elements in a key-value pair
	FieldPairSize = 2
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger            logging.Logger
	config            *ManagerConfig
	contextMiddleware *context.Middleware
}

// ManagerConfig represents the configuration for the middleware manager
type ManagerConfig struct {
	Logger         logging.Logger
	Security       *appconfig.SecurityConfig
	UserService    user.Service
	Config         *appconfig.Config
	SessionManager *session.Manager
	AccessManager  *access.AccessManager
	Sanitizer      sanitization.ServiceInterface
}

// NewManager creates a new middleware manager
func NewManager(cfg *ManagerConfig) *Manager {
	if cfg == nil {
		panic("config is required")
	}

	if cfg.Security == nil {
		panic("security config is required")
	}

	if cfg.UserService == nil {
		panic("user service is required")
	}

	if cfg.SessionManager == nil {
		panic("session manager is required")
	}

	if cfg.Sanitizer == nil {
		panic("sanitizer is required")
	}

	return &Manager{
		logger:            cfg.Logger,
		config:            cfg,
		contextMiddleware: context.NewMiddleware(cfg.Logger, cfg.Config.App.RequestTimeout),
	}
}

// GetSessionManager returns the session manager
func (m *Manager) GetSessionManager() *session.Manager {
	return m.config.SessionManager
}

// Setup registers all middleware with the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	versionInfo := version.GetInfo()
	m.logger.Info("setting up middleware",
		"app", "goforms",
		"version", versionInfo.Version,
		"environment", m.config.Config.App.Env,
		"build_time", versionInfo.BuildTime,
		"git_commit", versionInfo.GitCommit,
	)

	// Set Echo's logger to use our custom logger
	e.Logger = &EchoLogger{logger: m.logger}

	// Enable debug mode and set log level
	e.Debug = m.config.Security.Debug
	if m.config.Config.App.IsDevelopment() {
		e.Logger.SetLevel(log.DEBUG)
		m.logger.Info("development mode enabled",
			"app", "goforms",
			"version", versionInfo.Version,
			"environment", m.config.Config.App.Env,
			"build_time", versionInfo.BuildTime,
			"git_commit", versionInfo.GitCommit)
	} else {
		e.Logger.SetLevel(log.INFO)
	}

	// Add recovery middleware first to catch panics
	e.Use(Recovery(m.logger, m.config.Sanitizer))

	// Add context middleware to handle request context
	e.Use(m.contextMiddleware.WithContext())

	// Register basic middleware
	if m.config.Config.App.IsDevelopment() {
		// Use console format in development
		e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
			Format: "${time_rfc3339} ${status} ${method} ${uri} ${latency_human}\n",
			Output: os.Stdout,
		}))
	} else {
		// Use JSON format in production
		e.Use(echomw.Logger())
	}
	e.Use(echomw.Recover())
	e.Use(setupCORS(m.config.Security))

	// Register security middleware
	e.Use(setupSecurityHeadersMiddleware())
	e.Use(setupCSRF(m.config.Config.App.Env == "development"))
	e.Use(setupRateLimiter(m.config.Security))

	// Register session middleware
	m.logger.Info("registering session middleware",
		"app", "goforms",
		"version", versionInfo.Version,
		"environment", m.config.Config.App.Env,
		"build_time", versionInfo.BuildTime,
		"git_commit", versionInfo.GitCommit)
	e.Use(m.config.SessionManager.Middleware())

	// Register access control middleware
	m.logger.Info("registering access control middleware",
		"app", "goforms",
		"version", versionInfo.Version,
		"environment", m.config.Config.App.Env,
		"build_time", versionInfo.BuildTime,
		"git_commit", versionInfo.GitCommit)
	e.Use(access.Middleware(m.config.AccessManager, m.logger))

	m.logger.Info("middleware setup completed",
		"app", "goforms",
		"version", versionInfo.Version,
		"environment", m.config.Config.App.Env,
		"build_time", versionInfo.BuildTime,
		"git_commit", versionInfo.GitCommit)
}

// setupCSRF creates and configures CSRF middleware
func setupCSRF(isDevelopment bool) echo.MiddlewareFunc {
	return echomw.CSRFWithConfig(echomw.CSRFConfig{
		TokenLength:    DefaultTokenLength,
		TokenLookup:    "header:X-CSRF-Token,form:_csrf,cookie:_csrf",
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

			// Skip CSRF for static files
			if isStaticFile(path) {
				return true
			}

			// Skip CSRF validation for safe methods, but still generate token
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				return false
			}

			// Skip CSRF for API endpoints with valid Authorization header or CSRF token
			if strings.HasPrefix(path, "/api/") {
				authHeader := c.Request().Header.Get("Authorization")
				csrfToken := c.Request().Header.Get("X-Csrf-Token")
				if authHeader != "" || csrfToken != "" {
					return true
				}
			}

			// Skip CSRF for validation endpoints
			if strings.HasPrefix(path, "/api/validation/") {
				return true
			}

			// Never skip CSRF for login, signup, or password reset
			if path == "/login" || path == "/signup" || path == "/reset-password" {
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
			// For login and signup pages, use IP address as identifier
			path := c.Request().URL.Path
			if path == "/login" || path == "/signup" {
				return c.RealIP(), nil
			}

			// For form submissions, use form ID and origin
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

// EchoLogger implements echo.Logger interface using our custom logger
type EchoLogger struct {
	logger logging.Logger
}

func (l *EchoLogger) Print(i ...any) {
	l.logger.Info(fmt.Sprint(i...))
}

func (l *EchoLogger) Printf(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *EchoLogger) Printj(j log.JSON) {
	fields := make([]any, 0, len(j)*FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Info("", fields...)
}

func (l *EchoLogger) Debug(i ...any) {
	l.logger.Debug(fmt.Sprint(i...))
}

func (l *EchoLogger) Debugf(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l *EchoLogger) Debugj(j log.JSON) {
	fields := make([]any, 0, len(j)*FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Debug("", fields...)
}

func (l *EchoLogger) Info(i ...any) {
	l.logger.Info(fmt.Sprint(i...))
}

func (l *EchoLogger) Infof(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *EchoLogger) Infoj(j log.JSON) {
	fields := make([]any, 0, len(j)*FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Info("", fields...)
}

func (l *EchoLogger) Warn(i ...any) {
	l.logger.Warn(fmt.Sprint(i...))
}

func (l *EchoLogger) Warnf(format string, args ...any) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *EchoLogger) Warnj(j log.JSON) {
	fields := make([]any, 0, len(j)*FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Warn("", fields...)
}

func (l *EchoLogger) Error(i ...any) {
	l.logger.Error(fmt.Sprint(i...))
}

func (l *EchoLogger) Errorf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *EchoLogger) Errorj(j log.JSON) {
	fields := make([]any, 0, len(j)*FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Error("", fields...)
}

func (l *EchoLogger) Fatal(i ...any) {
	l.logger.Fatal(fmt.Sprint(i...))
}

func (l *EchoLogger) Fatalf(format string, args ...any) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

func (l *EchoLogger) Fatalj(j log.JSON) {
	fields := make([]any, 0, len(j)*FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Fatal("", fields...)
}

func (l *EchoLogger) Panic(i ...any) {
	l.logger.Error(fmt.Sprint(i...))
	panic(fmt.Sprint(i...))
}

func (l *EchoLogger) Panicf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
	panic(fmt.Sprintf(format, args...))
}

func (l *EchoLogger) Panicj(j log.JSON) {
	fields := make([]any, 0, len(j)*FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Error("", fields...)
	panic(fmt.Sprintf("%v", j))
}

func (l *EchoLogger) Level() log.Lvl {
	return log.INFO
}

func (l *EchoLogger) SetLevel(level log.Lvl) {
	// No-op as we use our own log level configuration
}

func (l *EchoLogger) SetHeader(h string) {
	// No-op as we use our own log format
}

func (l *EchoLogger) SetPrefix(p string) {
	// No-op as we use our own log format
}

func (l *EchoLogger) Prefix() string {
	return ""
}

func (l *EchoLogger) SetOutput(w io.Writer) {
	// No-op as we use our own output configuration
}

func (l *EchoLogger) Output() io.Writer {
	return os.Stdout
}

func setupCORS(securityConfig *appconfig.SecurityConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			method := c.Request().Method

			// Check if this is a form-related endpoint
			isFormEndpoint := strings.HasPrefix(path, "/api/v1/forms/") &&
				(strings.HasSuffix(path, "/schema") || strings.HasSuffix(path, "/submit"))

			corsConfig := getCORSConfig(securityConfig, isFormEndpoint)

			// Handle preflight requests
			if method == "OPTIONS" {
				return handlePreflightRequest(c, corsConfig)
			}

			// Handle actual requests
			return handleActualRequest(c, corsConfig, next)
		}
	}
}

// getCORSConfig returns the appropriate CORS configuration based on endpoint type
func getCORSConfig(securityConfig *appconfig.SecurityConfig, isFormEndpoint bool) *corsConfig {
	if isFormEndpoint {
		// Use form-specific CORS settings for form endpoints
		return &corsConfig{
			allowedOrigins:   securityConfig.FormCorsAllowedOrigins,
			allowedMethods:   securityConfig.FormCorsAllowedMethods,
			allowedHeaders:   securityConfig.FormCorsAllowedHeaders,
			allowCredentials: false, // Forms don't need credentials
			maxAge:           FormCorsMaxAge,
		}
	}

	// Use general CORS settings for other endpoints
	return &corsConfig{
		allowedOrigins:   securityConfig.CorsAllowedOrigins,
		allowedMethods:   securityConfig.CorsAllowedMethods,
		allowedHeaders:   securityConfig.CorsAllowedHeaders,
		allowCredentials: securityConfig.CorsAllowCredentials,
		maxAge:           securityConfig.CorsMaxAge,
	}
}

// corsConfig holds CORS configuration
type corsConfig struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	allowCredentials bool
	maxAge           int
}

// handlePreflightRequest handles OPTIONS requests
func handlePreflightRequest(c echo.Context, config *corsConfig) error {
	origin := c.Request().Header.Get("Origin")

	if isOriginAllowed(origin, config.allowedOrigins) {
		c.Response().Header().Set("Access-Control-Allow-Origin", origin)
		c.Response().Header().Set("Access-Control-Allow-Methods", strings.Join(config.allowedMethods, ","))
		c.Response().Header().Set("Access-Control-Allow-Headers", strings.Join(config.allowedHeaders, ","))
		if config.allowCredentials {
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		}
		c.Response().Header().Set("Access-Control-Max-Age", strconv.Itoa(config.maxAge))
		return c.NoContent(http.StatusNoContent)
	}

	return c.NoContent(http.StatusNoContent)
}

// handleActualRequest handles actual requests (non-OPTIONS)
func handleActualRequest(c echo.Context, config *corsConfig, next echo.HandlerFunc) error {
	origin := c.Request().Header.Get("Origin")
	if origin != "" && isOriginAllowed(origin, config.allowedOrigins) {
		c.Response().Header().Set("Access-Control-Allow-Origin", origin)
		if config.allowCredentials {
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		}
	}

	return next(c)
}

// isOriginAllowed checks if the origin is in the allowed origins list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}
