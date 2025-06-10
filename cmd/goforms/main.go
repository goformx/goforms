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

	"github.com/goformx/goforms/internal/application"
	"github.com/goformx/goforms/internal/application/handlers/web"
	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/presentation"
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
			params.Logger.Info("starting application",
				"app", params.Config.App.Name,
				"version", params.Config.App.Version,
				"environment", params.Config.App.Env,
			)

			// Start the server in a goroutine
			go func() {
				if err := params.Server.Start(); err != nil {
					params.Logger.Error("server error",
						"error", err,
						"app", params.Config.App.Name,
					)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("shutting down application",
				"app", params.Config.App.Name,
				"version", params.Config.App.Version,
			)

			// The server shutdown is handled by the server's lifecycle hooks
			return nil
		},
	})
}

func main() {
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger factory
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     cfg.App.Name,
		Version:     cfg.App.Version,
		Environment: cfg.App.Env,
		Fields:      map[string]any{},
	})

	// Create logger instance
	logger, err := factory.CreateLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Create the application
	app := fx.New(
		// Supply core dependencies
		fx.Supply(cfg),

		// Infrastructure module
		infrastructure.Module,

		// Domain module
		domain.Module,

		// Application module
		application.Module,

		// Presentation module
		presentation.Module,

		// Setup application
		fx.Invoke(setupApplication),
		fx.Invoke(setupLifecycle),
	)

	// Start the application
	if err := app.Start(context.Background()); err != nil {
		logger.Error("failed to start application", "error", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop the application
	if err := app.Stop(context.Background()); err != nil {
		logger.Error("failed to stop application", "error", err)
		os.Exit(1)
	}
}
