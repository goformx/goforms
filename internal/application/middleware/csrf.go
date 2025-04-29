package middleware

import (
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

// CSRF returns middleware for CSRF protection
func CSRF(cfg CSRFMiddlewareConfig) echo.MiddlewareFunc {
	if cfg.Logger != nil {
		cfg.Logger.Debug("creating CSRF middleware",
			logging.String("cookie_name", cfg.CookieName),
			logging.String("cookie_path", cfg.CookiePath),
			logging.Int("cookie_max_age", cfg.CookieMaxAge),
			logging.Bool("secure", cfg.Secure),
		)
	}

	config := middleware.CSRFConfig{
		TokenLength:    DefaultCSRFTokenLength,
		TokenLookup:    "header:X-CSRF-Token,form:_csrf",
		ContextKey:     "csrf",
		CookieName:     cfg.CookieName,
		CookiePath:     cfg.CookiePath,
		CookieMaxAge:   cfg.CookieMaxAge,
		CookieSecure:   cfg.Secure,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
	}

	return middleware.CSRFWithConfig(config)
}

// CSRFToken returns middleware to add CSRF token to templates
func CSRFToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("csrf").(string)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "CSRF token not found")
			}
			c.Response().Header().Set(echo.HeaderXCSRFToken, token)
			c.Set("csrf_token", token) // Make token available to templates
			return next(c)
		}
	}
}
