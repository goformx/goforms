// Package middleware provides middleware orchestration and management.
package middleware

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"

	"github.com/goformx/goforms/internal/application/middleware/chain"
	"github.com/goformx/goforms/internal/application/middleware/core"
)

// ChainInfo provides information about a middleware chain.
type ChainInfo struct {
	Type        core.ChainType
	Name        string
	Description string
	Categories  []core.MiddlewareCategory
	Middleware  []string
	Enabled     bool
}

// orchestrator implements the core.Orchestrator interface.
type orchestrator struct {
	registry core.Registry
	config   MiddlewareConfig
	logger   core.Logger
	cache    map[string]core.Chain
	cacheMu  sync.RWMutex
	chains   map[string]core.Chain
	chainsMu sync.RWMutex
}

// NewOrchestrator creates a new middleware orchestrator.
func NewOrchestrator(registry core.Registry, config MiddlewareConfig, logger core.Logger) core.Orchestrator {
	return &orchestrator{
		registry: registry,
		config:   config,
		logger:   logger,
		cache:    make(map[string]core.Chain),
		chains:   make(map[string]core.Chain),
	}
}

// CreateChain creates a new middleware chain with the specified type.
func (o *orchestrator) CreateChain(chainType core.ChainType) (core.Chain, error) {
	// Get middleware list from registry based on chain type
	middlewares, err := o.getOrderedMiddleware(chainType)
	if err != nil {
		return nil, fmt.Errorf("failed to get ordered middleware for chain type %s: %w", chainType, err)
	}

	// Filter based on configuration
	activeMiddlewares := o.filterByConfig(middlewares, chainType)

	// Validate dependencies and conflicts
	if err := o.validateChain(activeMiddlewares); err != nil {
		return nil, fmt.Errorf("chain validation failed for %s: %w", chainType, err)
	}

	// Build chain using chain builder
	chain := chain.NewChainImpl(activeMiddlewares)

	o.logger.Info("built middleware chain",
		"chain_type", chainType,
		"middleware_count", len(activeMiddlewares),
		"middleware_names", o.getMiddlewareNames(activeMiddlewares))

	return chain, nil
}

// GetChain retrieves a pre-configured chain by name.
func (o *orchestrator) GetChain(name string) (core.Chain, bool) {
	o.chainsMu.RLock()
	defer o.chainsMu.RUnlock()
	chain, ok := o.chains[name]
	return chain, ok
}

// RegisterChain registers a named chain for later retrieval.
func (o *orchestrator) RegisterChain(name string, chain core.Chain) error {
	o.chainsMu.Lock()
	defer o.chainsMu.Unlock()
	if _, exists := o.chains[name]; exists {
		return fmt.Errorf("chain with name %q already exists", name)
	}
	o.chains[name] = chain
	return nil
}

// ListChains returns all registered chain names.
func (o *orchestrator) ListChains() []string {
	o.chainsMu.RLock()
	defer o.chainsMu.RUnlock()
	names := make([]string, 0, len(o.chains))
	for name := range o.chains {
		names = append(names, name)
	}
	return names
}

// RemoveChain removes a chain by name.
func (o *orchestrator) RemoveChain(name string) bool {
	o.chainsMu.Lock()
	defer o.chainsMu.Unlock()
	if _, exists := o.chains[name]; exists {
		delete(o.chains, name)
		return true
	}
	return false
}

// BuildChainForPath creates a middleware chain for a specific path and chain type.
func (o *orchestrator) BuildChainForPath(chainType core.ChainType, requestPath string) (core.Chain, error) {
	// Get base chain
	baseChain, err := o.CreateChain(chainType)
	if err != nil {
		return nil, err
	}

	// Apply path-specific middleware
	pathChain := o.applyPathSpecificMiddleware(baseChain, requestPath)

	// Apply path-based filtering
	finalChain := o.filterByPath(pathChain, requestPath)

	o.logger.Info("built path-specific middleware chain",
		"chain_type", chainType,
		"path", requestPath,
		"middleware_count", finalChain.Length())

	return finalChain, nil
}

// GetChainForPath returns a cached chain for a path or builds a new one.
func (o *orchestrator) GetChainForPath(chainType core.ChainType, requestPath string) (core.Chain, error) {
	cacheKey := fmt.Sprintf("path:%s:%s", chainType, requestPath)

	// Check cache first
	o.cacheMu.RLock()
	if cached, exists := o.cache[cacheKey]; exists {
		o.cacheMu.RUnlock()
		return cached, nil
	}
	o.cacheMu.RUnlock()

	// Build new chain
	chain, err := o.BuildChainForPath(chainType, requestPath)
	if err != nil {
		return nil, err
	}

	// Cache the result
	o.cacheMu.Lock()
	o.cache[cacheKey] = chain
	o.cacheMu.Unlock()

	return chain, nil
}

// ClearCache clears the chain cache.
func (o *orchestrator) ClearCache() {
	o.cacheMu.Lock()
	defer o.cacheMu.Unlock()
	o.cache = make(map[string]core.Chain)
	o.logger.Info("cleared middleware chain cache")
}

// ValidateConfiguration validates the current middleware configuration.
func (o *orchestrator) ValidateConfiguration() error {
	// Validate registry dependencies
	if err := o.validateRegistryDependencies(); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
	}

	// Validate chain configurations
	for _, chainType := range []core.ChainType{
		core.ChainTypeDefault,
		core.ChainTypeAPI,
		core.ChainTypeWeb,
		core.ChainTypeAuth,
		core.ChainTypeAdmin,
		core.ChainTypePublic,
		core.ChainTypeStatic,
	} {
		if _, err := o.CreateChain(chainType); err != nil {
			return fmt.Errorf("chain validation failed for %s: %w", chainType, err)
		}
	}

	return nil
}

// GetChainInfo returns information about a chain type.
func (o *orchestrator) GetChainInfo(chainType core.ChainType) ChainInfo {
	chainConfig := o.config.GetChainConfig(chainType)

	// Get middleware for this chain type
	middlewares, _ := o.getOrderedMiddleware(chainType)
	middlewareNames := o.getMiddlewareNames(middlewares)

	// Determine categories based on chain type
	categories := o.getCategoriesForChainType(chainType)

	return ChainInfo{
		Type:        chainType,
		Name:        chainType.String(),
		Description: o.getChainDescription(chainType),
		Categories:  categories,
		Middleware:  middlewareNames,
		Enabled:     chainConfig.Enabled,
	}
}

// getOrderedMiddleware returns middleware ordered by priority for the given chain type.
func (o *orchestrator) getOrderedMiddleware(chainType core.ChainType) ([]core.Middleware, error) {
	var middlewares []core.Middleware

	// Get categories for this chain type
	categories := o.getCategoriesForChainType(chainType)

	// Collect middleware from all relevant categories
	for _, category := range categories {
		categoryMiddleware := o.getOrderedByCategory(category)
		middlewares = append(middlewares, categoryMiddleware...)
	}

	// Sort by priority (registry already does this, but ensure consistency)
	o.sortByPriority(middlewares)

	return middlewares, nil
}

// getOrderedByCategory returns middleware ordered by priority for a specific category.
func (o *orchestrator) getOrderedByCategory(category core.MiddlewareCategory) []core.Middleware {
	// This is a simplified implementation - in a real scenario, the registry would have this method
	// For now, we'll get all middleware and filter by category
	allNames := o.registry.List()
	var categoryMiddleware []core.Middleware

	for _, name := range allNames {
		if mw, ok := o.registry.Get(name); ok {
			// Check if middleware belongs to this category
			config := o.config.GetMiddlewareConfig(name)
			if catVal, ok := config["category"]; ok {
				if c, ok := catVal.(core.MiddlewareCategory); ok && c == category {
					categoryMiddleware = append(categoryMiddleware, mw)
				}
			} else if category == core.MiddlewareCategoryBasic {
				// Default to basic category if not specified
				categoryMiddleware = append(categoryMiddleware, mw)
			}
		}
	}

	// Sort by priority
	o.sortByPriority(categoryMiddleware)

	return categoryMiddleware
}

// filterByConfig filters middleware based on configuration settings.
func (o *orchestrator) filterByConfig(middlewares []core.Middleware, chainType core.ChainType) []core.Middleware {
	var filtered []core.Middleware
	chainConfig := o.config.GetChainConfig(chainType)

	for _, mw := range middlewares {
		name := mw.Name()

		// Check if middleware is enabled globally
		if !o.config.IsMiddlewareEnabled(name) {
			o.logger.Info("middleware disabled by config", "name", name)
			continue
		}

		// Check if middleware is enabled for this chain
		if !chainConfig.Enabled {
			o.logger.Info("chain disabled by config", "chain_type", chainType)
			continue
		}

		// Check if middleware is in the chain's middleware list
		if len(chainConfig.MiddlewareNames) > 0 {
			found := false
			for _, allowedName := range chainConfig.MiddlewareNames {
				if allowedName == name {
					found = true
					break
				}
			}
			if !found {
				o.logger.Info("middleware not in chain config", "name", name, "chain_type", chainType)
				continue
			}
		}

		filtered = append(filtered, mw)
	}

	return filtered
}

// validateChain validates middleware dependencies and conflicts.
func (o *orchestrator) validateChain(middlewares []core.Middleware) error {
	// Create a set of middleware names for quick lookup
	middlewareSet := make(map[string]bool)
	for _, mw := range middlewares {
		middlewareSet[mw.Name()] = true
	}

	// Check dependencies
	for _, mw := range middlewares {
		name := mw.Name()
		config := o.config.GetMiddlewareConfig(name)

		// Check dependencies
		if deps, ok := config["dependencies"]; ok {
			if depList, ok := deps.([]string); ok {
				for _, dep := range depList {
					if !middlewareSet[dep] {
						return fmt.Errorf("middleware %q requires missing dependency %q", name, dep)
					}
				}
			}
		}

		// Check conflicts
		if confs, ok := config["conflicts"]; ok {
			if confList, ok := confs.([]string); ok {
				for _, conf := range confList {
					if middlewareSet[conf] {
						return fmt.Errorf("middleware %q conflicts with %q", name, conf)
					}
				}
			}
		}
	}

	return nil
}

// validateRegistryDependencies validates dependencies in the registry.
func (o *orchestrator) validateRegistryDependencies() error {
	// This is a simplified implementation - in a real scenario, the registry would have ValidateDependencies
	// For now, we'll do basic validation
	allNames := o.registry.List()

	for _, name := range allNames {
		if mw, ok := o.registry.Get(name); ok {
			config := o.config.GetMiddlewareConfig(mw.Name())

			// Check dependencies
			if deps, ok := config["dependencies"]; ok {
				if depList, ok := deps.([]string); ok {
					for _, dep := range depList {
						if _, exists := o.registry.Get(dep); !exists {
							return fmt.Errorf("middleware %q requires missing dependency %q", name, dep)
						}
					}
				}
			}
		}
	}

	return nil
}

// applyPathSpecificMiddleware adds path-specific middleware to the chain.
func (o *orchestrator) applyPathSpecificMiddleware(baseChain core.Chain, requestPath string) core.Chain {
	// Clone the base chain
	middlewares := baseChain.List()
	pathChain := chain.NewChainImpl(middlewares)

	// Add path-specific middleware based on path patterns
	if strings.HasPrefix(requestPath, "/api/") {
		// Add API-specific middleware
		if apiMw, ok := o.registry.Get("api-logging"); ok {
			pathChain.Insert(0, apiMw)
		}
	} else if strings.HasPrefix(requestPath, "/admin/") {
		// Add admin-specific middleware
		if adminMw, ok := o.registry.Get("admin-auth"); ok {
			pathChain.Insert(0, adminMw)
		}
	} else if strings.HasPrefix(requestPath, "/static/") {
		// Add static-specific middleware
		if staticMw, ok := o.registry.Get("static-cache"); ok {
			pathChain.Insert(0, staticMw)
		}
	}

	return pathChain
}

// filterByPath filters middleware based on path patterns.
func (o *orchestrator) filterByPath(chainObj core.Chain, requestPath string) core.Chain {
	middlewares := chainObj.List()
	var filtered []core.Middleware

	for _, mw := range middlewares {
		config := o.config.GetMiddlewareConfig(mw.Name())

		// Check path inclusion patterns
		if includePaths, ok := config["include_paths"]; ok {
			if paths, ok := includePaths.([]string); ok {
				if !o.matchesAnyPath(requestPath, paths) {
					o.logger.Info("middleware excluded by include_paths", "name", mw.Name(), "path", requestPath)
					continue
				}
			}
		}

		// Check path exclusion patterns
		if excludePaths, ok := config["exclude_paths"]; ok {
			if paths, ok := excludePaths.([]string); ok {
				if o.matchesAnyPath(requestPath, paths) {
					o.logger.Info("middleware excluded by exclude_paths", "name", mw.Name(), "path", requestPath)
					continue
				}
			}
		}

		filtered = append(filtered, mw)
	}

	return chain.NewChainImpl(filtered)
}

// matchesAnyPath checks if the request path matches any of the given patterns.
func (o *orchestrator) matchesAnyPath(requestPath string, patterns []string) bool {
	for _, pattern := range patterns {
		if o.matchesPath(requestPath, pattern) {
			return true
		}
	}
	return false
}

// matchesPath checks if the request path matches the given pattern.
func (o *orchestrator) matchesPath(requestPath string, pattern string) bool {
	// Handle glob patterns
	if strings.Contains(pattern, "*") {
		// Convert glob to regex
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		regexPattern = "^" + regexPattern + "$"
		if matched, _ := regexp.MatchString(regexPattern, requestPath); matched {
			return true
		}
	}

	// Handle exact path matching
	if pattern == requestPath {
		return true
	}

	// Handle prefix matching
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		if strings.HasPrefix(requestPath, prefix) {
			return true
		}
	}

	// Handle path.Match patterns
	if matched, _ := path.Match(pattern, requestPath); matched {
		return true
	}

	return false
}

// sortByPriority sorts middleware by priority (lower number = higher priority).
func (o *orchestrator) sortByPriority(middlewares []core.Middleware) {
	// This is already handled by the registry, but we ensure consistency
	// The registry.GetOrdered method already sorts by priority
}

// getMiddlewareNames extracts middleware names from a slice of middleware.
func (o *orchestrator) getMiddlewareNames(middlewares []core.Middleware) []string {
	names := make([]string, len(middlewares))
	for i, mw := range middlewares {
		names[i] = mw.Name()
	}
	return names
}

// getCategoriesForChainType returns the middleware categories for a given chain type.
func (o *orchestrator) getCategoriesForChainType(chainType core.ChainType) []core.MiddlewareCategory {
	switch chainType {
	case core.ChainTypeDefault:
		return []core.MiddlewareCategory{
			core.MiddlewareCategoryBasic,
			core.MiddlewareCategorySecurity,
			core.MiddlewareCategoryLogging,
		}
	case core.ChainTypeAPI:
		return []core.MiddlewareCategory{
			core.MiddlewareCategoryBasic,
			core.MiddlewareCategorySecurity,
			core.MiddlewareCategoryAuth,
			core.MiddlewareCategoryLogging,
		}
	case core.ChainTypeWeb:
		return []core.MiddlewareCategory{
			core.MiddlewareCategoryBasic,
			core.MiddlewareCategorySecurity,
			core.MiddlewareCategoryAuth,
			core.MiddlewareCategoryLogging,
		}
	case core.ChainTypeAuth:
		return []core.MiddlewareCategory{
			core.MiddlewareCategoryBasic,
			core.MiddlewareCategorySecurity,
			core.MiddlewareCategoryAuth,
		}
	case core.ChainTypeAdmin:
		return []core.MiddlewareCategory{
			core.MiddlewareCategoryBasic,
			core.MiddlewareCategorySecurity,
			core.MiddlewareCategoryAuth,
			core.MiddlewareCategoryLogging,
		}
	case core.ChainTypePublic:
		return []core.MiddlewareCategory{
			core.MiddlewareCategoryBasic,
			core.MiddlewareCategorySecurity,
		}
	case core.ChainTypeStatic:
		return []core.MiddlewareCategory{
			core.MiddlewareCategoryBasic,
		}
	default:
		return []core.MiddlewareCategory{core.MiddlewareCategoryBasic}
	}
}

// getChainDescription returns a description for the given chain type.
func (o *orchestrator) getChainDescription(chainType core.ChainType) string {
	switch chainType {
	case core.ChainTypeDefault:
		return "Default middleware chain for most requests"
	case core.ChainTypeAPI:
		return "Middleware chain for API requests with authentication and logging"
	case core.ChainTypeWeb:
		return "Middleware chain for web page requests with session management"
	case core.ChainTypeAuth:
		return "Middleware chain for authentication endpoints"
	case core.ChainTypeAdmin:
		return "Middleware chain for admin-only endpoints with enhanced security"
	case core.ChainTypePublic:
		return "Middleware chain for public endpoints with basic security"
	case core.ChainTypeStatic:
		return "Middleware chain for static asset requests with caching"
	default:
		return "Unknown middleware chain type"
	}
}
