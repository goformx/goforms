package middleware

import (
	interfaces "github.com/goformx/goforms/internal/domain/common/interfaces"
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
	GetChainConfig(chainType interfaces.ChainType) ChainConfig
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
func (c *middlewareConfig) GetChainConfig(chainType interfaces.ChainType) ChainConfig {
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
func (c *middlewareConfig) getMiddlewareCategory(name string) string {
	categories := map[string]string{
		"recovery":         "basic",
		"cors":             "basic",
		"request-id":       "basic",
		"timeout":          "basic",
		"logging":          "logging",
		"security-headers": "security",
		"csrf":             "security",
		"rate-limit":       "security",
		"input-validation": "security",
		"session":          "auth",
		"authentication":   "auth",
		"authorization":    "auth",
	}

	if category, exists := categories[name]; exists {
		return category
	}

	return "basic"
}

// getMiddlewarePriority returns the priority for a middleware
func (c *middlewareConfig) getMiddlewarePriority(name string) int {
	priorities := map[string]int{
		"recovery":         10,
		"cors":             20,
		"request-id":       30,
		"timeout":          40,
		"security-headers": 50,
		"csrf":             60,
		"rate-limit":       70,
		"input-validation": 80,
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
			"expire_time":  3600, // 1 hour
		},
		"rate-limit": {
			"requests_per_minute": 60,
			"burst_size":          10,
			"window_size":         60, // seconds
		},
		"timeout": {
			"timeout_seconds": 30,
			"grace_period":    5,
		},
		"logging": {
			"log_level":     "info",
			"include_body":  false,
			"mask_headers":  []string{"authorization", "cookie"},
			"log_requests":  true,
			"log_responses": true,
		},
		"session": {
			"session_timeout": 3600, // 1 hour
			"refresh_timeout": 300,  // 5 minutes
			"secure_cookies":  true,
			"http_only":       true,
		},
		"authentication": {
			"jwt_secret":     "your-secret-key",
			"token_expiry":   3600,  // 1 hour
			"refresh_expiry": 86400, // 24 hours
		},
		"authorization": {
			"default_role": "user",
			"admin_role":   "admin",
			"cache_ttl":    300, // 5 minutes
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
func (c *middlewareConfig) getChainMiddleware(chainType interfaces.ChainType) []string {
	switch chainType {
	case interfaces.ChainTypeGlobal:
		return []string{"recovery", "cors", "security-headers", "request-id", "timeout", "logging"}
	case interfaces.ChainTypeAPI:
		return []string{"recovery", "cors", "security-headers", "request-id", "timeout", "logging", "authentication", "authorization"}
	case interfaces.ChainTypeWeb:
		return []string{"recovery", "cors", "security-headers", "request-id", "timeout", "logging", "csrf", "session", "authentication", "authorization"}
	case interfaces.ChainTypeAuth:
		return []string{"recovery", "cors", "security-headers", "request-id", "timeout", "logging", "authentication"}
	case interfaces.ChainTypeAdmin:
		return []string{"recovery", "cors", "security-headers", "request-id", "timeout", "logging", "authentication", "authorization"}
	default:
		return []string{"recovery", "cors", "request-id", "logging"}
	}
}

// getChainPaths returns path patterns for a specific chain type
func (c *middlewareConfig) getChainPaths(chainType interfaces.ChainType) []string {
	switch chainType {
	case interfaces.ChainTypeGlobal:
		return []string{"/*"}
	case interfaces.ChainTypeAPI:
		return []string{"/api/*"}
	case interfaces.ChainTypeWeb:
		return []string{"/web/*", "/pages/*"}
	case interfaces.ChainTypeAuth:
		return []string{"/auth/*", "/login", "/logout"}
	case interfaces.ChainTypeAdmin:
		return []string{"/admin/*"}
	default:
		return []string{"/*"}
	}
}

// getChainCustomConfig returns custom configuration for a specific chain type
func (c *middlewareConfig) getChainCustomConfig(chainType interfaces.ChainType) map[string]any {
	configs := map[interfaces.ChainType]map[string]any{
		interfaces.ChainTypeGlobal: {
			"timeout": "30s",
			"cors": map[string]any{
				"allowed_origins": []string{"*"},
				"allowed_methods": []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			},
		},
		interfaces.ChainTypeAPI: {
			"timeout": "60s",
			"rate_limit": map[string]any{
				"requests_per_minute": 100,
			},
		},
		interfaces.ChainTypeWeb: {
			"timeout": "30s",
			"csrf": map[string]any{
				"token_length": 32,
			},
		},
		interfaces.ChainTypeAuth: {
			"timeout": "30s",
			"session": map[string]any{
				"max_age": 3600,
			},
		},
		interfaces.ChainTypeAdmin: {
			"timeout": "60s",
			"rate_limit": map[string]any{
				"requests_per_minute": 50,
			},
		},
	}

	if config, exists := configs[chainType]; exists {
		return config
	}

	return make(map[string]any)
}
