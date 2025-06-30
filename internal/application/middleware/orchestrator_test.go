package middleware_test

import (
	"context"
	"testing"

	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockMiddleware implements core.Middleware for testing
type mockMiddleware struct {
	name     string
	priority int
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
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{
		middlewares: make(map[string]core.Middleware),
	}
}

func (m *mockRegistry) Register(name string, middleware core.Middleware) error {
	args := m.Called(name, middleware)
	m.middlewares[name] = middleware
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

// mockConfig implements MiddlewareConfig for testing
type mockConfig struct {
	mock.Mock
	enabledMiddleware map[string]bool
	middlewareConfig  map[string]map[string]interface{}
	chainConfigs      map[core.ChainType]ChainConfig
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		enabledMiddleware: make(map[string]bool),
		middlewareConfig:  make(map[string]map[string]interface{}),
		chainConfigs:      make(map[core.ChainType]ChainConfig),
	}
}

func (m *mockConfig) IsMiddlewareEnabled(name string) bool {
	args := m.Called(name)
	if enabled, exists := m.enabledMiddleware[name]; exists {
		return enabled
	}
	return args.Bool(0)
}

func (m *mockConfig) GetMiddlewareConfig(name string) map[string]interface{} {
	args := m.Called(name)
	if config, exists := m.middlewareConfig[name]; exists {
		return config
	}
	return args.Get(0).(map[string]interface{})
}

func (m *mockConfig) GetChainConfig(chainType core.ChainType) ChainConfig {
	args := m.Called(chainType)
	if config, exists := m.chainConfigs[chainType]; exists {
		return config
	}
	return args.Get(0).(ChainConfig)
}

// mockLogger implements core.Logger for testing
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func TestOrchestrator_CreateChain(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	corsMw := &mockMiddleware{name: "cors", priority: 10}
	authMw := &mockMiddleware{name: "auth", priority: 20}
	loggingMw := &mockMiddleware{name: "logging", priority: 30}

	registry.middlewares["cors"] = corsMw
	registry.middlewares["auth"] = authMw
	registry.middlewares["logging"] = loggingMw

	// Setup mock config
	config.enabledMiddleware["cors"] = true
	config.enabledMiddleware["auth"] = true
	config.enabledMiddleware["logging"] = true

	config.middlewareConfig["cors"] = map[string]interface{}{
		"category": core.MiddlewareCategoryBasic,
	}
	config.middlewareConfig["auth"] = map[string]interface{}{
		"category": core.MiddlewareCategoryAuth,
	}
	config.middlewareConfig["logging"] = map[string]interface{}{
		"category": core.MiddlewareCategoryLogging,
	}

	config.chainConfigs[core.ChainTypeAPI] = ChainConfig{
		Enabled: true,
	}

	// Setup expectations
	registry.On("List").Return([]string{"cors", "auth", "logging"})
	registry.On("Get", "cors").Return(corsMw, true)
	registry.On("Get", "auth").Return(authMw, true)
	registry.On("Get", "logging").Return(loggingMw, true)

	config.On("IsMiddlewareEnabled", "cors").Return(true)
	config.On("IsMiddlewareEnabled", "auth").Return(true)
	config.On("IsMiddlewareEnabled", "logging").Return(true)
	config.On("GetMiddlewareConfig", "cors").Return(map[string]interface{}{"category": core.MiddlewareCategoryBasic})
	config.On("GetMiddlewareConfig", "auth").Return(map[string]interface{}{"category": core.MiddlewareCategoryAuth})
	config.On("GetMiddlewareConfig", "logging").Return(map[string]interface{}{"category": core.MiddlewareCategoryLogging})
	config.On("GetChainConfig", core.ChainTypeAPI).Return(ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := NewOrchestrator(registry, config, logger)

	// Test creating a chain
	chain, err := orchestrator.CreateChain(core.ChainTypeAPI)
	assert.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, 3, chain.Length())

	// Verify middleware order (should be by priority)
	middlewares := chain.List()
	assert.Equal(t, "cors", middlewares[0].Name())
	assert.Equal(t, "auth", middlewares[1].Name())
	assert.Equal(t, "logging", middlewares[2].Name())

	// Verify mock expectations
	registry.AssertExpectations(t)
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_BuildChainForPath(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	basicMw := &mockMiddleware{name: "basic", priority: 10}
	apiMw := &mockMiddleware{name: "api-logging", priority: 5}

	registry.middlewares["basic"] = basicMw
	registry.middlewares["api-logging"] = apiMw

	// Setup mock config
	config.enabledMiddleware["basic"] = true
	config.enabledMiddleware["api-logging"] = true

	config.middlewareConfig["basic"] = map[string]interface{}{
		"category": core.MiddlewareCategoryBasic,
	}
	config.middlewareConfig["api-logging"] = map[string]interface{}{
		"category": core.MiddlewareCategoryLogging,
	}

	config.chainConfigs[core.ChainTypeAPI] = ChainConfig{
		Enabled: true,
	}

	// Setup expectations
	registry.On("List").Return([]string{"basic", "api-logging"})
	registry.On("Get", "basic").Return(basicMw, true)
	registry.On("Get", "api-logging").Return(apiMw, true)

	config.On("IsMiddlewareEnabled", "basic").Return(true)
	config.On("IsMiddlewareEnabled", "api-logging").Return(true)
	config.On("GetMiddlewareConfig", "basic").Return(map[string]interface{}{"category": core.MiddlewareCategoryBasic})
	config.On("GetMiddlewareConfig", "api-logging").Return(map[string]interface{}{"category": core.MiddlewareCategoryLogging})
	config.On("GetChainConfig", core.ChainTypeAPI).Return(ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := NewOrchestrator(registry, config, logger)

	// Test building chain for API path
	chain, err := orchestrator.BuildChainForPath(core.ChainTypeAPI, "/api/users")
	assert.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, 2, chain.Length())

	// Verify path-specific middleware was added
	middlewares := chain.List()
	assert.Equal(t, "api-logging", middlewares[0].Name()) // Should be first due to path-specific insertion
	assert.Equal(t, "basic", middlewares[1].Name())

	// Verify mock expectations
	registry.AssertExpectations(t)
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_GetChainInfo(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock config
	config.chainConfigs[core.ChainTypeWeb] = ChainConfig{
		Enabled:         true,
		MiddlewareNames: []string{"cors", "auth", "logging"},
	}

	// Setup expectations
	registry.On("List").Return([]string{})
	config.On("GetChainConfig", core.ChainTypeWeb).Return(ChainConfig{
		Enabled:         true,
		MiddlewareNames: []string{"cors", "auth", "logging"},
	})

	// Create orchestrator
	orchestrator := NewOrchestrator(registry, config, logger)

	// Test getting chain info
	info := orchestrator.GetChainInfo(core.ChainTypeWeb)
	assert.Equal(t, core.ChainTypeWeb, info.Type)
	assert.Equal(t, "web", info.Name)
	assert.Equal(t, "Middleware chain for web page requests with session management", info.Description)
	assert.True(t, info.Enabled)
	assert.Contains(t, info.Categories, core.MiddlewareCategoryBasic)
	assert.Contains(t, info.Categories, core.MiddlewareCategorySecurity)
	assert.Contains(t, info.Categories, core.MiddlewareCategoryAuth)
	assert.Contains(t, info.Categories, core.MiddlewareCategoryLogging)

	// Verify mock expectations
	registry.AssertExpectations(t)
	config.AssertExpectations(t)
}

func TestOrchestrator_ChainManagement(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Create orchestrator
	orchestrator := NewOrchestrator(registry, config, logger)

	// Test chain registration and retrieval
	mockChain := &mockChain{}

	err := orchestrator.RegisterChain("test-chain", mockChain)
	assert.NoError(t, err)

	// Test duplicate registration
	err = orchestrator.RegisterChain("test-chain", mockChain)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test retrieval
	retrievedChain, exists := orchestrator.GetChain("test-chain")
	assert.True(t, exists)
	assert.Equal(t, mockChain, retrievedChain)

	// Test non-existent chain
	_, exists = orchestrator.GetChain("non-existent")
	assert.False(t, exists)

	// Test listing chains
	chains := orchestrator.ListChains()
	assert.Contains(t, chains, "test-chain")

	// Test removing chain
	removed := orchestrator.RemoveChain("test-chain")
	assert.True(t, removed)

	// Test removing non-existent chain
	removed = orchestrator.RemoveChain("non-existent")
	assert.False(t, removed)

	// Verify chain was removed
	_, exists = orchestrator.GetChain("test-chain")
	assert.False(t, exists)
}

func TestOrchestrator_CacheManagement(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware
	basicMw := &mockMiddleware{name: "basic", priority: 10}
	registry.middlewares["basic"] = basicMw

	// Setup mock config
	config.enabledMiddleware["basic"] = true
	config.middlewareConfig["basic"] = map[string]interface{}{
		"category": core.MiddlewareCategoryBasic,
	}
	config.chainConfigs[core.ChainTypeDefault] = ChainConfig{Enabled: true}

	// Setup expectations
	registry.On("List").Return([]string{"basic"})
	registry.On("Get", "basic").Return(basicMw, true)
	config.On("IsMiddlewareEnabled", "basic").Return(true)
	config.On("GetMiddlewareConfig", "basic").Return(map[string]interface{}{"category": core.MiddlewareCategoryBasic})
	config.On("GetChainConfig", core.ChainTypeDefault).Return(ChainConfig{Enabled: true})
	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := NewOrchestrator(registry, config, logger)

	// Test caching
	chain1, err := orchestrator.GetChainForPath(core.ChainTypeDefault, "/test")
	assert.NoError(t, err)
	assert.NotNil(t, chain1)

	// Get the same chain again - should use cache
	chain2, err := orchestrator.GetChainForPath(core.ChainTypeDefault, "/test")
	assert.NoError(t, err)
	assert.Equal(t, chain1, chain2)

	// Clear cache
	orchestrator.ClearCache()

	// Get chain again - should rebuild
	chain3, err := orchestrator.GetChainForPath(core.ChainTypeDefault, "/test")
	assert.NoError(t, err)
	assert.NotNil(t, chain3)
	// Note: chain3 might not be equal to chain1 due to different instances

	// Verify mock expectations
	registry.AssertExpectations(t)
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_PathMatching(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	orchestrator := NewOrchestrator(registry, config, logger)

	// Test exact path matching
	assert.True(t, orchestrator.(*orchestrator).matchesPath("/api/users", "/api/users"))

	// Test prefix matching
	assert.True(t, orchestrator.(*orchestrator).matchesPath("/api/users/123", "/api/*"))
	assert.True(t, orchestrator.(*orchestrator).matchesPath("/api/users", "/api/*"))

	// Test glob pattern matching
	assert.True(t, orchestrator.(*orchestrator).matchesPath("/api/users/123", "/api/users/*"))
	assert.False(t, orchestrator.(*orchestrator).matchesPath("/api/posts/123", "/api/users/*"))

	// Test multiple patterns
	patterns := []string{"/api/*", "/admin/*"}
	assert.True(t, orchestrator.(*orchestrator).matchesAnyPath("/api/users", patterns))
	assert.True(t, orchestrator.(*orchestrator).matchesAnyPath("/admin/dashboard", patterns))
	assert.False(t, orchestrator.(*orchestrator).matchesAnyPath("/public/about", patterns))
}

// mockChain implements core.Chain for testing
type mockChain struct {
	mock.Mock
}

func (m *mockChain) Process(ctx context.Context, req core.Request) core.Response {
	args := m.Called(ctx, req)
	return args.Get(0).(core.Response)
}

func (m *mockChain) Add(middleware ...core.Middleware) core.Chain {
	args := m.Called(middleware)
	return args.Get(0).(core.Chain)
}

func (m *mockChain) Insert(position int, middleware ...core.Middleware) core.Chain {
	args := m.Called(position, middleware)
	return args.Get(0).(core.Chain)
}

func (m *mockChain) Remove(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *mockChain) Get(name string) core.Middleware {
	args := m.Called(name)
	return args.Get(0).(core.Middleware)
}

func (m *mockChain) List() []core.Middleware {
	args := m.Called()
	return args.Get(0).([]core.Middleware)
}

func (m *mockChain) Clear() core.Chain {
	args := m.Called()
	return args.Get(0).(core.Chain)
}

func (m *mockChain) Length() int {
	args := m.Called()
	return args.Int(0)
}
