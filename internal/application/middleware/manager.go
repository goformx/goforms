// Package middleware provides middleware management for the application.
package middleware

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/adapters"
	contextmw "github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/security"
	"github.com/goformx/goforms/internal/application/middleware/session"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/version"
)

const (
	// DefaultRateLimit is the default requests per second
	DefaultRateLimit = 60
	// DefaultBurst is the default burst size
	DefaultBurst = 10
	// DefaultWindow is the default rate limit window
	DefaultWindow = time.Minute
)

// PathChecker handles path-based logic for middleware
type PathChecker struct {
	authPaths   []string
	formPaths   []string
	staticPaths []string
	apiPaths    []string
	healthPaths []string
}

// NewPathChecker creates a new path checker with default paths
func NewPathChecker() *PathChecker {
	return &PathChecker{
		authPaths:   []string{"/login", "/signup", "/forgot-password", "/reset-password"},
		formPaths:   []string{"/forms/new", "/forms/", "/submit"},
		staticPaths: []string{"/assets/", "/static/", "/public/", "/favicon.ico"},
		apiPaths:    []string{"/api/"},
		healthPaths: []string{"/health", "/health/", "/healthz", "/healthz/"},
	}
}

// IsAuthPath checks if the path is an authentication page
func (pc *PathChecker) IsAuthPath(path string) bool {
	return pc.containsPath(path, pc.authPaths)
}

// IsFormPath checks if the path is a form page
func (pc *PathChecker) IsFormPath(path string) bool {
	return pc.containsPath(path, pc.formPaths)
}

// IsStaticPath checks if the path is a static asset
func (pc *PathChecker) IsStaticPath(path string) bool {
	return pc.containsPath(path, pc.staticPaths)
}

// IsAPIPath checks if the path is an API route
func (pc *PathChecker) IsAPIPath(path string) bool {
	return pc.containsPath(path, pc.apiPaths)
}

// IsHealthPath checks if the path is a health check route
func (pc *PathChecker) IsHealthPath(path string) bool {
	return pc.containsPath(path, pc.healthPaths)
}

func (pc *PathChecker) containsPath(path string, paths []string) bool {
	for _, p := range paths {
		if strings.Contains(path, p) || path == p {
			return true
		}
	}
	return false
}

// Manager manages all middleware for the application
type Manager struct {
	logger            logging.Logger
	config            *ManagerConfig
	contextMiddleware *contextmw.Middleware
	pathChecker       *PathChecker
}

// ManagerConfig contains all dependencies for the middleware manager
type ManagerConfig struct {
	Logger         logging.Logger
	Config         *appconfig.Config
	UserService    user.Service
	FormService    formdomain.Service
	SessionManager *session.Manager
	AccessManager  *access.Manager
	Sanitizer      sanitization.ServiceInterface
}

// Validate ensures all required configuration is present
func (cfg *ManagerConfig) Validate() error {
	if cfg.Logger == nil {
		return errors.New("logger is required")
	}
	if cfg.Config == nil {
		return errors.New("config is required")
	}
	if cfg.Sanitizer == nil {
		return errors.New("sanitizer is required")
	}
	return nil
}

// NewManager creates a new middleware manager
func NewManager(cfg *ManagerConfig) *Manager {
	if cfg == nil {
		panic("config is required")
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}

	return &Manager{
		logger:            cfg.Logger,
		config:            cfg,
		contextMiddleware: contextmw.NewMiddleware(cfg.Logger, cfg.Config.App.RequestTimeout),
		pathChecker:       NewPathChecker(),
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
		"version", versionInfo.Version,
		"environment", m.config.Config.App.Environment)

	// Set Echo's logger to use our custom logger
	e.Logger = adapters.NewEchoLogger(m.logger)

	// Enable debug mode and set log level
	e.Debug = m.config.Config.Security.Debug
	if m.config.Config.App.IsDevelopment() {
		e.Logger.SetLevel(log.DEBUG)
		m.logger.Info("development mode enabled")
	} else {
		e.Logger.SetLevel(log.INFO)
	}

	m.setupBasicMiddleware(e)
	m.setupSecurityMiddleware(e)
	m.setupAuthMiddleware(e)

	m.logger.Info("middleware setup completed")
}

func (m *Manager) setupBasicMiddleware(e *echo.Echo) {
	// Recovery middleware first
	e.Use(Recovery(m.logger, m.config.Sanitizer))

	// Timeout middleware
	e.Use(echomw.TimeoutWithConfig(echomw.TimeoutConfig{
		Timeout:      m.config.Config.App.RequestTimeout,
		ErrorMessage: "Request timeout",
	}))

	// Context middleware
	e.Use(m.contextMiddleware.WithContext())

	// Logging middleware
	if m.config.Config.App.IsDevelopment() {
		e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
			Format:  "${time_rfc3339} ${status} ${method} ${uri} ${latency_human}\n",
			Output:  os.Stdout,
			Skipper: isNoisePath,
		}))
	} else {
		e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
			Skipper: isNoisePath,
		}))
	}
}

func (m *Manager) setupSecurityMiddleware(e *echo.Echo) {
	// CORS middleware
	if m.config.Config.Security.CORS.Enabled {
		e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
			AllowOrigins:     m.config.Config.Security.CORS.AllowedOrigins,
			AllowMethods:     m.config.Config.Security.CORS.AllowedMethods,
			AllowHeaders:     m.config.Config.Security.CORS.AllowedHeaders,
			AllowCredentials: m.config.Config.Security.CORS.AllowCredentials,
			MaxAge:           m.config.Config.Security.CORS.MaxAge,
		}))
	}

	// Secure middleware
	e.Use(echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         m.config.Config.Security.SecurityHeaders.XXSSProtection,
		ContentTypeNosniff:    m.config.Config.Security.SecurityHeaders.XContentTypeOptions,
		XFrameOptions:         m.config.Config.Security.SecurityHeaders.XFrameOptions,
		HSTSMaxAge:            constants.HSTSOneYear,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: m.config.Config.Security.GetCSPDirectives(&m.config.Config.App),
	}))

	// Set security config in context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("security_config", m.config.Config.Security)
			return next(c)
		}
	})

	// Additional security headers
	e.Use(security.SetupSecurityHeaders())

	// CSRF middleware
	if m.config.Config.Security.CSRF.Enabled {
		csrfMiddleware := security.SetupCSRF(
			&m.config.Config.Security.CSRF,
			m.config.Config.App.Environment == "development",
		)
		e.Use(csrfMiddleware)
	}

	// Rate limiting
	if m.config.Config.Security.RateLimit.Enabled {
		rateLimiter := security.NewRateLimiter(m.logger, m.config.Config, m.pathChecker)
		e.Use(rateLimiter.Setup())
	}
}

func (m *Manager) setupAuthMiddleware(e *echo.Echo) {
	if m.config.SessionManager != nil {
		e.Use(m.config.SessionManager.Middleware())
	}

	e.Use(access.Middleware(m.config.AccessManager, m.logger))
}

// isNoisePath checks if the path should be suppressed from logging
func isNoisePath(c echo.Context) bool {
	path := c.Request().URL.Path
	return strings.HasPrefix(path, "/.well-known") ||
		path == "/favicon.ico" ||
		strings.HasPrefix(path, "/robots.txt") ||
		strings.Contains(path, "com.chrome.devtools") ||
		strings.Contains(path, "devtools") ||
		strings.Contains(path, "chrome-devtools")
}

// EchoLogger is exported for backward compatibility.
//
// Deprecated: Use adapters.NewEchoLogger instead.
type EchoLogger = adapters.EchoLogger
