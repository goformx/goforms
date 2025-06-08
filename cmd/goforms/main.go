// Package main is the entry point for the GoFormX application.
// It sets up the application using dependency injection (via fx),
// configures the server, and manages the application lifecycle.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/labstack/echo/v4"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 30 * time.Second
)

// appParams is the application parameters
type appParams struct {
	fx.In

	// Core dependencies
	Lifecycle         fx.Lifecycle
	Echo              *echo.Echo
	Server            *server.Server
	Logger            logging.Logger
	Handlers          []web.Handler `group:"handlers"`
	MiddlewareManager *appmiddleware.Manager
}

func setupLogger(cfg *config.Config) (logging.Logger, error) {
	return logging.NewFactory(logging.FactoryConfig{
		AppName:     cfg.App.Name,
		Version:     cfg.App.Version,
		Environment: cfg.App.Env,
		Fields:      map[string]any{},
	}).CreateLogger()
}

// setupEcho configures and starts the server
func setupEcho(params appParams) error {
	// Register all handlers
	registerHandlers(params.Logger, params.Echo, params.Handlers)

	// Setup middleware
	params.MiddlewareManager.Setup(params.Echo)

	// Start the server
	if err := params.Server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// registerHandlers registers all application handlers
func registerHandlers(logger logging.Logger, e *echo.Echo, handlers []web.Handler) {
	logger.Info("registering handlers")

	for _, h := range handlers {
		h.Register(e)
	}

	logger.Info("handlers registered successfully")
}

// createLifecycleHooks creates the application lifecycle hooks
func createLifecycleHooks(params appParams) fx.Hook {
	return fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("running setupEcho in lifecycle hooks")
			return setupEcho(params)
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("shutting down application in lifecycle hooks")
			return nil
		},
	}
}

func main() {
	// Create the application with all dependencies
	app := fx.New(
		// Core infrastructure
		fx.Provide(
			setupLogger,
			config.New,
			echo.New,
		),
		// Domain services
		domain.Module,
		// Infrastructure and handlers
		infrastructure.Module,
		// Application lifecycle
		fx.Invoke(func(params appParams) {
			params.Lifecycle.Append(createLifecycleHooks(params))
		}),
	)

	// Create a context that will be canceled on interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the application
	if err := app.Start(ctx); err != nil {
		stop() // Ensure signal handler is stopped
		log.Printf("Failed to start application: %v", err)

		return // Use return instead of os.Exit to allow deferred functions to run
	}

	// Wait for interrupt signal
	<-ctx.Done()

	// Create a new context for shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	// Stop the application
	if err := app.Stop(shutdownCtx); err != nil {
		stop() // Ensure signal handler is stopped
		log.Printf("Failed to stop application: %v", err)

		return // Use return instead of os.Exit to allow deferred functions to run
	}
}
