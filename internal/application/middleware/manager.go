// Package middleware provides middleware management for the application.
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/adapters"
	"github.com/goformx/goforms/internal/application/middleware/assertion"
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
	// HTTPStatusServerError is the threshold for server errors (5xx)
	HTTPStatusServerError = 500
	// HTTPStatusClientError is the threshold for client errors (4xx)
	HTTPStatusClientError = 400
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

	// Timeout middleware (using context-based timeout to avoid data races)
	e.Use(echomw.ContextTimeoutWithConfig(echomw.ContextTimeoutConfig{
		Timeout: m.config.Config.App.RequestTimeout,
	}))

	// Context middleware
	e.Use(m.contextMiddleware.WithContext())

	// Logging middleware (using RequestLoggerWithConfig for race-free logging)
	e.Use(echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		Skipper:     isNoisePath,
		LogValuesFunc: func(c echo.Context, v echomw.RequestLoggerValues) error {
			// Get request ID from header
			requestID := c.Request().Header.Get("X-Trace-Id")
			logger := m.logger
			if requestID != "" {
				logger = logger.WithRequestID(requestID)
			}

			// Build base fields
			fields := []any{
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency_ms", v.Latency.Milliseconds(),
				"remote_ip", c.RealIP(),
			}

			// Add user_id if authenticated
			if userID, ok := contextmw.GetUserID(c); ok {
				fields = append(fields, "user_id", userID)
			}

			// Add form_id if this is a form route
			if formID := c.Param("id"); formID != "" && isFormRoute(v.URI) {
				fields = append(fields, "form_id", formID)
			}

			// Add assertion failure reason when 401 from assertion middleware
			if r, ok := c.Get(assertion.FailureReasonContextKey).(string); ok && r != "" {
				fields = append(fields, "assertion_reason", r)
			}

			// Log based on status and error
			if v.Error != nil {
				fields = append(fields, "error", v.Error.Error())
				logger.Error("request failed", fields...)
			} else if v.Status >= HTTPStatusServerError {
				logger.Error("request completed with server error", fields...)
			} else if v.Status >= HTTPStatusClientError {
				logger.Warn("request completed with client error", fields...)
			} else {
				logger.Info("request completed", fields...)
			}

			return nil
		},
	}))

	// Slow request detection middleware
	e.Use(SlowRequestDetectorWithConfig(m.logger, SlowRequestConfig{
		Threshold:         DefaultSlowRequestThreshold,
		VerySlowThreshold: VerySlowRequestThreshold,
		Skipper:           NewSlowRequestSkipper(),
	}))
}

func (m *Manager) setupSecurityMiddleware(e *echo.Echo) {
	// CORS middleware
	if m.config.Config.Security.CORS.Enabled {
		corsConfig := echomw.CORSConfig{
			AllowOrigins:     m.config.Config.Security.CORS.AllowedOrigins,
			AllowMethods:     m.config.Config.Security.CORS.AllowedMethods,
			AllowHeaders:     m.config.Config.Security.CORS.AllowedHeaders,
			AllowCredentials: m.config.Config.Security.CORS.AllowCredentials,
			MaxAge:           m.config.Config.Security.CORS.MaxAge,
			Skipper:          shouldSkipGlobalCORS,
		}
		e.Use(echomw.CORSWithConfig(corsConfig))
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
	m.logger.Info("CSRF middleware configuration",
		"enabled", m.config.Config.Security.CSRF.Enabled,
		"context_key", m.config.Config.Security.CSRF.ContextKey,
		"cookie_name", m.config.Config.Security.CSRF.CookieName,
		"token_lookup", m.config.Config.Security.CSRF.TokenLookup)
	if m.config.Config.Security.CSRF.Enabled {
		csrfMiddleware := security.SetupCSRF(
			&m.config.Config.Security.CSRF,
			m.config.Config.App.Environment == "development",
			m.logger,
		)
		e.Use(csrfMiddleware)
		m.logger.Info("CSRF middleware registered",
			"context_key", m.config.Config.Security.CSRF.ContextKey,
			"cookie_name", m.config.Config.Security.CSRF.CookieName)
	} else {
		m.logger.Warn("CSRF middleware is DISABLED")
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

// isFormRoute checks if the path is a form-related route
func isFormRoute(path string) bool {
	return strings.HasPrefix(path, "/forms/") ||
		strings.HasPrefix(path, "/api/v1/forms/") ||
		strings.HasPrefix(path, "/submit/")
}

func shouldSkipGlobalCORS(c echo.Context) bool {
	return isPublicFormCORSPath(c.Request().Method, c.Request().URL.Path)
}

func isPublicFormCORSPath(method, requestPath string) bool {
	apiPrefix := constants.PathAPIForms + "/"
	publicPrefix := constants.PathFormsPublic + "/"
	if !strings.HasPrefix(requestPath, apiPrefix) && !strings.HasPrefix(requestPath, publicPrefix) {
		return false
	}

	switch {
	case strings.HasSuffix(requestPath, "/schema"):
		return method == http.MethodGet || method == http.MethodOptions
	case strings.HasSuffix(requestPath, "/validation"):
		return method == http.MethodGet || method == http.MethodOptions
	case strings.HasSuffix(requestPath, "/submit"):
		return method == http.MethodPost || method == http.MethodOptions
	case strings.HasSuffix(requestPath, "/embed"):
		return method == http.MethodGet || method == http.MethodOptions
	default:
		return false
	}
}
