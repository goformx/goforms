package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// CSRFMiddlewareConfig holds configuration for CSRF middleware
type CSRFMiddlewareConfig struct {
	Logger       logging.Logger
	CookieName   string
	CookiePath   string
	CookieMaxAge int
	Secure       bool
}

const (
	// DefaultCSRFCookieMaxAge is the default max age for CSRF cookies (24 hours)
	DefaultCSRFCookieMaxAge = 24 * time.Hour
	// DefaultCSRFTokenLength is the default length for CSRF tokens
	DefaultCSRFTokenLength = 32
)

// DefaultCSRFConfig returns the default CSRF configuration
func DefaultCSRFConfig() CSRFMiddlewareConfig {
	return CSRFMiddlewareConfig{
		CookieName:   "_csrf",
		CookiePath:   "/",
		CookieMaxAge: int(DefaultCSRFCookieMaxAge.Seconds()),
		Secure:       true,
	}
}

// generateToken generates a random token of the specified length
func generateToken(length uint8) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// CSRF returns middleware for CSRF protection
func CSRF() echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLength:  32,
		TokenLookup:  "header:X-CSRF-Token,form:_csrf",
		ContextKey:   "csrf",
		CookieName:   "_csrf",
		CookiePath:   "/",
		CookieSecure: true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		ErrorHandler: func(err error, c echo.Context) error {
			return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
		},
		// Skip CSRF for GET requests
		Skipper: func(c echo.Context) bool {
			return c.Request().Method == http.MethodGet
		},
	})
}

// CSRFToken returns middleware to add CSRF token to templates
func CSRFToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("csrf").(string)
			if !ok {
				// If token is not found, just continue without it
				c.Logger().Warn("CSRF token not found in context")
				c.Set("csrf", "") // Set empty token in templates
				return next(c)
			}
			c.Response().Header().Set(echo.HeaderXCSRFToken, token)
			c.Set("csrf", token) // Make token available to templates
			return next(c)
		}
	}
}
