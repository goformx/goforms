package middleware

import (
	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

type ChainConfig struct {
	Enabled         bool
	MiddlewareNames []string
	Paths           []string // Path patterns for this chain
	CustomConfig    map[string]any
}

// MiddlewareConfig defines the interface for middleware configuration
type MiddlewareConfig interface {
	// IsMiddlewareEnabled checks if a middleware is enabled
	IsMiddlewareEnabled(name string) bool

	// GetMiddlewareConfig returns configuration for a specific middleware
	GetMiddlewareConfig(name string) map[string]any

	// GetChainConfig returns configuration for a specific chain type
	GetChainConfig(chainType core.ChainType) ChainConfig
}

// middlewareConfig implements the MiddlewareConfig interface
type middlewareConfig struct {
	config *config.Config
	logger logging.Logger
}

// NewMiddlewareConfig creates a new middleware configuration provider
func NewMiddlewareConfig(cfg *config.Config, logger logging.Logger) MiddlewareConfig {
	return &middlewareConfig{
		config: cfg,
		logger: logger,
	}
}

// IsMiddlewareEnabled checks if a middleware is enabled based on configuration
func (c *middlewareConfig) IsMiddlewareEnabled(name string) bool {
	// Default enabled middleware based on environment
	defaultEnabled := c.getDefaultEnabledMiddleware()
	for _, enabled := range defaultEnabled {
		if enabled == name {
			return true
		}
	}

	// Check environment-specific overrides
	if c.config.App.IsDevelopment() {
		// In development, enable all middleware by default
		return true
	}

	// In production, be more selective
	productionEnabled := []string{
		"recovery",
		"cors",
		"security-headers",
		"request-id",
		"timeout",
		"logging",
		"csrf",
		"rate-limit",
		"session",
		"authentication",
		"authorization",
	}

	for _, enabled := range productionEnabled {
		if enabled == name {
			return true
		}
	}

	return false
}

// GetMiddlewareConfig returns configuration for a specific middleware
func (c *middlewareConfig) GetMiddlewareConfig(name string) map[string]any {
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

	// Get path patterns
	if paths := c.getMiddlewarePaths(name); len(paths) > 0 {
		mwConfig["paths"] = paths
		mwConfig["include_paths"] = paths
	}

	// Get exclude paths
	if excludePaths := c.getMiddlewareExcludePaths(name); len(excludePaths) > 0 {
		mwConfig["exclude_paths"] = excludePaths
	}

	// Get custom configuration
	if customConfig := c.getCustomMiddlewareConfig(name); len(customConfig) > 0 {
		for k, v := range customConfig {
			mwConfig[k] = v
		}
	}

	return mwConfig
}

// GetChainConfig returns configuration for a specific chain type
func (c *middlewareConfig) GetChainConfig(chainType core.ChainType) ChainConfig {
	chainConfig := ChainConfig{
		Enabled: true, // Default to enabled
	}

	// Get middleware names for this chain based on chain type
	chainConfig.MiddlewareNames = c.getChainMiddleware(chainType)

	// Get path patterns for this chain
	chainConfig.Paths = c.getChainPaths(chainType)

	// Get custom configuration
	chainConfig.CustomConfig = c.getChainCustomConfig(chainType)

	return chainConfig
}

// getDefaultEnabledMiddleware returns the list of middleware enabled by default
func (c *middlewareConfig) getDefaultEnabledMiddleware() []string {
	if c.config.App.IsDevelopment() {
		return []string{
			"recovery",
			"cors",
			"request-id",
			"logging",
			"session",
			"authentication",
			"authorization",
		}
	}

	return []string{
		"recovery",
		"cors",
		"security-headers",
		"request-id",
		"timeout",
		"logging",
		"csrf",
		"rate-limit",
		"session",
		"authentication",
		"authorization",
	}
}

// getMiddlewareCategory returns the category for a middleware
func (c *middlewareConfig) getMiddlewareCategory(name string) core.MiddlewareCategory {
	categories := map[string]core.MiddlewareCategory{
		"recovery":         core.MiddlewareCategoryBasic,
		"cors":             core.MiddlewareCategoryBasic,
		"request-id":       core.MiddlewareCategoryBasic,
		"timeout":          core.MiddlewareCategoryBasic,
		"logging":          core.MiddlewareCategoryLogging,
		"security-headers": core.MiddlewareCategorySecurity,
		"csrf":             core.MiddlewareCategorySecurity,
		"rate-limit":       core.MiddlewareCategorySecurity,
		"input-validation": core.MiddlewareCategorySecurity,
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
func (c *middlewareConfig) getMiddlewarePriority(name string) int {
	priorities := map[string]int{
		"recovery":         constants.PriorityRecovery,
		"cors":             constants.PriorityCORS,
		"request-id":       constants.PriorityRequestID,
		"timeout":          constants.PriorityTimeout,
		"security-headers": constants.PrioritySecurityHeaders,
		"csrf":             constants.PriorityCSRF,
		"rate-limit":       constants.PriorityRateLimit,
		"input-validation": constants.PriorityInputValidation,
		"logging":          constants.PriorityLogging,
		"session":          constants.PrioritySession,
		"authentication":   constants.PriorityAuthentication,
		"authorization":    constants.PriorityAuthorization,
	}

	if priority, exists := priorities[name]; exists {
		return priority
	}

	return constants.DefaultMiddlewarePriority
}

// getMiddlewareDependencies returns dependencies for a middleware
func (c *middlewareConfig) getMiddlewareDependencies(name string) []string {
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
func (c *middlewareConfig) getMiddlewareConflicts(name string) []string {
	conflicts := map[string][]string{
		"csrf": {"no-csrf"},
	}

	if confs, exists := conflicts[name]; exists {
		return confs
	}

	return nil
}

// getMiddlewarePaths returns path patterns for a middleware
func (c *middlewareConfig) getMiddlewarePaths(name string) []string {
	paths := map[string][]string{
		"csrf":       {"/api/*", "/forms/*"},
		"rate-limit": {"/api/*"},
	}

	if pathList, exists := paths[name]; exists {
		return pathList
	}

	return nil
}

// getMiddlewareExcludePaths returns exclude path patterns for a middleware
func (c *middlewareConfig) getMiddlewareExcludePaths(name string) []string {
	excludePaths := map[string][]string{
		"csrf":       {"/api/public/*", "/static/*"},
		"rate-limit": {"/health", "/metrics"},
	}

	if excludeList, exists := excludePaths[name]; exists {
		return excludeList
	}

	return nil
}

// getCustomMiddlewareConfig returns custom configuration for a middleware
func (c *middlewareConfig) getCustomMiddlewareConfig(name string) map[string]any {
	// Return custom configuration based on middleware name
	customConfigs := map[string]map[string]any{
		"csrf": {
			"token_header": "X-CSRF-Token",
			"cookie_name":  "csrf_token",
			"expire_time":  constants.SessionExpiry,
		},
		"rate-limit": {
			"requests_per_minute": constants.RequestsPerMinute,
			"burst_size":          constants.BurstSizeDefault,
			"window_size":         constants.WindowSizeDefault,
		},
		"timeout": {
			"timeout_seconds": constants.TimeoutDefault,
			"grace_period":    constants.GracePeriod,
		},
		"logging": {
			"log_level":     "info",
			"include_body":  false,
			"mask_headers":  []string{"authorization", "cookie"},
			"log_requests":  true,
			"log_responses": true,
		},
		"session": {
			"session_timeout": constants.SessionExpiry,
			"refresh_timeout": constants.RefreshTimeout,
			"secure_cookies":  true,
			"http_only":       true,
		},
		"authentication": {
			"jwt_secret":     "your-secret-key",
			"token_expiry":   constants.TokenExpiry,
			"refresh_expiry": constants.RefreshExpiry,
		},
		"authorization": {
			"default_role": "user",
			"admin_role":   "admin",
			"cache_ttl":    constants.CacheTTLShortSec,
		},
	}

	if customConfig, exists := customConfigs[name]; exists {
		return customConfig
	}

	// Return default configuration for unknown middleware
	return map[string]any{
		"enabled": true,
	}
}

// getChainMiddleware returns middleware names for a specific chain type
func (c *middlewareConfig) getChainMiddleware(chainType core.ChainType) []string {
	switch chainType {
	case core.ChainTypeDefault:
		return []string{"recovery", "cors", "request-id", "timeout"}
	case core.ChainTypeAPI:
		return []string{"security-headers", "session", "csrf", "rate-limit", "authentication", "authorization"}
	case core.ChainTypeWeb:
		return []string{"session", "authentication", "authorization"}
	case core.ChainTypeAuth:
		return []string{"session", "authentication"}
	case core.ChainTypeAdmin:
		return []string{"session", "authentication", "authorization"}
	case core.ChainTypePublic:
		return []string{"recovery", "cors"}
	case core.ChainTypeStatic:
		return []string{"recovery"}
	default:
		return []string{}
	}
}

// getChainPaths returns path patterns for a specific chain type
func (c *middlewareConfig) getChainPaths(chainType core.ChainType) []string {
	switch chainType {
	case core.ChainTypeDefault:
		return []string{"/*"}
	case core.ChainTypeAPI:
		return []string{"/api/*"}
	case core.ChainTypeWeb:
		return []string{"/dashboard/*", "/forms/*"}
	case core.ChainTypeAuth:
		return []string{"/login", "/signup", "/logout"}
	case core.ChainTypeAdmin:
		return []string{"/admin/*"}
	case core.ChainTypePublic:
		return []string{"/public/*"}
	case core.ChainTypeStatic:
		return []string{"/static/*", "/assets/*"}
	default:
		return []string{}
	}
}

// getChainCustomConfig returns custom configuration for a specific chain type
func (c *middlewareConfig) getChainCustomConfig(chainType core.ChainType) map[string]any {
	customConfigs := c.getChainCustomConfigs()

	if chainConfig, exists := customConfigs[chainType]; exists {
		return chainConfig
	}

	// Return default configuration for unknown chain types
	return map[string]any{
		"enabled": true,
		"timeout": constants.TimeoutDefault,
	}
}

// Chain custom configs as package-level variables
var chainCustomConfigDefault = map[string]any{
	"timeout":          constants.TimeoutDefault,
	"max_body_size":    "10MB",
	"compress":         true,
	"cors_origins":     []string{"*"},
	"security_headers": true,
}

var chainCustomConfigAPI = map[string]any{
	"timeout":          constants.TimeoutMedium,
	"max_body_size":    "50MB",
	"compress":         true,
	"cors_origins":     []string{"https://api.example.com"},
	"rate_limit":       true,
	"authentication":   true,
	"authorization":    true,
	"request_logging":  true,
	"response_logging": false,
}

var chainCustomConfigWeb = map[string]any{
	"timeout":          constants.TimeoutDefault,
	"max_body_size":    "25MB",
	"compress":         true,
	"cors_origins":     []string{"https://app.example.com"},
	"session":          true,
	"authentication":   true,
	"authorization":    true,
	"request_logging":  true,
	"response_logging": false,
}

var chainCustomConfigAuth = map[string]any{
	"timeout":          constants.TimeoutAuth,
	"max_body_size":    "5MB",
	"compress":         false,
	"cors_origins":     []string{"https://auth.example.com"},
	"session":          true,
	"authentication":   true,
	"csrf_protection":  true,
	"request_logging":  true,
	"response_logging": false,
}

var chainCustomConfigAdmin = map[string]any{
	"timeout":          constants.TimeoutMedium,
	"max_body_size":    "100MB",
	"compress":         true,
	"cors_origins":     []string{"https://admin.example.com"},
	"session":          true,
	"authentication":   true,
	"authorization":    true,
	"rate_limit":       true,
	"request_logging":  true,
	"response_logging": true,
	"audit_logging":    true,
}

var chainCustomConfigPublic = map[string]any{
	"timeout":          constants.TimeoutPublic,
	"max_body_size":    "1MB",
	"compress":         true,
	"cors_origins":     []string{"*"},
	"session":          false,
	"authentication":   false,
	"authorization":    false,
	"request_logging":  false,
	"response_logging": false,
}

var chainCustomConfigStatic = map[string]any{
	"timeout":          constants.TimeoutShort,
	"max_body_size":    "100MB",
	"compress":         true,
	"cors_origins":     []string{"*"},
	"session":          false,
	"authentication":   false,
	"authorization":    false,
	"request_logging":  false,
	"response_logging": false,
	"cache_headers":    true,
	"cache_duration":   constants.CacheDurationDay,
}

// getChainCustomConfigs returns the complete chain configuration map
func (c *middlewareConfig) getChainCustomConfigs() map[core.ChainType]map[string]any {
	return map[core.ChainType]map[string]any{
		core.ChainTypeDefault: chainCustomConfigDefault,
		core.ChainTypeAPI:     chainCustomConfigAPI,
		core.ChainTypeWeb:     chainCustomConfigWeb,
		core.ChainTypeAuth:    chainCustomConfigAuth,
		core.ChainTypeAdmin:   chainCustomConfigAdmin,
		core.ChainTypePublic:  chainCustomConfigPublic,
		core.ChainTypeStatic:  chainCustomConfigStatic,
	}
}
