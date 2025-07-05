package middleware

import (
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Middleware name constants
const (
	MiddlewareNameCSRF      = "csrf"
	MiddlewareNameRateLimit = "rate_limit"
	MiddlewareNameLogging   = "logging"
)

// MiddlewareConfig defines the interface for middleware configuration
type MiddlewareConfig interface {
	// IsMiddlewareEnabled checks if a middleware is enabled
	IsMiddlewareEnabled(name string) bool

	// GetMiddlewareConfig returns configuration for a specific middleware
	GetMiddlewareConfig(name string) map[string]any

	// GetChainConfig returns configuration for a specific chain type
	GetChainConfig(chainType core.ChainType) ChainConfig
}

// ChainConfig defines the middleware configuration for each chain type
type ChainConfig struct {
	Enabled         bool
	MiddlewareNames []string
	Paths           []string // Path patterns for this chain
	CustomConfig    map[string]any
}

// ViperMiddlewareConfig implements MiddlewareConfig using Viper configuration
type ViperMiddlewareConfig struct {
	config *config.Config
	logger logging.Logger
}

// NewViperMiddlewareConfig creates a new Viper-based middleware configuration
func NewViperMiddlewareConfig(cfg *config.Config, logger logging.Logger) MiddlewareConfig {
	return &ViperMiddlewareConfig{
		config: cfg,
		logger: logger,
	}
}

// IsMiddlewareEnabled checks if a middleware is enabled based on Viper configuration
func (c *ViperMiddlewareConfig) IsMiddlewareEnabled(name string) bool {
	// Get environment-specific enabled middleware
	enabledMiddleware := c.getEnvironmentEnabledMiddleware()

	c.logger.Debug("Checking if middleware is enabled",
		"name", name,
		"enabled_middleware", enabledMiddleware)

	// Check if middleware is in the enabled list
	for _, enabled := range enabledMiddleware {
		if enabled == name {
			c.logger.Debug("Middleware is enabled", "name", name)
			return true
		}
	}

	c.logger.Debug("Middleware is disabled", "name", name)
	return false
}

// GetMiddlewareConfig returns configuration for a specific middleware from Viper
func (c *ViperMiddlewareConfig) GetMiddlewareConfig(name string) map[string]any {
	mwConfig := make(map[string]any)

	// Get category
	if category := c.getMiddlewareCategory(name); category != "" {
		mwConfig["category"] = category
	}

	// Get priority
	if priority := c.getMiddlewarePriority(name); priority > 0 {
		mwConfig["priority"] = priority
	}

	// Get dependencies
	if deps := c.getMiddlewareDependencies(name); len(deps) > 0 {
		mwConfig["dependencies"] = deps
	}

	// Get conflicts
	if conflicts := c.getMiddlewareConflicts(name); len(conflicts) > 0 {
		mwConfig["conflicts"] = conflicts
	}

	// Get path patterns from Viper config
	if paths := c.getMiddlewarePaths(name); len(paths) > 0 {
		mwConfig["paths"] = paths
		mwConfig["include_paths"] = paths
	}

	// Get exclude paths from Viper config
	if excludePaths := c.getMiddlewareExcludePaths(name); len(excludePaths) > 0 {
		mwConfig["exclude_paths"] = excludePaths
	}

	// Get custom configuration from Viper
	if customConfig := c.getCustomMiddlewareConfig(name); len(customConfig) > 0 {
		for k, v := range customConfig {
			mwConfig[k] = v
		}
	}

	return mwConfig
}

// GetChainConfig returns configuration for a specific chain type from Viper
func (c *ViperMiddlewareConfig) GetChainConfig(chainType core.ChainType) ChainConfig {
	chainConfig := ChainConfig{
		Enabled: true, // Default to enabled
	}

	// Get chain configuration from Viper based on chain type
	switch chainType {
	case core.ChainTypeDefault:
		chainConfig.Enabled = c.config.Middleware.Chains.Default.Enabled
		chainConfig.MiddlewareNames = c.config.Middleware.Chains.Default.MiddlewareNames
		chainConfig.Paths = c.config.Middleware.Chains.Default.Paths
		chainConfig.CustomConfig = c.config.Middleware.Chains.Default.CustomConfig
	case core.ChainTypeAPI:
		chainConfig.Enabled = c.config.Middleware.Chains.API.Enabled
		chainConfig.MiddlewareNames = c.config.Middleware.Chains.API.MiddlewareNames
		chainConfig.Paths = c.config.Middleware.Chains.API.Paths
		chainConfig.CustomConfig = c.config.Middleware.Chains.API.CustomConfig
	case core.ChainTypeWeb:
		chainConfig.Enabled = c.config.Middleware.Chains.Web.Enabled
		chainConfig.MiddlewareNames = c.config.Middleware.Chains.Web.MiddlewareNames
		chainConfig.Paths = c.config.Middleware.Chains.Web.Paths
		chainConfig.CustomConfig = c.config.Middleware.Chains.Web.CustomConfig
	case core.ChainTypeAuth:
		chainConfig.Enabled = c.config.Middleware.Chains.Auth.Enabled
		chainConfig.MiddlewareNames = c.config.Middleware.Chains.Auth.MiddlewareNames
		chainConfig.Paths = c.config.Middleware.Chains.Auth.Paths
		chainConfig.CustomConfig = c.config.Middleware.Chains.Auth.CustomConfig
	case core.ChainTypeAdmin:
		chainConfig.Enabled = c.config.Middleware.Chains.Admin.Enabled
		chainConfig.MiddlewareNames = c.config.Middleware.Chains.Admin.MiddlewareNames
		chainConfig.Paths = c.config.Middleware.Chains.Admin.Paths
		chainConfig.CustomConfig = c.config.Middleware.Chains.Admin.CustomConfig
	case core.ChainTypePublic:
		chainConfig.Enabled = c.config.Middleware.Chains.Public.Enabled
		chainConfig.MiddlewareNames = c.config.Middleware.Chains.Public.MiddlewareNames
		chainConfig.Paths = c.config.Middleware.Chains.Public.Paths
		chainConfig.CustomConfig = c.config.Middleware.Chains.Public.CustomConfig
	case core.ChainTypeStatic:
		chainConfig.Enabled = c.config.Middleware.Chains.Static.Enabled
		chainConfig.MiddlewareNames = c.config.Middleware.Chains.Static.MiddlewareNames
		chainConfig.Paths = c.config.Middleware.Chains.Static.Paths
		chainConfig.CustomConfig = c.config.Middleware.Chains.Static.CustomConfig
	}

	return chainConfig
}

// getEnvironmentEnabledMiddleware returns the list of middleware enabled for the current environment
func (c *ViperMiddlewareConfig) getEnvironmentEnabledMiddleware() []string {
	switch c.config.App.Environment {
	case "development":
		return c.config.Middleware.Global.Development
	case "production":
		return c.config.Middleware.Global.Production
	case "staging":
		return c.config.Middleware.Global.Staging
	case "test":
		return c.config.Middleware.Global.Test
	default:
		return c.config.Middleware.Global.DefaultEnabled
	}
}

// getMiddlewareCategory returns the category for a middleware
func (c *ViperMiddlewareConfig) getMiddlewareCategory(name string) core.MiddlewareCategory {
	categories := map[string]core.MiddlewareCategory{
		"recovery":         core.MiddlewareCategoryBasic,
		"cors":             core.MiddlewareCategoryBasic,
		"request_id":       core.MiddlewareCategoryBasic,
		"timeout":          core.MiddlewareCategoryBasic,
		"logging":          core.MiddlewareCategoryLogging,
		"security_headers": core.MiddlewareCategorySecurity,
		"csrf":             core.MiddlewareCategorySecurity,
		"rate_limit":       core.MiddlewareCategorySecurity,
		"input_validation": core.MiddlewareCategorySecurity,
		"session":          core.MiddlewareCategoryAuth,
		"authentication":   core.MiddlewareCategoryAuth,
		"authorization":    core.MiddlewareCategoryAuth,
	}

	if category, exists := categories[name]; exists {
		return category
	}

	return core.MiddlewareCategoryBasic
}

// getMiddlewarePriority returns the priority for a middleware
func (c *ViperMiddlewareConfig) getMiddlewarePriority(name string) int {
	priorities := map[string]int{
		"recovery":         10,
		"cors":             20,
		"request_id":       30,
		"timeout":          40,
		"security_headers": 50,
		"csrf":             60,
		"rate_limit":       70,
		"input_validation": 80,
		"logging":          90,
		"session":          100,
		"authentication":   110,
		"authorization":    120,
	}

	if priority, exists := priorities[name]; exists {
		return priority
	}

	return 50 // Default priority
}

// getMiddlewareDependencies returns dependencies for a middleware
func (c *ViperMiddlewareConfig) getMiddlewareDependencies(name string) []string {
	dependencies := map[string][]string{
		"authorization": {"authentication"},
		"csrf":          {"session"},
	}

	if deps, exists := dependencies[name]; exists {
		return deps
	}

	return nil
}

// getMiddlewareConflicts returns conflicts for a middleware
func (c *ViperMiddlewareConfig) getMiddlewareConflicts(name string) []string {
	conflicts := map[string][]string{
		"csrf": {"no-csrf"},
	}

	if confs, exists := conflicts[name]; exists {
		return confs
	}

	return nil
}

// getMiddlewarePaths returns path patterns for a middleware from Viper config
func (c *ViperMiddlewareConfig) getMiddlewarePaths(name string) []string {
	switch name {
	case MiddlewareNameCSRF:
		return c.config.Middleware.CSRF.IncludePaths
	case MiddlewareNameRateLimit:
		return c.config.Middleware.RateLimit.IncludePaths
	case MiddlewareNameLogging:
		return c.config.Middleware.Logging.IncludePaths
	default:
		return nil
	}
}

// getMiddlewareExcludePaths returns exclude path patterns for a middleware from Viper config
func (c *ViperMiddlewareConfig) getMiddlewareExcludePaths(name string) []string {
	switch name {
	case MiddlewareNameCSRF:
		return c.config.Middleware.CSRF.ExcludePaths
	case MiddlewareNameRateLimit:
		return c.config.Middleware.RateLimit.ExcludePaths
	case MiddlewareNameLogging:
		return c.config.Middleware.Logging.ExcludePaths
	default:
		return nil
	}
}

// getCustomMiddlewareConfig returns custom configuration for a middleware from Viper
func (c *ViperMiddlewareConfig) getCustomMiddlewareConfig(name string) map[string]any {
	switch name {
	case MiddlewareNameCSRF:
		return map[string]any{
			"token_header": c.config.Middleware.CSRF.TokenHeader,
			"cookie_name":  c.config.Middleware.CSRF.CookieName,
			"expire_time":  c.config.Middleware.CSRF.ExpireTime,
		}
	case MiddlewareNameRateLimit:
		return map[string]any{
			"requests_per_minute": c.config.Middleware.RateLimit.RequestsPerMinute,
			"burst_size":          c.config.Middleware.RateLimit.BurstSize,
			"window_size":         c.config.Middleware.RateLimit.WindowSize,
		}
	case "timeout":
		return map[string]any{
			"timeout_seconds": c.config.Middleware.Timeout.TimeoutSeconds,
			"grace_period":    c.config.Middleware.Timeout.GracePeriod,
		}
	case MiddlewareNameLogging:
		return map[string]any{
			"log_level":     c.config.Middleware.Logging.LogLevel,
			"include_body":  c.config.Middleware.Logging.IncludeBody,
			"mask_headers":  c.config.Middleware.Logging.MaskHeaders,
			"log_requests":  c.config.Middleware.Logging.LogRequests,
			"log_responses": c.config.Middleware.Logging.LogResponses,
		}
	case "session":
		return map[string]any{
			"session_timeout": c.config.Middleware.Session.SessionTimeout,
			"refresh_timeout": c.config.Middleware.Session.RefreshTimeout,
			"secure_cookies":  c.config.Middleware.Session.SecureCookies,
			"http_only":       c.config.Middleware.Session.HTTPOnly,
		}
	case "authentication":
		return map[string]any{
			"token_expiry":   c.config.Middleware.Authentication.TokenExpiry,
			"refresh_expiry": c.config.Middleware.Authentication.RefreshExpiry,
		}
	case "authorization":
		return map[string]any{
			"default_role": c.config.Middleware.Authorization.DefaultRole,
			"admin_role":   c.config.Middleware.Authorization.AdminRole,
			"cache_ttl":    c.config.Middleware.Authorization.CacheTTL,
		}
	default:
		return map[string]any{
			"enabled": true,
		}
	}
}
