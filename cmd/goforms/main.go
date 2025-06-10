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
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/labstack/echo/v4"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 30 * time.Second
)

// appParams groups all dependencies injected via fx.
type appParams struct {
	fx.In
	Lifecycle         fx.Lifecycle
	Echo              *echo.Echo
	Server            *server.Server
	Logger            logging.Logger
	Handlers          []web.Handler `group:"handlers"`
	MiddlewareManager *appmiddleware.Manager
	Config            *config.Config
}

// setupHandlers registers all HTTP handlers.
func setupHandlers(handlers []web.Handler, e *echo.Echo) error {
	for _, handler := range handlers {
		if handler == nil {
			return errors.New("nil handler encountered during registration")
		}
		handler.Register(e)
	}
	return nil
}

// setupApplication initializes middleware and handlers.
func setupApplication(params appParams) error {
	params.MiddlewareManager.Setup(params.Echo)
	return setupHandlers(params.Handlers, params.Echo)
}

// setupLifecycle configures application lifecycle hooks.
func setupLifecycle(params appParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("Starting application...",
				"app_name", params.Config.App.Name,
				"version", params.Config.App.Version,
				"environment", params.Config.App.Env,
			)
			if err := setupApplication(params); err != nil {
				return fmt.Errorf("application setup failed: %w", err)
			}
			params.Logger.Info("Application started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Shutting down application...")
			return nil
		},
	})
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize core dependencies
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		return
	}

	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     cfg.App.Name,
		Version:     cfg.App.Version,
		Environment: cfg.App.Env,
		Fields:      map[string]any{},
	})
	logger, err := factory.CreateLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return
	}

	app := fx.New(
		// Provide core dependencies
		fx.Supply(cfg, logger),
		// Load infrastructure and domain modules
		infrastructure.Module,
		domain.Module,
		fx.Invoke(setupLifecycle),
	)

	// Create a context that will be canceled on interrupt
	stopCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the application
	if err := app.Start(stopCtx); err != nil {
		stop() // Ensure signal handler is stopped
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", err)
		return
	}

	// Wait for interrupt signal
	<-stopCtx.Done()

	// Create a new context for shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer shutdownCancel()

	// Stop the application
	if err := app.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop application gracefully: %v\n", err)
		return
	}
}
