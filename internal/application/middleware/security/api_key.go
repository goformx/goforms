package security

import (
	"net/http"
	"strings"

	appconfig "github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

const (
	// ErrMsgAPIMissing is returned when API key is missing
	ErrMsgAPIMissing = "API key is required"
	// ErrMsgAPIInvalid is returned when API key is invalid
	ErrMsgAPIInvalid = "Invalid API key"
)

// APIKeyAuth handles API key authentication middleware setup
type APIKeyAuth struct {
	logger logging.Logger
	config *appconfig.Config
}

// NewAPIKeyAuth creates a new API key authenticator
func NewAPIKeyAuth(logger logging.Logger, config *appconfig.Config) *APIKeyAuth {
	return &APIKeyAuth{
		logger: logger,
		config: config,
	}
}

// Setup creates and configures API key authentication middleware
func (a *APIKeyAuth) Setup() echo.MiddlewareFunc {
	apiKeyConfig := a.config.Security.APIKey

	if !apiKeyConfig.Enabled {
		return noopMiddleware()
	}

	if len(apiKeyConfig.Keys) == 0 {
		a.logger.Warn("API key authentication enabled but no keys configured")
		return noopMiddleware()
	}

	// Normalize header name (default to X-API-Key)
	headerName := apiKeyConfig.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}

	a.logger.Info("Setting up API key authentication",
		"enabled", apiKeyConfig.Enabled,
		"header_name", headerName,
		"keys_count", len(apiKeyConfig.Keys),
		"skip_paths", apiKeyConfig.SkipPaths,
		"skip_methods", apiKeyConfig.SkipMethods,
	)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if path should be skipped
			if a.shouldSkip(c, apiKeyConfig) {
				return next(c)
			}

			// Extract API key from header or query parameter
			apiKey := a.extractAPIKey(c, headerName, apiKeyConfig.QueryParam)

			if apiKey == "" {
				a.logger.Warn("API key missing",
					"path", c.Request().URL.Path,
					"method", c.Request().Method,
					"ip", c.RealIP(),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, ErrMsgAPIMissing)
			}

			// Validate API key
			if !a.validateAPIKey(apiKey, apiKeyConfig.Keys) {
				a.logger.Warn("Invalid API key",
					"path", c.Request().URL.Path,
					"method", c.Request().Method,
					"ip", c.RealIP(),
				)
				return echo.NewHTTPError(http.StatusUnauthorized, ErrMsgAPIInvalid)
			}

			// API key is valid, proceed
			return next(c)
		}
	}
}

// shouldSkip checks if the request should skip API key validation
func (a *APIKeyAuth) shouldSkip(c echo.Context, config appconfig.APIKeyConfig) bool {
	path := c.Request().URL.Path
	method := c.Request().Method

	// Check skip paths
	for _, skipPath := range config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// Check skip methods
	for _, skipMethod := range config.SkipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}

// extractAPIKey extracts the API key from header or query parameter
func (a *APIKeyAuth) extractAPIKey(c echo.Context, headerName, queryParam string) string {
	// Try header first
	apiKey := c.Request().Header.Get(headerName)
	if apiKey != "" {
		return strings.TrimSpace(apiKey)
	}

	// Try query parameter if configured
	if queryParam != "" {
		apiKey = c.QueryParam(queryParam)
		if apiKey != "" {
			return strings.TrimSpace(apiKey)
		}
	}

	return ""
}

// validateAPIKey validates the API key against the configured keys
func (a *APIKeyAuth) validateAPIKey(apiKey string, validKeys []string) bool {
	for _, validKey := range validKeys {
		// Use constant-time comparison to prevent timing attacks
		if subtleConstantTimeCompare(apiKey, validKey) {
			return true
		}
	}
	return false
}

// subtleConstantTimeCompare performs constant-time string comparison
// This is a basic implementation - in production, consider using crypto/subtle
func subtleConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	result := 0
	for i := range len(a) {
		result |= int(a[i]) ^ int(b[i])
	}

	return result == 0
}
