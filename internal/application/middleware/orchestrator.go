// Package middleware provides middleware orchestration and management.
package middleware

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/chain"
	"github.com/goformx/goforms/internal/application/middleware/core"
)

// orchestrator implements the core.Orchestrator interface.
type orchestrator struct {
	registry core.Registry
	config   MiddlewareConfig
	logger   core.Logger
	cache    map[string]core.Chain
	cacheMu  sync.RWMutex
	chains   map[string]core.Chain
	chainsMu sync.RWMutex
	// Performance tracking
	buildTimes map[string]time.Duration
	buildMu    sync.RWMutex
}

// NewOrchestrator creates a new middleware orchestrator.
func NewOrchestrator(registry core.Registry, config MiddlewareConfig, logger core.Logger) core.Orchestrator {
	return &orchestrator{
		registry:   registry,
		config:     config,
		logger:     logger,
		cache:      make(map[string]core.Chain),
		chains:     make(map[string]core.Chain),
		buildTimes: make(map[string]time.Duration),
	}
}

// CreateChain creates a new middleware chain with the specified type.
func (o *orchestrator) CreateChain(chainType core.ChainType) (core.Chain, error) {
	start := time.Now()

	defer func() {
		o.buildMu.Lock()
		o.buildTimes[chainType.String()] = time.Since(start)
		o.buildMu.Unlock()
	}()

	// Get middleware list from registry based on chain type
	middlewares := o.getOrderedMiddleware(chainType)

	// Filter based on configuration
	activeMiddlewares := o.filterByConfig(middlewares, chainType)

	// Validate dependencies and conflicts
	if validateErr := o.validateChain(activeMiddlewares); validateErr != nil {
		return nil, fmt.Errorf("chain validation failed for %s: %w", chainType, validateErr)
	}

	// Build chain using chain builder
	chainImpl := chain.NewChainImpl(activeMiddlewares)

	o.logger.Info("built middleware chain",
		"chain_type", chainType,
		"middleware_count", len(activeMiddlewares),
		"middleware_names", o.getMiddlewareNames(activeMiddlewares),
		"build_time", time.Since(start))

	return chainImpl, nil
}

// BuildChain is an alias for CreateChain for backward compatibility.
func (o *orchestrator) BuildChain(chainType core.ChainType) (core.Chain, error) {
	return o.CreateChain(chainType)
}

// GetChain retrieves a pre-configured chain by name.
func (o *orchestrator) GetChain(name string) (core.Chain, bool) {
	o.chainsMu.RLock()
	defer o.chainsMu.RUnlock()

	chainObj, ok := o.chains[name]

	return chainObj, ok
}

// RegisterChain registers a named chain for later retrieval.
func (o *orchestrator) RegisterChain(name string, chainObj core.Chain) error {
	o.chainsMu.Lock()
	defer o.chainsMu.Unlock()

	if _, exists := o.chains[name]; exists {
		return fmt.Errorf("chain with name %q already exists", name)
	}

	o.chains[name] = chainObj
	o.logger.Info("registered named chain", "name", name, "middleware_count", chainObj.Length())

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

	sort.Strings(names)

	return names
}

// RemoveChain removes a chain by name.
func (o *orchestrator) RemoveChain(name string) bool {
	o.chainsMu.Lock()
	defer o.chainsMu.Unlock()

	if _, exists := o.chains[name]; exists {
		delete(o.chains, name)
		o.logger.Info("removed named chain", "name", name)

		return true
	}

	return false
}

// BuildChainForPath creates a middleware chain for a specific path and chain type.
func (o *orchestrator) BuildChainForPath(chainType core.ChainType, requestPath string) (core.Chain, error) {
	start := time.Now()

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
		"middleware_count", finalChain.Length(),
		"build_time", time.Since(start))

	return finalChain, nil
}

// GetChainForPath returns a cached chain for a path or builds a new one.
func (o *orchestrator) GetChainForPath(chainType core.ChainType, requestPath string) (core.Chain, error) {
	cacheKey := fmt.Sprintf("path:%s:%s", chainType, requestPath)

	// Check cache first
	o.cacheMu.RLock()

	if cached, exists := o.cache[cacheKey]; exists {
		o.cacheMu.RUnlock()
		o.logger.Info("returned cached chain", "cache_key", cacheKey)

		return cached, nil
	}

	o.cacheMu.RUnlock()

	// Build new chain
	builtChain, err := o.BuildChainForPath(chainType, requestPath)
	if err != nil {
		return nil, err
	}

	// Cache the result
	o.cacheMu.Lock()
	o.cache[cacheKey] = builtChain
	o.cacheMu.Unlock()

	o.logger.Info("cached new chain", "cache_key", cacheKey)

	return builtChain, nil
}

// ClearCache clears the chain cache.
func (o *orchestrator) ClearCache() {
	o.cacheMu.Lock()
	defer o.cacheMu.Unlock()

	cacheSize := len(o.cache)
	o.cache = make(map[string]core.Chain)
	o.logger.Info("cleared middleware chain cache", "cleared_entries", cacheSize)
}

// GetCacheStats returns cache statistics.
func (o *orchestrator) GetCacheStats() map[string]any {
	o.cacheMu.RLock()
	defer o.cacheMu.RUnlock()

	o.buildMu.RLock()
	defer o.buildMu.RUnlock()

	return map[string]any{
		"cache_size":        len(o.cache),
		"build_times":       o.buildTimes,
		"registered_chains": len(o.chains),
	}
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

	o.logger.Info("configuration validation completed successfully")

	return nil
}

// GetChainInfo returns information about a chain type.
func (o *orchestrator) GetChainInfo(chainType core.ChainType) core.ChainInfo {
	chainConfig := o.config.GetChainConfig(chainType)

	// Get middleware for this chain type
	middlewares := o.getOrderedMiddleware(chainType)

	middlewareNames := o.getMiddlewareNames(middlewares)

	// Determine categories based on chain type
	categories := o.getCategoriesForChainType(chainType)

	return core.ChainInfo{
		Type:         chainType,
		Name:         chainType.String(),
		Description:  o.getChainDescription(chainType),
		Categories:   categories,
		Middleware:   middlewareNames,
		Enabled:      chainConfig.Enabled,
		PathPatterns: chainConfig.Paths,
		CustomConfig: chainConfig.CustomConfig,
	}
}

// GetChainPerformance returns performance metrics for chain building.
func (o *orchestrator) GetChainPerformance() map[string]time.Duration {
	o.buildMu.RLock()
	defer o.buildMu.RUnlock()

	result := make(map[string]time.Duration)
	for chainType, duration := range o.buildTimes {
		result[chainType] = duration
	}

	return result
}

// getOrderedMiddleware returns middleware ordered by priority for the given chain type.
func (o *orchestrator) getOrderedMiddleware(chainType core.ChainType) []core.Middleware {
	// Get categories for this chain type
	categories := o.getCategoriesForChainType(chainType)

	// Preallocate with estimated capacity (estimatedMiddlewarePerCategory is defined as a constant)
	const estimatedMiddlewarePerCategory = 4
	middlewares := make([]core.Middleware, 0, len(categories)*estimatedMiddlewarePerCategory)

	// Collect middleware from all relevant categories
	for _, category := range categories {
		categoryMiddleware := o.getOrderedByCategory(category)
		middlewares = append(middlewares, categoryMiddleware...)
	}

	// Sort by priority (registry already does this, but ensure consistency)
	o.sortByPriority(middlewares)

	return middlewares
}

// getOrderedByCategory returns middleware ordered by priority for a specific category.
func (o *orchestrator) getOrderedByCategory(category core.MiddlewareCategory) []core.Middleware {
	// Use registry's GetOrdered method if available
	if registry, ok := o.registry.(interface {
		GetOrdered(core.MiddlewareCategory) []core.Middleware
	}); ok {
		return registry.GetOrdered(category)
	}

	// Fallback implementation
	allNames := o.registry.List()

	var categoryMiddleware []core.Middleware

	for _, name := range allNames {
		if mw, ok := o.registry.Get(name); ok {
			if o.belongsToCategory(mw, name, category) {
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
	filtered := make([]core.Middleware, 0, len(middlewares))

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

	// Check dependencies and conflicts
	for _, mw := range middlewares {
		name := mw.Name()
		config := o.config.GetMiddlewareConfig(name)

		if err := o.validateDependencies(name, config, middlewareSet); err != nil {
			return err
		}

		if err := o.validateConflicts(name, config, middlewareSet); err != nil {
			return err
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
			if err := o.validateRegistryDependency(name, config); err != nil {
				return err
			}
		}
	}

	return nil
}

// applyPathSpecificMiddleware adds path-specific middleware to the chain.
func (o *orchestrator) applyPathSpecificMiddleware(baseChain core.Chain, requestPath string) core.Chain {
	// Check if there are any path-specific middleware configurations
	allNames := o.registry.List()

	for _, name := range allNames {
		if mw, ok := o.registry.Get(name); ok {
			config := o.config.GetMiddlewareConfig(name)
			if o.shouldAddPathSpecificMiddleware(config, requestPath) {
				baseChain.Add(mw)
				o.logger.Info("added path-specific middleware", "name", name, "path", requestPath)
			}
		}
	}

	return baseChain
}

// filterByPath filters middleware based on path patterns.
func (o *orchestrator) filterByPath(chainObj core.Chain, requestPath string) core.Chain {
	middlewares := chainObj.List()

	filteredMiddlewares := make([]core.Middleware, 0, len(middlewares))

	for _, mw := range middlewares {
		config := o.config.GetMiddlewareConfig(mw.Name())

		if o.shouldExcludeByPath(mw.Name(), config, requestPath) {
			o.logger.Info("excluded middleware by path", "name", mw.Name(), "path", requestPath)

			continue
		}

		if o.shouldExcludeByPathRequirement(config, requestPath) {
			o.logger.Info("excluded middleware by path requirement", "name", mw.Name(), "path", requestPath)

			continue
		}

		filteredMiddlewares = append(filteredMiddlewares, mw)
	}

	return chain.NewChainImpl(filteredMiddlewares)
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
func (o *orchestrator) matchesPath(requestPath, pattern string) bool {
	// Handle glob patterns
	if strings.Contains(pattern, "*") {
		// Convert glob to regex
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")

		regexPattern = "^" + regexPattern + "$"
		if matched, err := regexp.MatchString(regexPattern, requestPath); err == nil && matched {
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
	if matched, err := path.Match(pattern, requestPath); err == nil && matched {
		return true
	}

	return false
}

// sortByPriority sorts middleware by priority (lower number = higher priority).
func (o *orchestrator) sortByPriority(middlewares []core.Middleware) {
	sort.SliceStable(middlewares, func(i, j int) bool {
		return middlewares[i].Priority() < middlewares[j].Priority()
	})
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

// belongsToCategory checks if middleware belongs to the specified category.
func (o *orchestrator) belongsToCategory(_ core.Middleware, name string, category core.MiddlewareCategory) bool {
	config := o.config.GetMiddlewareConfig(name)
	if catVal, ok := config["category"]; ok {
		if c, catOk := catVal.(core.MiddlewareCategory); catOk && c == category {
			return true
		}
	} else if category == core.MiddlewareCategoryBasic {
		// Default to basic category if not specified
		return true
	}

	return false
}

// validateDependencies checks if middleware dependencies are satisfied.
func (o *orchestrator) validateDependencies(
	name string,
	config map[string]any,
	middlewareSet map[string]bool,
) error {
	if deps, exists := config["dependencies"]; exists {
		if depList, isStringSlice := deps.([]string); isStringSlice {
			for _, dep := range depList {
				if !middlewareSet[dep] {
					return fmt.Errorf("middleware %q requires missing dependency %q", name, dep)
				}
			}
		}
	}

	return nil
}

// validateConflicts checks if middleware conflicts are resolved.
func (o *orchestrator) validateConflicts(
	name string,
	config map[string]any,
	middlewareSet map[string]bool,
) error {
	if confs, exists := config["conflicts"]; exists {
		if confList, isStringSlice := confs.([]string); isStringSlice {
			for _, conf := range confList {
				if middlewareSet[conf] {
					return fmt.Errorf("middleware %q conflicts with %q", name, conf)
				}
			}
		}
	}

	return nil
}

// validateRegistryDependency checks if a middleware's dependencies exist in the registry.
func (o *orchestrator) validateRegistryDependency(name string, config map[string]any) error {
	if deps, exists := config["dependencies"]; exists {
		if depList, isStringSlice := deps.([]string); isStringSlice {
			for _, dep := range depList {
				if _, depExists := o.registry.Get(dep); !depExists {
					return fmt.Errorf("middleware %q requires missing dependency %q", name, dep)
				}
			}
		}
	}

	return nil
}

// shouldAddPathSpecificMiddleware determines if middleware should be added for the given path.
func (o *orchestrator) shouldAddPathSpecificMiddleware(
	config map[string]any,
	requestPath string,
) bool {
	if paths, exists := config["paths"]; exists {
		if pathList, isStringSlice := paths.([]string); isStringSlice {
			return o.matchesAnyPath(requestPath, pathList)
		}
	}

	return false
}

// shouldExcludeByPath checks if middleware should be excluded based on exclude_paths.
func (o *orchestrator) shouldExcludeByPath(_ string, config map[string]any, requestPath string) bool {
	if excludePaths, exists := config["exclude_paths"]; exists {
		if pathList, isStringSlice := excludePaths.([]string); isStringSlice {
			return o.matchesAnyPath(requestPath, pathList)
		}
	}

	return false
}

// shouldExcludeByPathRequirement checks if middleware should be excluded based on include_paths.
func (o *orchestrator) shouldExcludeByPathRequirement(
	config map[string]any,
	requestPath string,
) bool {
	if includePaths, exists := config["include_paths"]; exists {
		if pathList, isStringSlice := includePaths.([]string); isStringSlice {
			return !o.matchesAnyPath(requestPath, pathList)
		}
	}

	return false
}
