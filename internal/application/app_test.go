package application_test

import (
	"context"
	"embed"
	"strings"
	"testing"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/application"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	mockconfig "github.com/goformx/goforms/test/mocks/config"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	mockserver "github.com/goformx/goforms/test/mocks/server"
)

// Create a simple embed.FS for testing
var testDistFS = embed.FS{}

// createTestConfig returns a valid *config.Config for testing
func createTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:            "test-app",
			Version:         "0.0.1-test",
			Environment:     "test",
			Debug:           true,
			LogLevel:        "debug",
			URL:             "http://localhost:8080",
			Scheme:          "http",
			Port:            8080,
			Host:            "localhost",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     5 * time.Second,
			RequestTimeout:  5 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		Database: config.DatabaseConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			Name:            "testdb",
			Username:        "testuser",
			Password:        "testpass",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 5 * time.Minute,
			SSLMode:         "disable",
			Logging: config.DatabaseLoggingConfig{
				SlowThreshold:  1 * time.Second,
				Parameterized:  true,
				IgnoreNotFound: true,
				LogLevel:       "silent",
			},
		},
		Security: config.SecurityConfig{
			CSRF: config.CSRFConfig{
				Enabled:        true,
				Secret:         "abcdefghijklmnopqrstuvwxyz123456", // 32 chars
				TokenName:      "_csrf",
				HeaderName:     "X-Csrf-Token",
				TokenLength:    32,
				ContextKey:     "csrf",
				CookieName:     "_csrf",
				CookiePath:     "/",
				CookieHTTPOnly: true,
				CookieSameSite: "Lax",
				CookieMaxAge:   86400,
			},
			CORS: config.CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"http://localhost"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: false,
			},
		},
		// ... add other required config sections with minimal valid values ...
	}
}

func TestNewApplication(t *testing.T) {
	t.Run("creates application successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mocks
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockServer := mockserver.NewMockServerInterface(ctrl)
		mockConfig := mockconfig.NewMockConfigInterface(ctrl)

		// Set up mock expectations
		mockConfig.EXPECT().GetApp().Return(config.AppConfig{
			Name:            "test-app",
			Environment:     "test",
			ShutdownTimeout: 5 * time.Second,
		}).AnyTimes()

		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockServer.EXPECT().Start().Return(nil).AnyTimes()

		testModules := []fx.Option{
			infrastructure.Module, // Only infrastructure, no config
			domain.Module,         // Domain services
			fx.Provide(
				func() server.ServerInterface { return mockServer },
				func() logging.Logger { return mockLogger },
				func() *config.Config { return createTestConfig() },
				func() config.ConfigInterface { return createTestConfig() },
			),
		}

		// Create application with test modules (no config module)
		app := fx.New(testModules...)

		if app == nil {
			t.Fatal("Expected application to be created, got nil")
		}
	})
}

func TestApplicationLifecycle(t *testing.T) {
	t.Run("starts and stops successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mocks
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockServer := mockserver.NewMockServerInterface(ctrl)
		mockConfig := mockconfig.NewMockConfigInterface(ctrl)

		// Set up mock expectations
		mockConfig.EXPECT().GetApp().Return(config.AppConfig{
			Name:            "test-app",
			Environment:     "test",
			ShutdownTimeout: 5 * time.Second,
		}).AnyTimes()

		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockServer.EXPECT().Start().Return(nil).AnyTimes()

		testModules := []fx.Option{
			infrastructure.Module, // Only infrastructure, no config
			domain.Module,         // Domain services
			fx.Provide(
				func() server.ServerInterface { return mockServer },
				func() logging.Logger { return mockLogger },
				func() *config.Config { return createTestConfig() },
				func() config.ConfigInterface { return createTestConfig() },
			),
		}

		// Create application with test modules
		app := fx.New(testModules...)

		// Start the application
		if err := app.Start(context.Background()); err != nil {
			t.Fatalf("Expected application to start successfully, got error: %v", err)
		}

		// Stop the application
		if err := app.Stop(context.Background()); err != nil {
			t.Fatalf("Expected application to stop successfully, got error: %v", err)
		}
	})
}

func TestApplicationWithFxtest(t *testing.T) {
	t.Run("uses fxtest for faster testing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mocks
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockServer := mockserver.NewMockServerInterface(ctrl)
		mockConfig := mockconfig.NewMockConfigInterface(ctrl)

		// Set up mock expectations
		mockConfig.EXPECT().GetApp().Return(config.AppConfig{
			Name:            "test-app",
			Environment:     "test",
			ShutdownTimeout: 5 * time.Second,
		}).AnyTimes()

		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockServer.EXPECT().Start().Return(nil).AnyTimes()

		testModules := []fx.Option{
			infrastructure.Module, // Only infrastructure, no config
			domain.Module,         // Domain services
			fx.Provide(
				func() server.ServerInterface { return mockServer },
				func() logging.Logger { return mockLogger },
				func() *config.Config { return createTestConfig() },
				func() config.ConfigInterface { return createTestConfig() },
			),
		}

		// Use fxtest.New for faster testing
		app := fxtest.New(t, testModules...)

		// Start the application
		app.RequireStart()

		// Stop the application
		app.RequireStop()
	})
}

func TestLifecycleManager(t *testing.T) {
	t.Run("handles startup and shutdown", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mocks
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockServer := mockserver.NewMockServerInterface(ctrl)
		mockConfig := mockconfig.NewMockConfigInterface(ctrl)

		// Set up mock expectations with correct argument counts
		mockConfig.EXPECT().GetApp().Return(config.AppConfig{
			Name:        "test-app",
			Environment: "test",
		}).AnyTimes()

		// Logger.Info calls with correct argument patterns
		mockLogger.EXPECT().Info("starting application", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockLogger.EXPECT().Info("server started successfully").Times(1)
		mockLogger.EXPECT().Info("shutting down application", gomock.Any(), gomock.Any()).Times(1)
		mockServer.EXPECT().Start().Return(nil)

		// Create lifecycle manager
		params := application.LifecycleParams{
			Logger: mockLogger,
			Server: mockServer,
			Config: mockConfig,
		}

		manager := application.NewLifecycleManager(params)

		// Test startup
		ctx := context.Background()

		err := manager.HandleStartup(ctx)
		if err != nil {
			t.Fatalf("Expected startup to succeed, got error: %v", err)
		}

		// Test shutdown
		err = manager.HandleShutdown(ctx)
		if err != nil {
			t.Fatalf("Expected shutdown to succeed, got error: %v", err)
		}
	})
}

func TestLifecycleManager_StartupError(t *testing.T) {
	t.Run("handles server startup error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mocks
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockServer := mockserver.NewMockServerInterface(ctrl)
		mockConfig := mockconfig.NewMockConfigInterface(ctrl)

		// Set up mock expectations
		mockConfig.EXPECT().GetApp().Return(config.AppConfig{
			Name:        "test-app",
			Environment: "test",
		}).AnyTimes()

		mockLogger.EXPECT().Info("starting application", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		mockLogger.EXPECT().Error("server startup failed", gomock.Any()).Times(1)
		mockServer.EXPECT().Start().Return(context.DeadlineExceeded)

		// Create lifecycle manager
		params := application.LifecycleParams{
			Logger: mockLogger,
			Server: mockServer,
			Config: mockConfig,
		}

		manager := application.NewLifecycleManager(params)

		// Test startup with error
		ctx := context.Background()

		err := manager.HandleStartup(ctx)
		if err == nil {
			t.Fatal("Expected startup to fail, but it succeeded")
		}

		expectedError := "server failed to start"
		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestLifecycleManager_ContextCancellation(t *testing.T) {
	t.Run("handles context cancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mocks
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockServer := mockserver.NewMockServerInterface(ctrl)
		mockConfig := mockconfig.NewMockConfigInterface(ctrl)

		// Set up mock expectations
		mockConfig.EXPECT().GetApp().Return(config.AppConfig{
			Name:        "test-app",
			Environment: "test",
		}).AnyTimes()

		mockLogger.EXPECT().Info("starting application", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
		// Server.Start() will block, but context will be canceled
		mockServer.EXPECT().Start().DoAndReturn(func() error {
			// Simulate a blocking server that never returns
			select {}
		})

		// Create lifecycle manager
		params := application.LifecycleParams{
			Logger: mockLogger,
			Server: mockServer,
			Config: mockConfig,
		}

		manager := application.NewLifecycleManager(params)

		// Test startup with canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := manager.HandleStartup(ctx)
		if err == nil {
			t.Fatal("Expected startup to fail due to context cancellation, but it succeeded")
		}

		expectedError := "application startup canceled"
		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestNewLifecycleManager(t *testing.T) {
	t.Run("creates lifecycle manager with correct dependencies", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mocks
		mockLogger := mocklogging.NewMockLogger(ctrl)
		mockServer := mockserver.NewMockServerInterface(ctrl)
		mockConfig := mockconfig.NewMockConfigInterface(ctrl)

		// Create lifecycle params
		params := application.LifecycleParams{
			Logger: mockLogger,
			Server: mockServer,
			Config: mockConfig,
		}

		// Create lifecycle manager
		manager := application.NewLifecycleManager(params)

		// Verify manager was created
		if manager == nil {
			t.Fatal("Expected lifecycle manager to be created, got nil")
		}
	})
}
