package middleware

import (
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

type ChainConfig struct {
	Enabled         bool
	MiddlewareNames []string
	Paths           []string // Path patterns for this chain
	CustomConfig    map[string]interface{}
}

// MiddlewareConfig defines the interface for middleware configuration
type MiddlewareConfig interface {
	// IsMiddlewareEnabled checks if a middleware is enabled
	IsMiddlewareEnabled(name string) bool

	// GetMiddlewareConfig returns configuration for a specific middleware
	GetMiddlewareConfig(name string) map[string]interface{}

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
func (c *middlewareConfig) GetMiddlewareConfig(name string) map[string]interface{} {
	config := make(map[string]interface{})

	// Get category
	if category := c.getMiddlewareCategory(name); category != "" {
		config["category"] = category
	}

	// Get priority
	if priority := c.getMiddlewarePriority(name); priority > 0 {
		config["priority"] = priority
	}

	// Get dependencies
	if deps := c.getMiddlewareDependencies(name); len(deps) > 0 {
		config["dependencies"] = deps
	}

	// Get conflicts
	if conflicts := c.getMiddlewareConflicts(name); len(conflicts) > 0 {
		config["conflicts"] = conflicts
	}

	// Get path patterns
	if paths := c.getMiddlewarePaths(name); len(paths) > 0 {
		config["paths"] = paths
		config["include_paths"] = paths
	}

	// Get exclude paths
	if excludePaths := c.getMiddlewareExcludePaths(name); len(excludePaths) > 0 {
		config["exclude_paths"] = excludePaths
	}

	// Get custom configuration
	if customConfig := c.getCustomMiddlewareConfig(name); len(customConfig) > 0 {
		for k, v := range customConfig {
			config[k] = v
		}
	}

	return config
}

// GetChainConfig returns configuration for a specific chain type
func (c *middlewareConfig) GetChainConfig(chainType core.ChainType) ChainConfig {
	config := ChainConfig{
		Enabled: true, // Default to enabled
	}

	// Get middleware names for this chain based on chain type
	config.MiddlewareNames = c.getChainMiddleware(chainType)

	// Get path patterns for this chain
	config.Paths = c.getChainPaths(chainType)

	// Get custom configuration
	config.CustomConfig = c.getChainCustomConfig(chainType)

	return config
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
func (c *middlewareConfig) getCustomMiddlewareConfig(name string) map[string]interface{} {
	// For now, return empty map - can be extended later with actual config access
	return nil
}

// getChainMiddleware returns middleware names for a specific chain type
func (c *middlewareConfig) getChainMiddleware(chainType core.ChainType) []string {
	switch chainType {
	case core.ChainTypeDefault:
		return []string{"recovery", "cors", "request-id", "timeout"}
	case core.ChainTypeAPI:
		return []string{"security-headers", "csrf", "rate-limit"}
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
		return []string{"/", "/public/*"}
	case core.ChainTypeStatic:
		return []string{"/static/*", "/assets/*"}
	default:
		return []string{}
	}
}

// getChainCustomConfig returns custom configuration for a specific chain type
func (c *middlewareConfig) getChainCustomConfig(chainType core.ChainType) map[string]interface{} {
	// For now, return empty map - can be extended later with actual config access
	return nil
}
