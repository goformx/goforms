package middleware_test

import (
	"context"
	"testing"

	"github.com/goformx/goforms/internal/application/middleware"
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

// mockConfig implements middleware.MiddlewareConfig for testing
type mockConfig struct {
	mock.Mock
	enabledMiddleware map[string]bool
	middlewareConfig  map[string]map[string]interface{}
	chainConfigs      map[core.ChainType]middleware.ChainConfig
}

func newMockConfig() *mockConfig {
	return &mockConfig{
		enabledMiddleware: make(map[string]bool),
		middlewareConfig:  make(map[string]map[string]interface{}),
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

func (m *mockConfig) GetMiddlewareConfig(name string) map[string]interface{} {
	args := m.Called(name)
	if config, exists := m.middlewareConfig[name]; exists {
		return config
	}
	return args.Get(0).(map[string]interface{})
}

func (m *mockConfig) GetChainConfig(chainType core.ChainType) middleware.ChainConfig {
	args := m.Called(chainType)
	if config, exists := m.chainConfigs[chainType]; exists {
		return config
	}
	return args.Get(0).(middleware.ChainConfig)
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

	config.chainConfigs[core.ChainTypeAPI] = middleware.ChainConfig{
		Enabled: true,
	}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "cors").Return(true)
	config.On("IsMiddlewareEnabled", "auth").Return(true)
	config.On("IsMiddlewareEnabled", "logging").Return(true)
	config.On("GetMiddlewareConfig", "cors").Return(map[string]interface{}{"category": core.MiddlewareCategoryBasic})
	config.On("GetMiddlewareConfig", "auth").Return(map[string]interface{}{"category": core.MiddlewareCategoryAuth})
	config.On("GetMiddlewareConfig", "logging").Return(map[string]interface{}{"category": core.MiddlewareCategoryLogging})
	config.On("GetChainConfig", core.ChainTypeAPI).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

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
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestOrchestrator_ChainManagement(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

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

func TestOrchestrator_ConfigurationValidation(t *testing.T) {
	registry := newMockRegistry()
	config := newMockConfig()
	logger := &mockLogger{}

	// Setup mock middleware with dependencies
	authMw := &mockMiddleware{name: "auth", priority: 10}
	sessionMw := &mockMiddleware{name: "session", priority: 20}

	registry.middlewares["auth"] = authMw
	registry.middlewares["session"] = sessionMw

	// Setup mock config with dependencies
	config.enabledMiddleware["auth"] = true
	config.enabledMiddleware["session"] = true

	config.middlewareConfig["auth"] = map[string]interface{}{
		"category": core.MiddlewareCategoryAuth,
	}
	config.middlewareConfig["session"] = map[string]interface{}{
		"category":     core.MiddlewareCategoryAuth,
		"dependencies": []string{"auth"},
	}

	config.chainConfigs[core.ChainTypeWeb] = middleware.ChainConfig{
		Enabled: true,
	}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "auth").Return(true)
	config.On("IsMiddlewareEnabled", "session").Return(true)
	config.On("GetMiddlewareConfig", "auth").Return(map[string]interface{}{"category": core.MiddlewareCategoryAuth})
	config.On("GetMiddlewareConfig", "session").Return(map[string]interface{}{
		"category":     core.MiddlewareCategoryAuth,
		"dependencies": []string{"auth"},
	})
	config.On("GetChainConfig", core.ChainTypeWeb).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Test creating a chain with dependencies
	chain, err := orchestrator.CreateChain(core.ChainTypeWeb)
	assert.NoError(t, err)
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
	corsMw := &mockMiddleware{name: "cors", priority: 10}
	authMw := &mockMiddleware{name: "auth", priority: 20}

	registry.middlewares["cors"] = corsMw
	registry.middlewares["auth"] = authMw

	// Setup mock config - disable auth middleware
	config.enabledMiddleware["cors"] = true
	config.enabledMiddleware["auth"] = false

	config.middlewareConfig["cors"] = map[string]interface{}{
		"category": core.MiddlewareCategoryBasic,
	}
	config.middlewareConfig["auth"] = map[string]interface{}{
		"category": core.MiddlewareCategoryAuth,
	}

	config.chainConfigs[core.ChainTypeAPI] = middleware.ChainConfig{
		Enabled: true,
	}

	// Setup expectations
	config.On("IsMiddlewareEnabled", "cors").Return(true)
	config.On("IsMiddlewareEnabled", "auth").Return(false)
	config.On("GetMiddlewareConfig", "cors").Return(map[string]interface{}{"category": core.MiddlewareCategoryBasic})
	config.On("GetMiddlewareConfig", "auth").Return(map[string]interface{}{"category": core.MiddlewareCategoryAuth})
	config.On("GetChainConfig", core.ChainTypeAPI).Return(middleware.ChainConfig{Enabled: true})

	logger.On("Info", mock.Anything, mock.Anything).Return()

	// Create orchestrator
	orchestrator := middleware.NewOrchestrator(registry, config, logger)

	// Test creating a chain - should only include enabled middleware
	chain, err := orchestrator.CreateChain(core.ChainTypeAPI)
	assert.NoError(t, err)
	assert.NotNil(t, chain)
	assert.Equal(t, 1, chain.Length())

	// Verify only enabled middleware is included
	middlewares := chain.List()
	assert.Equal(t, "cors", middlewares[0].Name())

	// Verify mock expectations
	config.AssertExpectations(t)
	logger.AssertExpectations(t)
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
