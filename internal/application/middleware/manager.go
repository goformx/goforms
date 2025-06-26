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

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/goformx/goforms/internal/application/middleware/session"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/version"
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
	Config         *appconfig.Config // Single source of truth
	UserService    user.Service
	FormService    formdomain.Service
	SessionManager *session.Manager
	AccessManager  *access.Manager
	Sanitizer      sanitization.ServiceInterface
}

// NewManager creates a new middleware manager
func NewManager(cfg *ManagerConfig) *Manager {
	if cfg == nil {
		panic("config is required")
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
		"version", versionInfo.Version,
		"environment", m.config.Config.App.Env)

	// Set Echo's logger to use our custom logger
	e.Logger = &EchoLogger{logger: m.logger, config: m.config}

	// Enable debug mode and set log level
	e.Debug = m.config.Config.Security.Debug
	if m.config.Config.App.IsDevelopment() {
		e.Logger.SetLevel(log.DEBUG)
		m.logger.Info("development mode enabled")
	} else {
		e.Logger.SetLevel(log.INFO)
	}

	// Setup basic middleware
	m.setupBasicMiddleware(e)

	// Setup security middleware
	m.setupSecurityMiddleware(e)

	// Setup authentication middleware
	m.setupAuthMiddleware(e)

	m.logger.Info("middleware setup completed")
}

// setupBasicMiddleware sets up basic middleware like recovery, context, and logging
func (m *Manager) setupBasicMiddleware(e *echo.Echo) {
	// Add recovery middleware first to catch panics
	e.Use(Recovery(m.logger, m.config.Sanitizer))

	// Add context middleware to handle request context
	e.Use(m.contextMiddleware.WithContext())

	// Register basic middleware with custom skipper to suppress noise paths
	if m.config.Config.App.IsDevelopment() {
		// Use console format in development
		e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
			Format: "${time_rfc3339} ${status} ${method} ${uri} ${latency_human}\n",
			Output: os.Stdout,
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				return isNoisePath(path)
			},
		}))
	} else {
		// Use JSON format in production
		e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				return isNoisePath(path)
			},
		}))
	}
}

// setupSecurityMiddleware sets up security-related middleware
func (m *Manager) setupSecurityMiddleware(e *echo.Echo) {
	// Use PerFormCORS middleware for form-specific CORS handling
	// This middleware will handle CORS for form routes and fallback to global CORS for other routes
	perFormCORSConfig := NewPerFormCORSConfig(m.config.FormService, m.logger, &m.config.Config.Security)
	e.Use(PerFormCORS(perFormCORSConfig))

	// Register security middleware
	e.Use(echomw.SecureWithConfig(echomw.SecureConfig{
		XSSProtection:         m.config.Config.Security.Headers.XXSSProtection,
		ContentTypeNosniff:    m.config.Config.Security.Headers.XContentTypeOptions,
		XFrameOptions:         m.config.Config.Security.Headers.XFrameOptions,
		HSTSMaxAge:            constants.HSTSOneYear,
		HSTSExcludeSubdomains: false,
		ContentSecurityPolicy: m.config.Config.Security.GetCSPDirectives(&m.config.Config.App),
	}))

	// Set security config in context for other middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("security_config", m.config.Config.Security)
			return next(c)
		}
	})

	// Add additional security headers not covered by Echo's Secure middleware
	e.Use(setupAdditionalSecurityHeadersMiddleware())

	if m.config.Config.Security.CSRF.Enabled {
		csrfMiddleware := setupCSRF(&m.config.Config.Security.CSRF, m.config.Config.App.Env == "development")
		e.Use(csrfMiddleware)
	}

	// Setup rate limiting using infrastructure config
	if m.config.Config.Security.RateLimit.Enabled {
		e.Use(m.setupRateLimiting())
	}
}

// setupAuthMiddleware sets up authentication-related middleware
func (m *Manager) setupAuthMiddleware(e *echo.Echo) {
	// Register session middleware
	e.Use(m.config.SessionManager.Middleware())

	// Register access control middleware
	e.Use(access.Middleware(m.config.AccessManager, m.logger))
}

// setupCSRF creates and configures CSRF middleware
func setupCSRF(csrfConfig *appconfig.CSRFConfig, isDevelopment bool) echo.MiddlewareFunc {
	sameSite := getSameSite(csrfConfig.CookieSameSite, isDevelopment)
	tokenLength := getTokenLength(csrfConfig.TokenLength)

	return echomw.CSRFWithConfig(echomw.CSRFConfig{
		TokenLength:    uint8(tokenLength), // #nosec G115
		TokenLookup:    csrfConfig.TokenLookup,
		ContextKey:     csrfConfig.ContextKey,
		CookieName:     csrfConfig.CookieName,
		CookiePath:     csrfConfig.CookiePath,
		CookieDomain:   csrfConfig.CookieDomain,
		CookieSecure:   !isDevelopment, // In development, don't require HTTPS
		CookieHTTPOnly: csrfConfig.CookieHTTPOnly,
		CookieSameSite: sameSite,
		CookieMaxAge:   csrfConfig.CookieMaxAge,
		Skipper:        createCSRFSkipper(isDevelopment),
		ErrorHandler:   createCSRFErrorHandler(csrfConfig, isDevelopment),
	})
}

// getSameSite converts string SameSite to http.SameSite
func getSameSite(cookieSameSite string, isDevelopment bool) http.SameSite {
	switch cookieSameSite {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		// In development, default to Lax for cross-origin support
		if isDevelopment {
			return http.SameSiteLaxMode
		}
		return http.SameSiteStrictMode
	}
}

// getTokenLength ensures token length is within bounds for uint8
func getTokenLength(tokenLength int) int {
	if tokenLength <= 0 || tokenLength > 255 {
		return constants.DefaultTokenLength
	}
	return tokenLength
}

// createCSRFSkipper creates a function that determines if CSRF protection should be skipped
func createCSRFSkipper(isDevelopment bool) func(c echo.Context) bool {
	return func(c echo.Context) bool {
		// For GET requests, only skip CSRF if it's not a page that needs token generation
		if isSafeMethod(c.Request().Method) {
			// Allow CSRF token generation for auth pages and form pages
			if isAuthPage(c.Request().URL.Path) || isFormPage(c.Request().URL.Path) {
				return false
			}
			return true
		}

		// Skip CSRF for API routes in development
		if isDevelopment && isAPIRoute(c.Request().URL.Path) {
			return true
		}

		// Skip CSRF for health check routes
		if isHealthRoute(c.Request().URL.Path) {
			return true
		}

		// Skip CSRF for static asset routes
		if isStaticRoute(c.Request().URL.Path) {
			return true
		}

		// Skip CSRF for form submission endpoints (handled by form-specific CORS)
		if isFormSubmissionRoute(c.Request().URL.Path) {
			return true
		}

		return false
	}
}

// isSafeMethod checks if the HTTP method is safe (doesn't modify state)
func isSafeMethod(method string) bool {
	safeMethods := []string{"GET", "HEAD", "OPTIONS"}
	for _, safe := range safeMethods {
		if method == safe {
			return true
		}
	}
	return false
}

// isAPIRoute checks if the path is an API route
func isAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

// isHealthRoute checks if the path is a health check route
func isHealthRoute(path string) bool {
	healthRoutes := []string{"/health", "/health/", "/healthz", "/healthz/"}
	for _, route := range healthRoutes {
		if path == route {
			return true
		}
	}
	return false
}

// isStaticRoute checks if the path is a static asset route
func isStaticRoute(path string) bool {
	staticPrefixes := []string{"/assets/", "/static/", "/public/", "/favicon.ico"}
	for _, prefix := range staticPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// isFormSubmissionRoute checks if the path is a form submission endpoint
func isFormSubmissionRoute(path string) bool {
	// Check for form submission patterns
	submissionPatterns := []string{
		"/forms/", // Form endpoints
		"/submit", // Direct submission endpoints
	}

	for _, pattern := range submissionPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// isAuthPage checks if the path is an authentication page that needs CSRF token generation
func isAuthPage(path string) bool {
	authPages := []string{"/login", "/signup", "/forgot-password", "/reset-password"}
	for _, page := range authPages {
		if path == page {
			return true
		}
	}
	return false
}

// isFormPage checks if the path is a form page that needs CSRF token generation
func isFormPage(path string) bool {
	formPages := []string{"/forms/new", "/forms/", "/submit"}
	for _, page := range formPages {
		if strings.Contains(path, page) {
			return true
		}
	}
	return false
}

// createCSRFErrorHandler creates the CSRF error handler function
func createCSRFErrorHandler(
	csrfConfig *appconfig.CSRFConfig,
	isDevelopment bool,
) func(err error, c echo.Context) error {
	return func(err error, c echo.Context) error {
		// Add debugging in development mode
		if isDevelopment {
			// Get the actual token from the request
			csrfToken := c.Request().Header.Get("X-Csrf-Token")

			// Get the token from context (if available)
			contextToken := ""
			if token, ok := c.Get(csrfConfig.ContextKey).(string); ok {
				contextToken = token
			}

			// Get cookies for debugging
			cookies := c.Request().Header.Get("Cookie")

			c.Logger().Error("CSRF validation failed",
				"error", err.Error(),
				"path", c.Request().URL.Path,
				"method", c.Request().Method,
				"token_lookup", csrfConfig.TokenLookup,
				"origin", c.Request().Header.Get("Origin"),
				"csrf_token_present", csrfToken != "",
				"csrf_token_length", len(csrfToken),
				"csrf_token_value", csrfToken,
				"context_token_present", contextToken != "",
				"context_token_length", len(contextToken),
				"context_token_value", contextToken,
				"cookies", cookies,
				"content_type", c.Request().Header.Get("Content-Type"),
				"user_agent", c.Request().UserAgent())
		}
		return c.NoContent(http.StatusForbidden)
	}
}

// setupAdditionalSecurityHeadersMiddleware creates and configures additional security headers middleware
func setupAdditionalSecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get security config from context or use defaults
			securityConfig, ok := c.Get("security_config").(*appconfig.SecurityConfig)
			if !ok {
				// Fallback to default values if config not available
				c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
				c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			} else {
				// Use configured values
				c.Response().Header().Set("Referrer-Policy", securityConfig.Headers.ReferrerPolicy)
				c.Response().Header().Set("Strict-Transport-Security", securityConfig.Headers.StrictTransportSecurity)
				c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			}

			return next(c)
		}
	}
}

// EchoLogger implements echo.Logger interface using our custom logger
type EchoLogger struct {
	logger logging.Logger
	config *ManagerConfig
}

// Print logs a message at info level
func (l *EchoLogger) Print(i ...any) {
	l.logger.Info(fmt.Sprint(i...))
}

// Printf logs a formatted message at info level
func (l *EchoLogger) Printf(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Printj logs a JSON message at info level
func (l *EchoLogger) Printj(j log.JSON) {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Info("", fields...)
}

// Debug logs a message at debug level
func (l *EchoLogger) Debug(i ...any) {
	l.logger.Debug(fmt.Sprint(i...))
}

// Debugf logs a formatted message at debug level
func (l *EchoLogger) Debugf(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

// Debugj logs a JSON message at debug level
func (l *EchoLogger) Debugj(j log.JSON) {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Debug("", fields...)
}

// Info logs a message at info level
func (l *EchoLogger) Info(i ...any) {
	l.logger.Info(fmt.Sprint(i...))
}

// Infof logs a formatted message at info level
func (l *EchoLogger) Infof(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Infoj logs a JSON message at info level
func (l *EchoLogger) Infoj(j log.JSON) {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Info("", fields...)
}

// Warn logs a message at warn level
func (l *EchoLogger) Warn(i ...any) {
	l.logger.Warn(fmt.Sprint(i...))
}

// Warnf logs a formatted message at warn level
func (l *EchoLogger) Warnf(format string, args ...any) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

// Warnj logs a JSON message at warn level
func (l *EchoLogger) Warnj(j log.JSON) {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Warn("", fields...)
}

// Error logs a message at error level
func (l *EchoLogger) Error(i ...any) {
	l.logger.Error(fmt.Sprint(i...))
}

// Errorf logs a formatted message at error level
func (l *EchoLogger) Errorf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

// Errorj logs a JSON message at error level
func (l *EchoLogger) Errorj(j log.JSON) {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Error("", fields...)
}

// Fatal logs a message at fatal level
func (l *EchoLogger) Fatal(i ...any) {
	l.logger.Fatal(fmt.Sprint(i...))
}

// Fatalf logs a formatted message at fatal level
func (l *EchoLogger) Fatalf(format string, args ...any) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

// Fatalj logs a JSON message at fatal level
func (l *EchoLogger) Fatalj(j log.JSON) {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Fatal("", fields...)
}

// Panic logs a message at error level and panics
func (l *EchoLogger) Panic(i ...any) {
	l.logger.Error(fmt.Sprint(i...))
	panic(fmt.Sprint(i...))
}

// Panicf logs a formatted message at error level and panics
func (l *EchoLogger) Panicf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
	panic(fmt.Sprintf(format, args...))
}

// Panicj logs a JSON message at error level and panics
func (l *EchoLogger) Panicj(j log.JSON) {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	l.logger.Error("", fields...)
	panic(fmt.Sprintf("%v", j))
}

// Level returns the current log level
func (l *EchoLogger) Level() log.Lvl {
	return log.INFO
}

// SetLevel sets the log level (no-op as we use our own configuration)
func (l *EchoLogger) SetLevel(_ log.Lvl) {
	// No-op as we use our own log level configuration
}

// SetHeader sets the log header (no-op as we use our own format)
func (l *EchoLogger) SetHeader(_ string) {
	// No-op as we use our own log format
}

// SetPrefix sets the log prefix (no-op as we use our own format)
func (l *EchoLogger) SetPrefix(_ string) {
	// No-op as we use our own log format
}

// Prefix returns the current log prefix
func (l *EchoLogger) Prefix() string {
	return ""
}

// SetOutput sets the log output (no-op as we use our own configuration)
func (l *EchoLogger) SetOutput(_ io.Writer) {
	// No-op as we use our own output configuration
}

// Output returns the current log output writer
func (l *EchoLogger) Output() io.Writer {
	return os.Stdout
}

// setupRateLimiting creates and configures rate limiting middleware using infrastructure config
func (m *Manager) setupRateLimiting() echo.MiddlewareFunc {
	rateLimitConfig := m.config.Config.Security.RateLimit

	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			method := c.Request().Method

			// Skip paths from config
			for _, skipPath := range rateLimitConfig.SkipPaths {
				if strings.HasPrefix(path, skipPath) {
					return true
				}
			}

			// Skip methods from config
			for _, skipMethod := range rateLimitConfig.SkipMethods {
				if method == skipMethod {
					return true
				}
			}

			return false
		},
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(rateLimitConfig.Requests),
				Burst:     rateLimitConfig.Burst,
				ExpiresIn: rateLimitConfig.Window,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// For login and signup pages, use IP address as identifier
			path := c.Request().URL.Path
			if path == constants.PathLogin || path == constants.PathSignup || path == constants.PathResetPassword {
				return c.RealIP(), nil
			}

			// For form submissions, use form ID and origin
			formID := c.Param("formID")
			origin := c.Request().Header.Get("Origin")
			if formID == "" {
				formID = constants.DefaultUnknown
			}
			if origin == "" {
				origin = constants.DefaultUnknown
			}
			return fmt.Sprintf("%s:%s", formID, origin), nil
		},
		ErrorHandler: func(_ echo.Context, _ error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests,
				"Rate limit exceeded: too many requests from the same form or origin")
		},
		DenyHandler: func(_ echo.Context, _ string, _ error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests,
				"Rate limit exceeded: please try again later")
		},
	})
}

// isNoisePath checks if the path should be suppressed from logging
func isNoisePath(path string) bool {
	const faviconPath = "/favicon.ico"
	return strings.HasPrefix(path, "/.well-known") ||
		path == faviconPath ||
		strings.HasPrefix(path, "/robots.txt") ||
		strings.Contains(path, "com.chrome.devtools") ||
		strings.Contains(path, "devtools") ||
		strings.Contains(path, "chrome-devtools")
}
