package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
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
	CSRF        CSRFMiddlewareConfig
}

// New creates a new middleware manager
func New(logger logging.Logger) *Manager {
	return &Manager{
		logger: logger,
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

	// Security headers
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			nonce, ok := c.Get("nonce").(string)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "nonce not found")
			}
			m.setSecurityHeaders(c, m.buildCSP(nonce))
			return next(c)
		}
	})

	// CSRF if enabled
	if m.config != nil && m.config.EnableCSRF {
		csrfConfig := m.config.CSRF
		if csrfConfig.Logger == nil {
			csrfConfig.Logger = m.logger
		}
		e.Use(CSRF(csrfConfig))
		e.Use(CSRFToken())
	}

	m.logger.Debug("middleware setup complete")
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

// CSRFConfig holds configuration for CSRF middleware
type CSRFConfig struct {
	SecretKey string
	Secure    bool
}

func (m *Manager) buildCSP(nonce string) string {
	return fmt.Sprintf(
		"default-src 'self'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"script-src 'self' 'nonce-%s'; "+
			"img-src 'self' data:; "+
			"font-src 'self'; "+
			"connect-src 'self'; "+
			"base-uri 'self'; "+
			"form-action 'self'",
		nonce,
	)
}

// setSecurityHeaders sets all security-related headers
func (m *Manager) setSecurityHeaders(c echo.Context, csp string) {
	headers := []struct {
		key   string
		value string
	}{
		{"Content-Security-Policy", csp},
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "SAMEORIGIN"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Permissions-Policy", "geolocation=(), microphone=(), camera=()"},
		{"Cross-Origin-Opener-Policy", "same-origin"},
		{"Cross-Origin-Embedder-Policy", "require-corp"},
		{"Cross-Origin-Resource-Policy", "same-origin"},
	}

	for _, header := range headers {
		m.logger.Debug("set security header",
			logging.String("header", header.key),
			logging.String("value", header.value),
		)
		c.Response().Header().Set(header.key, header.value)
	}

	// Remove Server header
	c.Response().Header().Del("Server")
	m.logger.Debug("removed Server header")
	m.logger.Debug("security headers processing complete")
}
