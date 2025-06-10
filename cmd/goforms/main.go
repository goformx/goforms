// Package main is the entry point for the GoFormX application.
// It sets up the application using dependency injection (via fx),
// configures the server, and manages the application lifecycle.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
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

// setupEcho configures and starts the server
func setupEcho(params appParams) error {
	// Register middleware
	params.MiddlewareManager.Setup(params.Echo)

	// Register handlers
	for _, handler := range params.Handlers {
		if handler == nil {
			return errors.New("nil handler encountered during registration")
		}
		handler.Register(params.Echo)
	}

	return nil
}

// createFallbackLogger creates a basic logger for startup errors
func createFallbackLogger() logging.Logger {
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "goforms",
		Version:     "unknown",
		Environment: "development",
	})
	logger, err := factory.CreateLogger()
	if err != nil {
		// If we can't create a logger, we can't do much else
		os.Exit(1)
	}
	return logger
}

func main() {
	// Create the application with all dependencies
	app := fx.New(
		// Core modules
		infrastructure.Module, // Provides core infrastructure (config, logger, db, etc.)
		domain.Module,         // Provides domain services and interfaces
		// Application lifecycle
		fx.Invoke(func(params appParams) {
			params.Lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					params.Logger.Info("Starting application...")
					if setupErr := setupEcho(params); setupErr != nil {
						return fmt.Errorf("failed to setup echo: %w", setupErr)
					}
					params.Logger.Info("Application started successfully")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					params.Logger.Info("Shutting down application...")
					return nil
				},
			})
		}),
	)

	// Create a context that will be canceled on interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the application
	if startErr := app.Start(ctx); startErr != nil {
		stop() // Ensure signal handler is stopped
		fallbackLogger := createFallbackLogger()
		fallbackLogger.Fatal("Failed to start application",
			"error", startErr,
		)
	}

	// Wait for interrupt signal
	<-ctx.Done()

	// Create a new context for shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	// Stop the application
	if stopErr := app.Stop(shutdownCtx); stopErr != nil {
		fallbackLogger := createFallbackLogger()
		fallbackLogger.Error("Failed to stop application gracefully",
			"error", stopErr,
		)
		os.Exit(1)
	}
}
