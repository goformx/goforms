package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

const (
	nonceSize     = 32
	requestIDSize = 16
)

// Manager handles middleware configuration and setup
type Manager struct {
	logger logging.Logger
}

// New creates a new middleware manager
func New(logger logging.Logger) *Manager {
	logger.Debug("creating new middleware manager")
	return &Manager{
		logger: logger,
	}
}

// Setup configures middleware for the Echo instance
func (m *Manager) Setup(e *echo.Echo) {
	m.logger.Debug("setting up middleware")
	m.logger.Debug("adding security headers middleware")
	e.Use(m.securityHeaders())
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

			// Generate and set nonce
			nonceStr, err := m.generateNonce()
			if err != nil {
				return err
			}
			c.Set("csp-nonce", nonceStr)

			// Build and set CSP
			csp := m.buildCSP(nonceStr)
			m.logger.Debug("built CSP directives",
				logging.String("csp", csp),
			)

			// Set security headers
			m.setSecurityHeaders(c, csp)

			return next(c)
		}
	}
}

// generateNonce generates a random nonce for CSP
func (m *Manager) generateNonce() (string, error) {
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		m.logger.Error("failed to generate nonce",
			logging.Error(err),
		)
		return "", err
	}
	m.logger.Debug("generated nonce for request")
	return base64.StdEncoding.EncodeToString(nonce), nil
}

// buildCSP builds the Content Security Policy string
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

// requestID adds a unique request ID to each request
func (m *Manager) requestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = generateRequestID()
			}

			m.logger.Debug("processing request ID middleware",
				logging.String("request_id", requestID),
				logging.String("method", c.Request().Method),
				logging.String("path", c.Request().URL.Path),
				logging.String("remote_addr", c.Request().RemoteAddr),
			)

			c.Response().Header().Set(echo.HeaderXRequestID, requestID)
			c.Set("request_id", requestID)

			m.logger.Debug("request ID middleware complete",
				logging.String("request_id", requestID),
			)

			return next(c)
		}
	}
}

func generateRequestID() string {
	b := make([]byte, requestIDSize)
	if _, err := rand.Read(b); err != nil {
		return "error-generating-request-id"
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
