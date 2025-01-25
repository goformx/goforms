package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger logging.Logger
}

// New creates a new middleware manager
func New(log logging.Logger) *Manager {
	log.Debug("creating new middleware manager")
	return &Manager{
		logger: log,
	}
}

// generateNonce creates a cryptographically secure random nonce
func generateNonce() (string, error) {
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)
	return nonce, nil
}

// Setup configures middleware for the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Debug("setting up middleware")

	// Add security middleware
	m.logger.Debug("adding security headers middleware")
	e.Use(m.securityHeaders())

	// Add request ID middleware
	m.logger.Debug("adding request ID middleware")
	e.Use(m.requestID())

	m.logger.Debug("middleware setup complete")
}

// securityHeaders adds security headers to all responses
func (m *Manager) securityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			m.logger.Debug("processing security headers",
				logging.String("path", c.Request().URL.Path),
				logging.String("method", c.Request().Method),
			)

			// Generate a nonce for this request
			nonce, err := generateNonce()
			if err != nil {
				m.logger.Error("failed to generate nonce", logging.Error(err))
				return err
			}
			m.logger.Debug("generated nonce for request")

			// Add nonce to context for templ
			c.SetRequest(c.Request().WithContext(templ.WithNonce(c.Request().Context(), nonce)))
			m.logger.Debug("added nonce to request context")

			// Build CSP directives
			directives := []string{
				"default-src 'self'",
				"style-src 'self' 'unsafe-inline'",
				fmt.Sprintf("script-src 'self' 'nonce-%s'", nonce),
				"img-src 'self' data:",
				"font-src 'self'",
				"connect-src 'self'",
				"base-uri 'self'",
				"form-action 'self'",
			}

			// Join directives with semicolons
			csp := strings.Join(directives, "; ")
			m.logger.Debug("built CSP directives", logging.String("csp", csp))

			// Set security headers
			headers := map[string]string{
				"Content-Security-Policy":      csp,
				"X-Content-Type-Options":       "nosniff",
				"X-Frame-Options":              "SAMEORIGIN",
				"X-XSS-Protection":             "1; mode=block",
				"Referrer-Policy":              "strict-origin-when-cross-origin",
				"Permissions-Policy":           "geolocation=(), microphone=(), camera=()",
				"Cross-Origin-Opener-Policy":   "same-origin",
				"Cross-Origin-Embedder-Policy": "require-corp",
				"Cross-Origin-Resource-Policy": "same-origin",
			}

			for key, value := range headers {
				c.Response().Header().Set(key, value)
				m.logger.Debug("set security header",
					logging.String("header", key),
					logging.String("value", value),
				)
			}

			// Remove unnecessary headers
			c.Response().Header().Del("Server")
			m.logger.Debug("removed Server header")

			m.logger.Debug("security headers processing complete")
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

			m.logger.Debug("processing request ID middleware",
				logging.String("request_id", requestID),
				logging.String("method", c.Request().Method),
				logging.String("path", c.Request().URL.Path),
				logging.String("remote_addr", c.Request().RemoteAddr),
			)

			m.logger.Debug("request ID middleware complete",
				logging.String("request_id", requestID),
			)

			return next(c)
		}
	}
}
