package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

const (
	// NonceSize is the size of the nonce in bytes (32 bytes = 256 bits)
	NonceSize = 32
	// HSTSOneYear is the number of seconds in one year
	HSTSOneYear = 31536000
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger logging.Logger
	config *ManagerConfig
}

// ManagerConfig holds middleware configuration
type ManagerConfig struct {
	Logger      logging.Logger
	JWTSecret   string
	UserService any
	EnableCSRF  bool
}

// New creates a new middleware manager
func New(config *ManagerConfig) *Manager {
	if config.Logger == nil {
		panic("logger is required for Manager")
	}

	return &Manager{
		logger: config.Logger,
		config: config,
	}
}

// Setup configures all middleware for an Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Debug("setting up middleware")

	// Basic middleware
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(echomw.Secure())
	e.Use(echomw.BodyLimit("2M"))

	// Security middleware with comprehensive configuration
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
			"connect-src 'self'",
		ReferrerPolicy: "strict-origin-when-cross-origin",
	}))

	// CORS
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-CSRF-Token",
		},
	}))

	// CSRF if enabled
	if m.config != nil && m.config.EnableCSRF {
		e.Use(CSRF())
	}

	m.logger.Debug("middleware setup complete")
}

// CSRFConfig holds configuration for CSRF middleware
type CSRFConfig struct {
	SecretKey string
	Secure    bool
}

// CSRF returns CSRF middleware with the given configuration
func (m *Manager) CSRF(config CSRFConfig) echo.MiddlewareFunc {
	m.logger.Debug("creating CSRF middleware",
		logging.Bool("secure", config.Secure),
		logging.String("cookie_name", "csrf_token"),
	)

	csrfMiddleware := csrf.Protect(
		[]byte(config.SecretKey),
		csrf.Secure(config.Secure),
		csrf.Path("/"),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.logger.Error("CSRF validation failed",
				logging.String("path", r.URL.Path),
				logging.String("method", r.Method),
			)
			http.Error(w, "CSRF validation failed", http.StatusForbidden)
		})),
	)

	return echo.WrapMiddleware(func(next http.Handler) http.Handler {
		return csrfMiddleware(next)
	})
}

// CSRFToken returns middleware to add CSRF token to templates
func (m *Manager) CSRFToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := csrf.Token(c.Request())
			c.Set("csrf", token)
			return next(c)
		}
	}
}
