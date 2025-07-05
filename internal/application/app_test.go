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
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	mockconfig "github.com/goformx/goforms/test/mocks/config"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	mockserver "github.com/goformx/goforms/test/mocks/server"
)

// Create a simple embed.FS for testing
var testDistFS = embed.FS{}

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

		mockLogger.EXPECT().Info("starting application", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockServer.EXPECT().Start().Return(nil).AnyTimes()

		testModules := []fx.Option{
			fx.Provide(
				func() server.ServerInterface { return mockServer },
				func() logging.Logger { return mockLogger },
				func() config.ConfigInterface { return mockConfig },
			),
		}

		// Create application with test modules
		app := application.NewApplication(testDistFS, testModules...)

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

		mockLogger.EXPECT().Info("starting application", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockServer.EXPECT().Start().Return(nil).AnyTimes()

		testModules := []fx.Option{
			fx.Provide(
				func() server.ServerInterface { return mockServer },
				func() logging.Logger { return mockLogger },
				func() config.ConfigInterface { return mockConfig },
			),
		}

		// Create application with test modules
		app := application.NewApplication(testDistFS, testModules...)

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

		mockLogger.EXPECT().Info("starting application", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockServer.EXPECT().Start().Return(nil).AnyTimes()

		testModules := []fx.Option{
			fx.Provide(
				func() server.ServerInterface { return mockServer },
				func() logging.Logger { return mockLogger },
				func() config.ConfigInterface { return mockConfig },
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
