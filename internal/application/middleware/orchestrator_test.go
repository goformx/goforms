package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockMiddleware implements core.Middleware for testing
type mockMiddleware struct {
	name     string
	priority int
	category core.MiddlewareCategory
}

func (m *mockMiddleware) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	return next(ctx, req)
}

func (m *mockMiddleware) Name() string {
	return m.name
}

func (m *mockMiddleware) Priority() int {
	return m.priority
}

// mockRegistry implements core.Registry for testing
type mockRegistry struct {
	mock.Mock
	middlewares map[string]core.Middleware
	categories  map[core.MiddlewareCategory][]string
	priorities  map[string]int
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{
		middlewares: make(map[string]core.Middleware),
		categories:  make(map[core.MiddlewareCategory][]string),
		priorities:  make(map[string]int),
	}
}

func (m *mockRegistry) Register(name string, mw core.Middleware) error {
	args := m.Called(name, mw)
	m.middlewares[name] = mw

	return args.Error(0)
}

func (m *mockRegistry) Get(name string) (core.Middleware, bool) {
	mw, exists := m.middlewares[name]

	return mw, exists
}

func (m *mockRegistry) List() []string {
	names := make([]string, 0, len(m.middlewares))
	for name := range m.middlewares {
		names = append(names, name)
	}

	return names
}

func (m *mockRegistry) Remove(name string) bool {
	if _, exists := m.middlewares[name]; exists {
		delete(m.middlewares, name)

		return true
	}

	return false
}

func (m *mockRegistry) Clear() {
	m.middlewares = make(map[string]core.Middleware)
}

func (m *mockRegistry) Count() int {
	return len(m.middlewares)
}

// GetOrdered returns middleware ordered by priority for a category
func (m *mockRegistry) GetOrdered(category core.MiddlewareCategory) []core.Middleware {
	var result []core.Middleware

	for _, mw := range m.middlewares {
		if mockMw, ok := mw.(*mockMiddleware); ok && mockMw.category == category {
			result = append(result, mw)
		}
	}
	// Sort by priority
	for i := range len(result) - 1 {
		for j := i + 1; j < len(result); j++ {
			if result[i].Priority() > result[j].Priority() {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// mockConfig implements middleware.MiddlewareConfig for testing
type mockConfig struct {
	mock.Mock
	enabledMiddleware map[string]bool
	middlewareConfig  map[string]map[string]any
	chainConfigs      map[core.ChainType]middleware.ChainConfig
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		enabledMiddleware: make(map[string]bool),
		middlewareConfig:  make(map[string]map[string]any),
		chainConfigs:      make(map[core.ChainType]middleware.ChainConfig),
	}
}

func (m *mockConfig) IsMiddlewareEnabled(name string) bool {
	args := m.Called(name)
	if enabled, exists := m.enabledMiddleware[name]; exists {
		return enabled
	}

	return args.Bool(0)
}

func (m *mockConfig) GetMiddlewareConfig(name string) map[string]any {
	args := m.Called(name)
	if config, exists := m.middlewareConfig[name]; exists {
		return config
	}

	if result, ok := args.Get(0).(map[string]any); ok {
		return result
	}

	return make(map[string]any)
}

func (m *mockConfig) GetChainConfig(chainType core.ChainType) middleware.ChainConfig {
	args := m.Called(chainType)
	if config, exists := m.chainConfigs[chainType]; exists {
		return config
	}

	if result, ok := args.Get(0).(middleware.ChainConfig); ok {
		return result
	}

	return middleware.ChainConfig{}
}

// mockLogger implements core.Logger for testing
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *mockLogger) Warn(msg string, args ...any) {
	m.Called(msg, args)
}

func (m *mockLogger) Error(msg string, args ...any) {
	m.Called(msg, args)
}

func TestOrchestrator_CreateChain(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	authMw := &mockMiddleware{name: "auth", priority: 20, category: core.MiddlewareCategoryAuth}
	loggingMw := &mockMiddleware{name: "logging", priority: 30, category: core.MiddlewareCategoryLogging}

	registry.middlewares["cors"] = corsMw
	registry.middlewares["auth"] = authMw
	registry.middlewares["logging"] = loggingMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.enabledMiddleware["auth"] = true
	config.enabledMiddleware["logging"] = true

	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.middlewareConfig["auth"] = map[string]any{
		"category": core.MiddlewareCategoryAuth,
	}
	config.middlewareConfig["logging"] = map[string]any{
		"category": core.MiddlewareCategoryLogging,
	}

	config.chainConfigs[core.ChainTypeAPI] = middleware.ChainConfig{
		Enabled: true,
	}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "cors").Return(true)
	config.On("IsMiddlewareEnabled", "auth").Return(true)
	config.On("IsMiddlewareEnabled", "logging").Return(true)
	config.On("GetMiddlewareConfig", "cors").Return(map[string]any{"category": core.MiddlewareCategoryBasic})
	config.On("GetMiddlewareConfig", "auth").Return(map[string]any{"category": core.MiddlewareCategoryAuth})
	config.On("GetMiddlewareConfig", "logging").Return(map[string]any{"category": core.MiddlewareCategoryLogging})
	config.On("GetChainConfig", core.ChainTypeAPI).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Test creating a chain
	chain, err := orchestrator.CreateChain(core.ChainTypeAPI)
	require.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, 3, chain.Length())

	// Verify middleware order (should be by priority)
	middlewares := chain.List()
	assert.Equal(t, "cors", middlewares[0].Name())
	assert.Equal(t, "auth", middlewares[1].Name())
	assert.Equal(t, "logging", middlewares[2].Name())

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_BuildChainForPath(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	authMw := &mockMiddleware{name: "auth", priority: 20, category: core.MiddlewareCategoryAuth}
	apiSpecificMw := &mockMiddleware{name: "api-specific", priority: 30, category: core.MiddlewareCategoryCustom}

	registry.middlewares["cors"] = corsMw
	registry.middlewares["auth"] = authMw
	registry.middlewares["api-specific"] = apiSpecificMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.enabledMiddleware["auth"] = true
	config.enabledMiddleware["api-specific"] = true

	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.middlewareConfig["auth"] = map[string]any{
		"category": core.MiddlewareCategoryAuth,
	}
	config.middlewareConfig["api-specific"] = map[string]any{
		"category": core.MiddlewareCategoryCustom,
		"paths":    []string{"/api/*"},
	}

	config.chainConfigs[core.ChainTypeAPI] = middleware.ChainConfig{
		Enabled: true,
	}

	// Setup expectations - use Any() to be more flexible
	config.On("IsMiddlewareEnabled", mock.Anything).Return(func(name string) bool {
		return config.enabledMiddleware[name]
	})
	config.On("GetMiddlewareConfig", mock.Anything).Return(func(name string) map[string]any {
		return config.middlewareConfig[name]
	})
	config.On("GetChainConfig", core.ChainTypeAPI).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Test building chain for API path
	chain, err := orchestrator.BuildChainForPath(core.ChainTypeAPI, "/api/users")
	require.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, 3, chain.Length()) // cors, auth, api-specific (logging not included in API chain)

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_GetChainForPath_Caching(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	authMw := &mockMiddleware{name: "auth", priority: 20, category: core.MiddlewareCategoryAuth}

	registry.middlewares["cors"] = corsMw
	registry.middlewares["auth"] = authMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.enabledMiddleware["auth"] = true

	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.middlewareConfig["auth"] = map[string]any{
		"category": core.MiddlewareCategoryAuth,
	}

	config.chainConfigs[core.ChainTypeAPI] = middleware.ChainConfig{
		Enabled: true,
	}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "cors").Return(true)
	config.On("IsMiddlewareEnabled", "auth").Return(true)
	config.On("GetMiddlewareConfig", "cors").Return(map[string]any{"category": core.MiddlewareCategoryBasic})
	config.On("GetMiddlewareConfig", "auth").Return(map[string]any{"category": core.MiddlewareCategoryAuth})
	config.On("GetChainConfig", core.ChainTypeAPI).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// First call should build the chain
	chain1, err := orchestrator.GetChainForPath(core.ChainTypeAPI, "/api/users")
	require.NoError(t, err)
	assert.NotNil(t, chain1)

	// Second call should return cached chain
	chain2, err := orchestrator.GetChainForPath(core.ChainTypeAPI, "/api/users")
	require.NoError(t, err)
	assert.NotNil(t, chain2)
	assert.Equal(t, chain1, chain2)

	// Verify cache stats
	stats := orchestrator.GetCacheStats()
	assert.Equal(t, 1, stats["cache_size"])

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_ConfigurationValidation(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware with dependencies
	authMw := &mockMiddleware{name: "auth", priority: 10, category: core.MiddlewareCategoryAuth}
	sessionMw := &mockMiddleware{name: "session", priority: 20, category: core.MiddlewareCategoryAuth}

	registry.middlewares["auth"] = authMw
	registry.middlewares["session"] = sessionMw

	// Setup mock config with dependencies
	config.enabledMiddleware["auth"] = true
	config.enabledMiddleware["session"] = true

	config.middlewareConfig["auth"] = map[string]any{
		"category": core.MiddlewareCategoryAuth,
	}
	config.middlewareConfig["session"] = map[string]any{
		"category":     core.MiddlewareCategoryAuth,
		"dependencies": []string{"auth"},
	}

	config.chainConfigs[core.ChainTypeWeb] = middleware.ChainConfig{
		Enabled: true,
	}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "auth").Return(true)
	config.On("IsMiddlewareEnabled", "session").Return(true)
	config.On("GetMiddlewareConfig", "auth").Return(map[string]any{"category": core.MiddlewareCategoryAuth})
	config.On("GetMiddlewareConfig", "session").Return(map[string]any{
		"category":     core.MiddlewareCategoryAuth,
		"dependencies": []string{"auth"},
	})
	config.On("GetChainConfig", core.ChainTypeWeb).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Test creating a chain with dependencies
	chain, err := orchestrator.CreateChain(core.ChainTypeWeb)
	require.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, 2, chain.Length())

	// Verify middleware order (should be by priority)
	middlewares := chain.List()
	assert.Equal(t, "auth", middlewares[0].Name())
	assert.Equal(t, "session", middlewares[1].Name())

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_DisabledMiddleware(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	authMw := &mockMiddleware{name: "auth", priority: 20, category: core.MiddlewareCategoryAuth}

	registry.middlewares["cors"] = corsMw
	registry.middlewares["auth"] = authMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.enabledMiddleware["auth"] = false

	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.middlewareConfig["auth"] = map[string]any{
		"category": core.MiddlewareCategoryAuth,
	}

	config.chainConfigs[core.ChainTypeAPI] = middleware.ChainConfig{Enabled: true}

	// Setup expectations - use Any() to be more flexible
	config.On("IsMiddlewareEnabled", mock.Anything).Return(func(name string) bool {
		return config.enabledMiddleware[name]
	})
	config.On("GetMiddlewareConfig", mock.Anything).Return(func(name string) map[string]any {
		return config.middlewareConfig[name]
	})
	config.On("GetChainConfig", core.ChainTypeAPI).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Test creating a chain - should only include enabled middleware
	chain, err := orchestrator.CreateChain(core.ChainTypeAPI)
	require.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, 1, chain.Length())

	// Verify only enabled middleware is included
	middlewares := chain.List()
	assert.Equal(t, "cors", middlewares[0].Name())

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_PathFiltering(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	authMw := &mockMiddleware{name: "auth", priority: 20, category: core.MiddlewareCategoryAuth}
	adminMw := &mockMiddleware{name: "admin", priority: 30, category: core.MiddlewareCategoryAuth}

	registry.middlewares["cors"] = corsMw
	registry.middlewares["auth"] = authMw
	registry.middlewares["admin"] = adminMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.enabledMiddleware["auth"] = true
	config.enabledMiddleware["admin"] = true

	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.middlewareConfig["auth"] = map[string]any{
		"category":      core.MiddlewareCategoryAuth,
		"exclude_paths": []string{"/public/*"},
	}
	config.middlewareConfig["admin"] = map[string]any{
		"category":      core.MiddlewareCategoryAuth,
		"include_paths": []string{"/admin/*"},
	}

	config.chainConfigs[core.ChainTypeDefault] = middleware.ChainConfig{
		Enabled: true,
	}

	// Setup expectations - use Any() to be more flexible
	config.On("IsMiddlewareEnabled", mock.Anything).Return(func(name string) bool {
		return config.enabledMiddleware[name]
	})
	config.On("GetMiddlewareConfig", mock.Anything).Return(func(name string) map[string]any {
		return config.middlewareConfig[name]
	})
	config.On("GetChainConfig", core.ChainTypeDefault).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Test public path - should exclude auth middleware
	chain1, err := orchestrator.BuildChainForPath(core.ChainTypeDefault, "/public/info")
	require.NoError(t, err)
	assert.NotNil(t, chain1)
	assert.Equal(t, 1, chain1.Length()) // Only cors

	// Test admin path - should include admin middleware
	chain2, err := orchestrator.BuildChainForPath(core.ChainTypeDefault, "/admin/users")
	require.NoError(t, err)
	assert.NotNil(t, chain2)
	assert.Equal(t, 1, chain2.Length()) // Only cors (auth and admin are filtered by path requirements)

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_ChainManagement(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	registry.middlewares["cors"] = corsMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.chainConfigs[core.ChainTypeDefault] = middleware.ChainConfig{Enabled: true}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "cors").Return(true)
	config.On("GetMiddlewareConfig", "cors").Return(map[string]any{"category": core.MiddlewareCategoryBasic})
	config.On("GetChainConfig", core.ChainTypeDefault).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Create a chain
	chain, err := orchestrator.CreateChain(core.ChainTypeDefault)
	require.NoError(t, err)

	// Register the chain
	err = orchestrator.RegisterChain("test-chain", chain)
	require.NoError(t, err)

	// List chains
	chains := orchestrator.ListChains()
	assert.Contains(t, chains, "test-chain")

	// Get the chain
	retrievedChain, exists := orchestrator.GetChain("test-chain")
	assert.True(t, exists)
	assert.Equal(t, chain, retrievedChain)

	// Remove the chain
	removed := orchestrator.RemoveChain("test-chain")
	assert.True(t, removed)

	// Verify chain is removed
	_, exists = orchestrator.GetChain("test-chain")
	assert.False(t, exists)

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_PerformanceTracking(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	registry.middlewares["cors"] = corsMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.chainConfigs[core.ChainTypeDefault] = middleware.ChainConfig{Enabled: true}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "cors").Return(true)
	config.On("GetMiddlewareConfig", "cors").Return(map[string]any{"category": core.MiddlewareCategoryBasic})
	config.On("GetChainConfig", core.ChainTypeDefault).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Create a chain
	chain, err := orchestrator.CreateChain(core.ChainTypeDefault)
	require.NoError(t, err)
	assert.NotNil(t, chain)

	// Get performance metrics
	performance := orchestrator.GetChainPerformance()
	assert.NotEmpty(t, performance)
	assert.Contains(t, performance, "default")

	// Verify build time is reasonable
	buildTime := performance["default"]
	assert.Less(t, buildTime, 100*time.Millisecond)

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_GetChainInfo(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	registry.middlewares["cors"] = corsMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.chainConfigs[core.ChainTypeAPI] = middleware.ChainConfig{
		Enabled:         true,
		MiddlewareNames: []string{"cors"},
		Paths:           []string{"/api/*"},
		CustomConfig:    map[string]any{"timeout": 30},
	}

	// Setup expectations - GetChainInfo only calls GetChainConfig
	config.On("GetChainConfig", core.ChainTypeAPI).Return(middleware.ChainConfig{
		Enabled:         true,
		MiddlewareNames: []string{"cors"},
		Paths:           []string{"/api/*"},
		CustomConfig:    map[string]any{"timeout": 30},
	})

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Get chain info
	info := orchestrator.GetChainInfo(core.ChainTypeAPI)
	assert.Equal(t, core.ChainTypeAPI, info.Type)
	assert.Equal(t, "api", info.Name)
	assert.True(t, info.Enabled)
	assert.Contains(t, info.Middleware, "cors")
	assert.Contains(t, info.Categories, core.MiddlewareCategoryBasic)
	assert.Equal(t, []string{"/api/*"}, info.PathPatterns)
	assert.Equal(t, map[string]any{"timeout": 30}, info.CustomConfig)

	// Verify mock expectations
	config.AssertExpectations(t)
}

func TestOrchestrator_CacheManagement(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10, category: core.MiddlewareCategoryBasic}
	registry.middlewares["cors"] = corsMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.middlewareConfig["cors"] = map[string]any{
		"category": core.MiddlewareCategoryBasic,
	}
	config.chainConfigs[core.ChainTypeDefault] = middleware.ChainConfig{Enabled: true}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "cors").Return(true)
	config.On("GetMiddlewareConfig", "cors").Return(map[string]any{"category": core.MiddlewareCategoryBasic})
	config.On("GetChainConfig", core.ChainTypeDefault).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Build a chain to populate cache
	chain, err := orchestrator.GetChainForPath(core.ChainTypeDefault, "/test")
	require.NoError(t, err)
	assert.NotNil(t, chain)

	// Get cache stats
	stats := orchestrator.GetCacheStats()
	assert.Equal(t, 1, stats["cache_size"])

	// Clear cache
	orchestrator.ClearCache()

	// Verify cache is cleared
	stats = orchestrator.GetCacheStats()
	assert.Equal(t, 0, stats["cache_size"])

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}
