// Package security provides security-related middleware configuration.
package security

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/goformx/goforms/internal/application/constants"
	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
)

// SetupCSRF creates and configures CSRF middleware
func SetupCSRF(csrfConfig *appconfig.CSRFConfig, isDevelopment bool) echo.MiddlewareFunc {
	sameSite := getSameSite(csrfConfig.CookieSameSite, isDevelopment)
	tokenLength := getTokenLength(csrfConfig.TokenLength)

	// Log CSRF configuration
	if isDevelopment {
		println("[CSRF] Setting up CSRF middleware with context_key:", csrfConfig.ContextKey)
	}

	return echomw.CSRFWithConfig(echomw.CSRFConfig{
		TokenLength:    uint8(tokenLength), // #nosec G115
		TokenLookup:    csrfConfig.TokenLookup,
		ContextKey:     csrfConfig.ContextKey,
		CookieName:     csrfConfig.CookieName,
		CookiePath:     csrfConfig.CookiePath,
		CookieDomain:   csrfConfig.CookieDomain,
		CookieSecure:   !isDevelopment,
		CookieHTTPOnly: csrfConfig.CookieHTTPOnly,
		CookieSameSite: sameSite,
		CookieMaxAge:   csrfConfig.CookieMaxAge,
		Skipper:        CreateCSRFSkipper(isDevelopment),
		ErrorHandler:   CreateCSRFErrorHandler(csrfConfig, isDevelopment),
	})
}

// getSameSite converts string SameSite to http.SameSite
func getSameSite(cookieSameSite string, isDevelopment bool) http.SameSite {
	switch cookieSameSite {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		if isDevelopment {
			return http.SameSiteLaxMode
		}
		return http.SameSiteStrictMode
	}
}

// getTokenLength ensures token length is within bounds for uint8
func getTokenLength(tokenLength int) int {
	if tokenLength <= 0 || tokenLength > 255 {
		return constants.DefaultTokenLength
	}
	return tokenLength
}

// CreateCSRFSkipper creates a CSRF skipper function
func CreateCSRFSkipper(isDevelopment bool) func(c echo.Context) bool {
	return func(c echo.Context) bool {
		path := c.Request().URL.Path
		method := c.Request().Method

		if isDevelopment {
			logCSRFSkipperDebug(c, path, method)
		}

		if IsSafeMethod(method) {
			return handleSafeMethodCSRF(c, path, isDevelopment)
		}

		if shouldSkipCSRFForRoute(path, isDevelopment) {
			return true
		}

		if isDevelopment {
			c.Logger().Debug("CSRF not skipped - requires protection", "path", path, "method", method)
		}

		return false
	}
}

// logCSRFSkipperDebug logs debug information for CSRF skipper
func logCSRFSkipperDebug(c echo.Context, path, method string) {
	c.Logger().Debug("CSRF skipper check",
		"path", path,
		"method", method,
		"is_safe_method", IsSafeMethod(method),
		"is_auth_page", IsAuthPage(path),
		"is_form_page", IsFormPage(path),
		"is_api_route", IsAPIRoute(path),
		"is_health_route", IsHealthRoute(path),
		"is_static_route", IsStaticRoute(path),
		"is_form_submission_route", IsFormSubmissionRoute(path),
		"is_auth_endpoint", IsAuthEndpoint(path))
}

// handleSafeMethodCSRF handles CSRF logic for safe HTTP methods
func handleSafeMethodCSRF(c echo.Context, path string, isDevelopment bool) bool {
	if IsAuthPage(path) || IsFormPage(path) {
		if isDevelopment {
			c.Logger().Debug("CSRF not skipped - token generation needed", "path", path)
		}
		return false
	}

	if isDevelopment {
		c.Logger().Debug("CSRF skipped - safe method", "path", path, "method", c.Request().Method)
	}

	return true
}

// shouldSkipCSRFForRoute checks if CSRF should be skipped for the given route
func shouldSkipCSRFForRoute(path string, isDevelopment bool) bool {
	if IsAuthEndpoint(path) {
		return true
	}

	if isDevelopment && IsAPIRoute(path) {
		return true
	}

	if IsHealthRoute(path) {
		return true
	}

	if IsStaticRoute(path) {
		return true
	}

	if IsFormSubmissionRoute(path) {
		return true
	}

	return false
}

// CreateCSRFErrorHandler creates the CSRF error handler function
func CreateCSRFErrorHandler(
	csrfConfig *appconfig.CSRFConfig,
	isDevelopment bool,
) func(err error, c echo.Context) error {
	return func(err error, c echo.Context) error {
		if isDevelopment {
			csrfToken := c.Request().Header.Get("X-Csrf-Token")
			contextToken := ""
			if token, ok := c.Get(csrfConfig.ContextKey).(string); ok {
				contextToken = token
			}
			cookies := c.Request().Header.Get("Cookie")

			c.Logger().Error("CSRF validation failed",
				"error", err.Error(),
				"path", c.Request().URL.Path,
				"method", c.Request().Method,
				"token_lookup", csrfConfig.TokenLookup,
				"origin", c.Request().Header.Get("Origin"),
				"csrf_token_present", csrfToken != "",
				"csrf_token_length", len(csrfToken),
				"csrf_token_value", csrfToken,
				"context_token_present", contextToken != "",
				"context_token_length", len(contextToken),
				"context_token_value", contextToken,
				"cookies", cookies,
				"content_type", c.Request().Header.Get("Content-Type"),
				"user_agent", c.Request().UserAgent(),
				"is_development", isDevelopment,
				"csrf_enabled", true)
		}

		return c.NoContent(http.StatusForbidden)
	}
}

// IsSafeMethod checks if the HTTP method is safe (doesn't modify state)
func IsSafeMethod(method string) bool {
	return method == "GET" || method == "HEAD" || method == "OPTIONS"
}

// IsAPIRoute checks if the path is an API route
func IsAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api/")
}

// IsHealthRoute checks if the path is a health check route
func IsHealthRoute(path string) bool {
	return path == "/health" || path == "/health/" || path == "/healthz" || path == "/healthz/"
}

// IsStaticRoute checks if the path is a static asset route
func IsStaticRoute(path string) bool {
	return strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/public/") ||
		strings.HasPrefix(path, "/favicon.ico")
}

// IsFormSubmissionRoute checks if the path is a form submission endpoint
func IsFormSubmissionRoute(path string) bool {
	return strings.Contains(path, "/api/v1/forms/") ||
		strings.Contains(path, "/forms/") ||
		strings.Contains(path, "/submit")
}

// IsAuthPage checks if the path is an authentication page
func IsAuthPage(path string) bool {
	return path == "/login" || path == "/signup" ||
		path == "/forgot-password" || path == "/reset-password"
}

// IsAuthEndpoint checks if the path is an authentication endpoint
func IsAuthEndpoint(path string) bool {
	return path == "/login" || path == "/signup" || path == "/logout" ||
		path == "/forgot-password" || path == "/reset-password"
}

// IsFormPage checks if the path is a form page
func IsFormPage(path string) bool {
	return path == "/" || strings.Contains(path, "/forms/new") ||
		strings.Contains(path, "/forms/") || strings.Contains(path, "/submit") ||
		strings.Contains(path, "/dashboard")
}
