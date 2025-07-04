// Package application provides application setup and lifecycle management.
package application

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
)

// ApplicationSetup handles application initialization and setup
type ApplicationSetup struct {
	logger       logging.Logger
	echo         *echo.Echo
	orchestrator *appmiddleware.EchoOrchestratorAdapter
}

// NewApplicationSetup creates a new ApplicationSetup instance
func NewApplicationSetup(
	logger logging.Logger,
	echoInstance *echo.Echo,
	orchestrator *appmiddleware.EchoOrchestratorAdapter,
) *ApplicationSetup {
	return &ApplicationSetup{
		logger:       logger,
		echo:         echoInstance,
		orchestrator: orchestrator,
	}
}

// Setup initializes the application by setting up middleware
func (s *ApplicationSetup) Setup() error {
	s.logger.Info("setting up application middleware")

	if err := s.orchestrator.SetupMiddleware(s.echo); err != nil {
		return fmt.Errorf("failed to setup middleware: %w", err)
	}

	s.logger.Info("application middleware setup completed")

	return nil
}

// LifecycleParams contains dependencies for application lifecycle management
type LifecycleParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    logging.Logger
	Server    server.ServerInterface
	Config    config.ConfigInterface
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(params LifecycleParams) *LifecycleManager {
	return &LifecycleManager{
		lifecycle: params.Lifecycle,
		logger:    params.Logger,
		server:    params.Server,
		config:    params.Config,
	}
}

// LifecycleManager handles application lifecycle events
type LifecycleManager struct {
	lifecycle fx.Lifecycle
	logger    logging.Logger
	server    server.ServerInterface
	config    config.ConfigInterface
}

// SetupLifecycle configures the application lifecycle hooks
func (lm *LifecycleManager) SetupLifecycle() {
	lm.lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return lm.HandleStartup(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return lm.HandleShutdown(ctx)
		},
	})
}

// HandleStartup manages application startup (exported for testing)
func (lm *LifecycleManager) HandleStartup(ctx context.Context) error {
	lm.logger.Info("starting application",
		"app", lm.config.GetApp().Name,
		"environment", lm.config.GetApp().Environment,
	)

	// Create channels for server startup coordination
	serverReady := make(chan struct{})
	serverError := make(chan error, 1)

	// Start the server in a goroutine with proper error handling
	go func() {
		if err := lm.server.Start(); err != nil {
			lm.logger.Error("server startup failed", "error", err)

			serverError <- err

			return
		}

		close(serverReady)
	}()

	// Wait for server to be ready or fail
	select {
	case err := <-serverError:
		return fmt.Errorf("server failed to start: %w", err)
	case <-serverReady:
		lm.logger.Info("server started successfully")

		return nil
	case <-ctx.Done():
		return fmt.Errorf("application startup canceled: %w", ctx.Err())
	}
}

// HandleShutdown manages application shutdown (exported for testing)
func (lm *LifecycleManager) HandleShutdown(ctx context.Context) error {
	lm.logger.Info("shutting down application",
		"app", lm.config.GetApp().Name,
	)

	return nil
}
