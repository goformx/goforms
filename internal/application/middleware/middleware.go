package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/logger"
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger logger.Logger
	config *config.Config
}

// New creates a new middleware manager
func New(log logger.Logger, cfg *config.Config) *Manager {
	return &Manager{
		logger: log,
		config: cfg,
	}
}

// generateNonce creates a cryptographically secure random nonce
func generateNonce() (string, error) {
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return base64.StdEncoding.EncodeToString(nonceBytes), nil
}

// Setup configures middleware for the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Info("middleware configuration",
		logger.String("cors", formatConfig(m.config.Security.CorsAllowedOrigins)),
		logger.String("rate_limit", formatConfig(m.config.RateLimit)),
		logger.String("request_timeout", m.config.Security.RequestTimeout.String()),
	)

	// Add security middleware
	e.Use(m.securityHeaders())

	// Add request ID middleware
	e.Use(m.requestID())
}

// securityHeaders adds security headers to all responses
func (m *Manager) securityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Generate a nonce for this request
			nonce, err := generateNonce()
			if err != nil {
				m.logger.Error("failed to generate nonce", logger.Error(err))
				return err
			}

			m.logger.Debug("generated nonce",
				logger.String("nonce", nonce),
				logger.String("request_id", c.Get("request_id").(string)),
			)

			// Add nonce to context for templ
			c.SetRequest(c.Request().WithContext(templ.WithNonce(c.Request().Context(), nonce)))

			// Build CSP directives with proper ordering and spacing
			directives := []string{
				"default-src 'self'",
				"style-src 'self' 'unsafe-inline'",                 // templ uses static style blocks
				fmt.Sprintf("script-src 'self' 'nonce-%s'", nonce), // scripts still need nonce
				"img-src 'self' data:",
				"font-src 'self'",
				"connect-src 'self'",
				"base-uri 'self'",
				"form-action 'self'",
			}

			// Join directives with semicolons and spaces for better browser parsing
			csp := strings.Join(directives, "; ")

			// Enhanced logging for CSP debugging
			m.logger.Debug("setting security headers",
				logger.String("csp", csp),
				logger.String("nonce", nonce),
				logger.String("request_id", c.Get("request_id").(string)),
				logger.String("path", c.Request().URL.Path),
				logger.String("method", c.Request().Method),
				logger.String("origin", c.Request().Header.Get("Origin")),
				logger.String("referer", c.Request().Header.Get("Referer")),
				logger.String("user_agent", c.Request().Header.Get("User-Agent")),
			)

			// Set CSP header
			c.Response().Header().Set("Content-Security-Policy", csp)

			// Log all response headers for debugging
			m.logger.Debug("response headers set",
				logger.String("request_id", c.Get("request_id").(string)),
				logger.String("headers", formatHeaders(c.Response().Header())),
			)

			// Set other security headers
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			c.Response().Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
			c.Response().Header().Set("Cross-Origin-Resource-Policy", "same-origin")

			// Remove unnecessary headers
			c.Response().Header().Del("X-XSS-Protection")
			c.Response().Header().Del("Server")

			return next(c)
		}
	}
}

// requestID adds a unique request ID to each request
func (m *Manager) requestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := uuid.New().String()
			c.Set("request_id", requestID)

			m.logger.Debug("incoming request",
				logger.String("request_id", requestID),
				logger.String("origin", c.Request().Header.Get("Origin")),
				logger.String("method", c.Request().Method),
				logger.String("path", c.Request().URL.Path),
				logger.String("remote_addr", c.Request().RemoteAddr),
				logger.String("headers", formatHeaders(c.Request().Header)),
			)

			return next(c)
		}
	}
}

// formatConfig converts a config value to a string representation
func formatConfig(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// formatHeaders converts HTTP headers to a string representation
func formatHeaders(h http.Header) string {
	b, err := json.Marshal(h)
	if err != nil {
		return ""
	}
	return string(b)
}
