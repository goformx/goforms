package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	logger            logging.Logger
	config            *ManagerConfig
	contextMiddleware *ContextMiddleware
}

// ManagerConfig represents the configuration for the middleware manager
type ManagerConfig struct {
	Logger         logging.Logger
	Security       *appconfig.SecurityConfig
	UserService    user.Service
	Config         *appconfig.Config
	SessionManager *SessionManager
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

	return &Manager{
		logger:            cfg.Logger,
		config:            cfg,
		contextMiddleware: NewContextMiddleware(cfg.Logger),
	}
}

// GetSessionManager returns the session manager
func (m *Manager) GetSessionManager() *SessionManager {
	return m.config.SessionManager
}

// Setup registers all middleware with the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Info("middleware setup: starting")

	// Set Echo's logger to use our custom logger
	e.Logger = &EchoLogger{logger: m.logger}

	// Enable debug mode and set log level
	e.Debug = m.config.Security.Debug
	m.logger.Debug("middleware setup: echo log level set", logging.StringField("level", m.config.Security.LogLevel))

	// Add recovery middleware first to catch panics
	e.Use(Recovery(m.logger))

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
	fields := make([]any, 0, len(j))
	for k, v := range j {
		fields = append(fields, logging.StringField(k, fmt.Sprint(v)))
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
	fields := make([]any, 0, len(j))
	for k, v := range j {
		fields = append(fields, logging.StringField(k, fmt.Sprint(v)))
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
	fields := make([]any, 0, len(j))
	for k, v := range j {
		fields = append(fields, logging.StringField(k, fmt.Sprint(v)))
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
	fields := make([]any, 0, len(j))
	for k, v := range j {
		fields = append(fields, logging.StringField(k, fmt.Sprint(v)))
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
	fields := make([]any, 0, len(j))
	for k, v := range j {
		fields = append(fields, logging.StringField(k, fmt.Sprint(v)))
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
	fields := make([]any, 0, len(j))
	for k, v := range j {
		fields = append(fields, logging.StringField(k, fmt.Sprint(v)))
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
	fields := make([]any, 0, len(j))
	for k, v := range j {
		fields = append(fields, logging.StringField(k, fmt.Sprint(v)))
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
