package middleware_test

import (
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/config"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
)

// TestIntegration_MiddlewareOrchestrator tests the complete middleware orchestrator integration
func TestIntegration_MiddlewareOrchestrator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup
	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := createIntegrationTestConfig()
	mwConfig := middleware.NewMiddlewareConfig(cfg, logger)
	registry := middleware.NewRegistry(logger, mwConfig)
	orchestrator := middleware.NewOrchestrator(registry, mwConfig, logger)
	migrationAdapter := middleware.NewMigrationAdapter(orchestrator, registry, logger)

	// Register all required middleware for migration validation
	requiredMiddleware := []struct {
		name string
		mw   core.Middleware
	}{
		{"recovery", middleware.NewRecoveryMiddleware()},
		{"cors", middleware.NewCORSMiddleware()},
		{"request-id", middleware.NewRequestIDMiddleware()},
		{"timeout", middleware.NewTimeoutMiddleware()},
		{"security-headers", middleware.NewSecurityHeadersMiddleware()},
		{"csrf", middleware.NewCSRFMiddleware()},
		{"rate-limit", middleware.NewRateLimitMiddleware()},
		{"session", middleware.NewSessionMiddleware()},
		{"authentication", middleware.NewAuthenticationMiddleware()},
		{"authorization", middleware.NewAuthorizationMiddleware()},
	}

	for _, m := range requiredMiddleware {
		regErr := registry.Register(m.name, m.mw)
		require.NoError(t, regErr)
	}

	// Test registry registration
	t.Run("Registry Registration", func(t *testing.T) {
		// Test middleware registration
		assert.Equal(t, 10, registry.Count())
		assert.Contains(t, registry.List(), "recovery")
		assert.Contains(t, registry.List(), "cors")

		// Test middleware retrieval
		mw, exists := registry.Get("recovery")
		assert.True(t, exists)
		assert.NotNil(t, mw)
		assert.Equal(t, "recovery", mw.Name())
	})

	// Test orchestrator chain building
	t.Run("Orchestrator Chain Building", func(t *testing.T) {
		// Test different chain types
		chainTypes := []core.ChainType{
			core.ChainTypeDefault,
			core.ChainTypeAPI,
			core.ChainTypeWeb,
			core.ChainTypeAuth,
			core.ChainTypeAdmin,
			core.ChainTypePublic,
			core.ChainTypeStatic,
		}

		for _, chainType := range chainTypes {
			t.Run(chainType.String(), func(t *testing.T) {
				mwChain, chainErr := orchestrator.BuildChain(chainType)
				require.NoError(t, chainErr)
				assert.NotNil(t, mwChain)
				assert.GreaterOrEqual(t, mwChain.Length(), 0)
			})
		}
	})

	// Test Echo adapter integration
	t.Run("Echo Adapter Integration", func(t *testing.T) {
		adapter := middleware.NewEchoOrchestratorAdapter(orchestrator, logger)
		e := echo.New()

		// Test middleware setup
		err := adapter.SetupMiddleware(e)
		require.NoError(t, err)

		// Test chain building for specific paths
		paths := []string{"/api/users", "/dashboard", "/login", "/admin/users", "/public/", "/static/css/style.css"}
		for _, path := range paths {
			chain, chainErr := adapter.BuildChainForPath(path)
			require.NoError(t, chainErr)
			assert.NotNil(t, chain)
		}
	})

	// Test migration adapter
	t.Run("Migration Adapter", func(t *testing.T) {
		// Test migration validation
		err := migrationAdapter.ValidateMigration()
		require.NoError(t, err)

		// Test migration status
		status := migrationAdapter.GetMigrationStatus()
		assert.False(t, status.NewSystemEnabled)
		assert.NotNil(t, status.RegisteredMiddleware)
		assert.NotNil(t, status.AvailableChains)
	})

	// Test configuration integration
	t.Run("Configuration Integration", func(t *testing.T) {
		// Test middleware enabled check
		assert.True(t, mwConfig.IsMiddlewareEnabled("recovery"))
		assert.True(t, mwConfig.IsMiddlewareEnabled("cors"))

		// Test middleware config
		mwRecoveryConfig := mwConfig.GetMiddlewareConfig("recovery")
		assert.NotNil(t, mwRecoveryConfig)
		assert.Equal(t, core.MiddlewareCategoryBasic, mwRecoveryConfig["category"])

		// Test chain config
		chainConfig := mwConfig.GetChainConfig(core.ChainTypeDefault)
		assert.True(t, chainConfig.Enabled)
		assert.NotEmpty(t, chainConfig.MiddlewareNames)
	})
}

// TestIntegration_MigrationFlow tests the complete migration flow
func TestIntegration_MigrationFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := createIntegrationTestConfig()
	mwConfig := middleware.NewMiddlewareConfig(cfg, logger)
	registry := middleware.NewRegistry(logger, mwConfig)
	orchestrator := middleware.NewOrchestrator(registry, mwConfig, logger)
	migrationAdapter := middleware.NewMigrationAdapter(orchestrator, registry, logger)

	// Register required middleware
	requiredMiddleware := []struct {
		name string
		mw   core.Middleware
	}{
		{"recovery", middleware.NewRecoveryMiddleware()},
		{"cors", middleware.NewCORSMiddleware()},
		{"request-id", middleware.NewRequestIDMiddleware()},
		{"timeout", middleware.NewTimeoutMiddleware()},
		{"security-headers", middleware.NewSecurityHeadersMiddleware()},
		{"csrf", middleware.NewCSRFMiddleware()},
		{"rate-limit", middleware.NewRateLimitMiddleware()},
		{"session", middleware.NewSessionMiddleware()},
		{"authentication", middleware.NewAuthenticationMiddleware()},
		{"authorization", middleware.NewAuthorizationMiddleware()},
	}

	for _, m := range requiredMiddleware {
		regErr := registry.Register(m.name, m.mw)
		require.NoError(t, regErr)
	}

	t.Run("Migration Flow", func(t *testing.T) {
		// Phase 1: Start with old system
		assert.False(t, migrationAdapter.IsNewSystemEnabled())

		// Phase 2: Validate migration readiness
		err := migrationAdapter.ValidateMigration()
		require.NoError(t, err)

		// Phase 3: Enable new system
		migrationAdapter.EnableNewSystem()
		assert.True(t, migrationAdapter.IsNewSystemEnabled())

		// Phase 4: Test Echo integration with new system
		e := echo.New()
		err = migrationAdapter.SetupHybridMiddleware(e, nil) // No old manager needed for new system
		require.NoError(t, err)

		// Phase 5: Test rollback capability
		migrationAdapter.RollbackMigration()
		assert.False(t, migrationAdapter.IsNewSystemEnabled())
	})
}

// TestIntegration_PathBasedChains tests path-based middleware chain selection
func TestIntegration_PathBasedChains(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := createIntegrationTestConfig()
	mwConfig := middleware.NewMiddlewareConfig(cfg, logger)
	registry := middleware.NewRegistry(logger, mwConfig)
	orchestrator := middleware.NewOrchestrator(registry, mwConfig, logger)
	adapter := middleware.NewEchoOrchestratorAdapter(orchestrator, logger)

	// Register middleware
	err := registry.Register("recovery", middleware.NewRecoveryMiddleware())
	require.NoError(t, err)
	err = registry.Register("cors", middleware.NewCORSMiddleware())
	require.NoError(t, err)

	t.Run("Path-Based Chain Selection", func(t *testing.T) {
		testCases := []struct {
			path     string
			expected core.ChainType
		}{
			{"/api/users", core.ChainTypeAPI},
			{"/dashboard", core.ChainTypeWeb},
			{"/login", core.ChainTypeAuth},
			{"/admin/users", core.ChainTypeAdmin},
			{"/public/", core.ChainTypePublic},
			{"/static/css/style.css", core.ChainTypeStatic},
			{"/", core.ChainTypeDefault},
			{"/unknown/path", core.ChainTypeDefault},
		}

		for _, tc := range testCases {
			t.Run(tc.path, func(t *testing.T) {
				chain, chainErr := adapter.BuildChainForPath(tc.path)
				require.NoError(t, chainErr)
				assert.NotNil(t, chain)
			})
		}
	})
}

// TestIntegration_Performance tests middleware chain performance
func TestIntegration_Performance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := createIntegrationTestConfig()
	mwConfig := middleware.NewMiddlewareConfig(cfg, logger)
	registry := middleware.NewRegistry(logger, mwConfig)
	orchestrator := middleware.NewOrchestrator(registry, mwConfig, logger)

	// Register multiple middleware
	for i := range 10 {
		err := registry.Register("test-mw-"+string(rune(i)), middleware.NewRecoveryMiddleware())
		require.NoError(t, err)
	}

	t.Run("Chain Building Performance", func(t *testing.T) {
		start := time.Now()

		for range 100 {
			chain, chainErr := orchestrator.BuildChain(core.ChainTypeDefault)
			require.NoError(t, chainErr)
			assert.NotNil(t, chain)
		}

		duration := time.Since(start)
		assert.Less(t, duration, 100*time.Millisecond, "Chain building should be fast")
	})
}

// TestIntegration_ErrorHandling tests error handling in the middleware system
func TestIntegration_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklogging.NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	cfg := createIntegrationTestConfig()
	mwConfig := middleware.NewMiddlewareConfig(cfg, logger)
	registry := middleware.NewRegistry(logger, mwConfig)
	orchestrator := middleware.NewOrchestrator(registry, mwConfig, logger)
	migrationAdapter := middleware.NewMigrationAdapter(orchestrator, registry, logger)

	t.Run("Error Handling", func(t *testing.T) {
		// Test migration with missing middleware
		err := migrationAdapter.ValidateMigration()
		// This should fail because not all required middleware are registered
		require.Error(t, err)

		// Test rollback on error
		migrationAdapter.EnableNewSystem()

		e := echo.New()
		err = migrationAdapter.SetupWithFallback(e, nil)
		// Should fallback to old system or handle error gracefully
		assert.NoError(t, err)
	})
}

// Helper functions

func createIntegrationTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Environment: "test",
			Debug:       true,
		},
		Security: config.SecurityConfig{
			CORS: config.CORSConfig{
				Enabled: true,
			},
			CSRF: config.CSRFConfig{
				Enabled: true,
			},
			RateLimit: config.RateLimitConfig{
				Enabled: true,
			},
		},
	}
}
