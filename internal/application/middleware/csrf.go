package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	// TokenLength is the length of the CSRF token in bytes
	TokenLength = 32
	// TokenCookieName is the name of the CSRF token cookie
	TokenCookieName = "csrf_token"
	// TokenHeaderName is the name of the CSRF token header
	TokenHeaderName = "X-CSRF-Token"
	// DefaultCSRFCookieMaxAge is the default max age for CSRF cookies (24 hours)
	DefaultCSRFCookieMaxAge = 24 * time.Hour
	// DefaultCSRFTokenLength is the default length for CSRF tokens
	DefaultCSRFTokenLength = 32
)

// CSRFMiddlewareConfig holds configuration for CSRF middleware
type CSRFMiddlewareConfig struct {
	Logger          logging.Logger
	CookieName      string
	CookiePath      string
	CookieMaxAge    int
	Secure          bool
	TokenHeaderName string
}

// DefaultCSRFConfig returns the default CSRF configuration
func DefaultCSRFConfig() CSRFMiddlewareConfig {
	return CSRFMiddlewareConfig{
		CookieName:      TokenCookieName,
		CookiePath:      "/",
		CookieMaxAge:    int(DefaultCSRFCookieMaxAge.Seconds()),
		Secure:          true,
		TokenHeaderName: "X-CSRF-Token",
	}
}

// CSRF returns middleware for CSRF protection
func CSRF(config CSRFMiddlewareConfig) echo.MiddlewareFunc {
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLength:    TokenLength,
		TokenLookup:    fmt.Sprintf("header:%s,form:_csrf", config.TokenHeaderName),
		ContextKey:     "csrf",
		CookieName:     config.CookieName,
		CookiePath:     config.CookiePath,
		CookieSecure:   config.Secure,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		ErrorHandler: func(err error, c echo.Context) error {
			return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
		},
		Skipper: func(c echo.Context) bool {
			return c.Request().Method == http.MethodGet
		},
	})
}

// CSRFToken returns middleware to add CSRF token to templates
func CSRFToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Always generate a new token for GET requests
			if c.Request().Method == http.MethodGet {
				token, err := generateToken()
				if err != nil {
					return err
				}
				c.Set("csrf", token)
				c.Response().Header().Set(echo.HeaderXCSRFToken, token)
				return next(c)
			}

			// For other methods, use existing token or generate new one
			token, ok := c.Get("csrf").(string)
			if !ok || token == "" {
				var err error
				token, err = generateToken()
				if err != nil {
					return err
				}
				c.Set("csrf", token)
			}
			c.Response().Header().Set(echo.HeaderXCSRFToken, token)
			return next(c)
		}
	}
}

// generateToken generates a random token
func generateToken() (string, error) {
	b := make([]byte, TokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
