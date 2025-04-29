package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

const (
	// NonceSize is the size of the nonce in bytes (32 bytes = 256 bits)
	NonceSize = 32
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

// generateNonce creates a cryptographically secure random nonce
func (m *Manager) generateNonce() (string, error) {
	nonceBytes := make([]byte, NonceSize)
	if _, err := rand.Read(nonceBytes); err != nil {
		m.logger.Error("failed to generate nonce", logging.Error(err))
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}

// isStaticAsset checks if the request is for a static asset
func isStaticAsset(path string) bool {
	return strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/favicon.ico") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".jpeg") ||
		strings.HasSuffix(path, ".gif") ||
		strings.HasSuffix(path, ".svg") ||
		strings.HasSuffix(path, ".ico")
}

// buildStaticCSP builds a Content Security Policy for static assets following OWASP recommendations
func (m *Manager) buildStaticCSP() string {
	return strings.Join([]string{
		"default-src 'self'",                    // Only allow resources from same origin
		"script-src 'self'",                     // Only allow scripts from same origin
		"style-src 'self' 'unsafe-inline'",      // Allow inline styles and from same origin
		"img-src 'self' data:",                  // Allow images from same origin and data URIs
		"font-src 'self'",                       // Only allow fonts from same origin
		"connect-src 'self'",                    // Only allow XHR/WebSocket to same origin
		"media-src 'self'",                      // Only allow media from same origin
		"object-src 'none'",                     // Disable plugins
		"child-src 'none'",                      // Disable child iframes
		"frame-ancestors 'none'",                // Disable framing
		"form-action 'self'",                    // Only allow forms to submit to same origin
		"base-uri 'self'",                       // Restrict base tag to same origin
		"manifest-src 'self'",                   // Restrict manifest files
		"upgrade-insecure-requests",             // Upgrade HTTP to HTTPS
		"block-all-mixed-content",               // Block mixed content
	}, "; ")
}

// buildCSP builds a Content Security Policy with nonce for dynamic content
func (m *Manager) buildCSP(nonce string) string {
	return strings.Join([]string{
		"default-src 'self'",                    // Only allow resources from same origin
		fmt.Sprintf("script-src 'self' 'nonce-%s'", nonce), // Scripts from same origin + nonce
		"style-src 'self' 'unsafe-inline'",      // Allow inline styles and from same origin
		"img-src 'self' data:",                  // Allow images from same origin and data URIs
		"font-src 'self'",                       // Only allow fonts from same origin
		"connect-src 'self'",                    // Only allow XHR/WebSocket to same origin
		"media-src 'self'",                      // Only allow media from same origin
		"object-src 'none'",                     // Disable plugins
		"child-src 'none'",                      // Disable child iframes
		"frame-ancestors 'none'",                // Disable framing
		"form-action 'self'",                    // Only allow forms to submit to same origin
		"base-uri 'self'",                       // Restrict base tag to same origin
		"manifest-src 'self'",                   // Restrict manifest files
		"upgrade-insecure-requests",             // Upgrade HTTP to HTTPS
		"block-all-mixed-content",               // Block mixed content
	}, "; ")
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

	// Security headers with nonce generation
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			var csp string

			if isStaticAsset(path) {
				// Use static CSP without nonce for static assets
				m.logger.Debug("using static CSP for asset",
					logging.String("path", path))
				csp = m.buildStaticCSP()
			} else {
				// Generate nonce for dynamic content
				nonce, err := m.generateNonce()
				if err != nil {
					m.logger.Error("failed to generate nonce", logging.Error(err))
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate security nonce")
				}

				m.logger.Debug("generated nonce for dynamic content",
					logging.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					logging.String("path", path))

				c.Set("nonce", nonce)
				csp = m.buildCSP(nonce)
			}

			m.setSecurityHeaders(c, csp)
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

// setSecurityHeaders sets all security-related headers following OWASP recommendations
func (m *Manager) setSecurityHeaders(c echo.Context, csp string) {
	headers := []struct {
		key   string
		value string
	}{
		{"Content-Security-Policy", csp},
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()"},
		{"Cross-Origin-Opener-Policy", "same-origin"},
		{"Cross-Origin-Embedder-Policy", "require-corp"},
		{"Cross-Origin-Resource-Policy", "same-origin"},
		{"Strict-Transport-Security", "max-age=31536000; includeSubDomains"},
		{"Cache-Control", "no-store, max-age=0"},
		{"Clear-Site-Data", "\"cache\",\"cookies\",\"storage\""},
	}

	for _, header := range headers {
		m.logger.Debug("set security header",
			logging.String("header", header.key),
			logging.String("value", header.value),
		)
		c.Response().Header().Set(header.key, header.value)
	}

	// Remove potentially dangerous headers
	dangerousHeaders := []string{
		"Server",
		"X-Powered-By",
		"X-AspNet-Version",
		"X-AspNetMvc-Version",
	}
	for _, header := range dangerousHeaders {
		c.Response().Header().Del(header)
		m.logger.Debug("removed dangerous header", logging.String("header", header))
	}

	m.logger.Debug("security headers processing complete")
}
