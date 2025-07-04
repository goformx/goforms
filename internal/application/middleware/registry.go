package middleware

import (
	"fmt"
	"sort"
	"sync"

	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

type registry struct {
	mu           sync.RWMutex
	middlewares  map[string]core.Middleware
	categories   map[core.MiddlewareCategory][]string
	priorities   map[string]int
	dependencies map[string][]string
	conflicts    map[string][]string
	logger       logging.Logger
	config       MiddlewareConfig
}

func NewRegistry(logger logging.Logger, config MiddlewareConfig) core.Registry {
	return &registry{
		middlewares:  make(map[string]core.Middleware),
		categories:   make(map[core.MiddlewareCategory][]string),
		priorities:   make(map[string]int),
		dependencies: make(map[string][]string),
		conflicts:    make(map[string][]string),
		logger:       logger,
		config:       config,
	}
}

func (r *registry) Register(name string, mw core.Middleware) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.middlewares[name]; exists {
		return fmt.Errorf("middleware %q already registered", name)
	}

	if !r.config.IsMiddlewareEnabled(name) {
		return nil
	}

	r.middlewares[name] = mw
	r.setupMiddlewareMetadata(name)

	return nil
}

// setupMiddlewareMetadata extracts and sets up middleware metadata from config
func (r *registry) setupMiddlewareMetadata(name string) {
	config := r.config.GetMiddlewareConfig(name)

	// Set category
	cat := r.extractCategory(config)
	r.categories[cat] = append(r.categories[cat], name)

	// Set priority
	prio := r.extractPriority(config)
	r.priorities[name] = prio

	// Set dependencies
	if deps := r.extractDependencies(config); deps != nil {
		r.dependencies[name] = deps
	}

	// Set conflicts
	if confs := r.extractConflicts(config); confs != nil {
		r.conflicts[name] = confs
	}
}

// extractCategory extracts the category from middleware config
func (r *registry) extractCategory(config map[string]any) core.MiddlewareCategory {
	if cat, ok := config["category"].(core.MiddlewareCategory); ok {
		return cat
	}

	return core.MiddlewareCategoryBasic
}

// extractPriority extracts the priority from middleware config
func (r *registry) extractPriority(config map[string]any) int {
	if prio, ok := config["priority"].(int); ok {
		return prio
	}

	return 50 // Default priority
}

// extractDependencies extracts dependencies from middleware config
func (r *registry) extractDependencies(config map[string]any) []string {
	if deps, ok := config["dependencies"].([]string); ok {
		return deps
	}

	return nil
}

// extractConflicts extracts conflicts from middleware config
func (r *registry) extractConflicts(config map[string]any) []string {
	if confs, ok := config["conflicts"].([]string); ok {
		return confs
	}

	return nil
}

func (r *registry) Get(name string) (core.Middleware, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mw, exists := r.middlewares[name]

	return mw, exists
}

func (r *registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.middlewares))
	for name := range r.middlewares {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

func (r *registry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.middlewares[name]; !exists {
		return false
	}

	delete(r.middlewares, name)

	// Remove from categories
	for cat, names := range r.categories {
		for i, n := range names {
			if n == name {
				r.categories[cat] = append(names[:i], names[i+1:]...)

				break
			}
		}
	}

	// Remove from other maps
	delete(r.priorities, name)
	delete(r.dependencies, name)
	delete(r.conflicts, name)

	return true
}

func (r *registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.middlewares = make(map[string]core.Middleware)
	r.categories = make(map[core.MiddlewareCategory][]string)
	r.priorities = make(map[string]int)
	r.dependencies = make(map[string][]string)
	r.conflicts = make(map[string][]string)
}

func (r *registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.middlewares)
}

// GetOrdered returns middleware names in priority order, filtered by category and config
func (r *registry) GetOrdered(category core.MiddlewareCategory) []core.Middleware {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := r.categories[category]
	// Filter by config
	filtered := make([]string, 0, len(names))
	for _, name := range names {
		if r.config.IsMiddlewareEnabled(name) {
			filtered = append(filtered, name)
		}
	}
	// Sort by priority (lower number = higher priority)
	sort.SliceStable(filtered, func(i, j int) bool {
		return r.priorities[filtered[i]] < r.priorities[filtered[j]]
	})

	result := make([]core.Middleware, 0, len(filtered))
	for _, name := range filtered {
		if mw, ok := r.middlewares[name]; ok {
			result = append(result, mw)
		}
	}

	return result
}

// ValidateDependencies checks for missing dependencies and conflicts
func (r *registry) ValidateDependencies() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, deps := range r.dependencies {
		for _, dep := range deps {
			if _, ok := r.middlewares[dep]; !ok {
				return fmt.Errorf("middleware %q requires missing dependency %q", name, dep)
			}
		}
	}

	for name, confs := range r.conflicts {
		for _, conf := range confs {
			if _, ok := r.middlewares[conf]; ok {
				return fmt.Errorf("middleware %q conflicts with %q", name, conf)
			}
		}
	}

	return nil
}
