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
	return base64.StdEncoding.EncodeToString(nonceBytes), nil
}

// Setup configures middleware for the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Info("middleware configuration")

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
				m.logger.Error("failed to generate nonce", logging.Error(err))
				return err
			}

			// Add nonce to context for templ
			c.SetRequest(c.Request().WithContext(templ.WithNonce(c.Request().Context(), nonce)))

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

			// Set security headers
			c.Response().Header().Set("Content-Security-Policy", csp)
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "SAMEORIGIN")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			c.Response().Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
			c.Response().Header().Set("Cross-Origin-Resource-Policy", "same-origin")

			// Remove unnecessary headers
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
				logging.String("request_id", requestID),
				logging.String("method", c.Request().Method),
				logging.String("path", c.Request().URL.Path),
				logging.String("remote_addr", c.Request().RemoteAddr),
			)

			return next(c)
		}
	}
}
