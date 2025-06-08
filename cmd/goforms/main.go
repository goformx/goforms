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

type appParams struct {
	fx.In

	Lifecycle         fx.Lifecycle
	Echo              *echo.Echo
	Server            *server.Server
	Logger            logging.Logger
	Handlers          []web.Handler `group:"handlers"`
	MiddlewareManager *appmiddleware.Manager
}

func main() {
	// Create the application with all dependencies
	app := fx.New(
		// Core infrastructure
		fx.Provide(
			setupLogger,
			echo.New,
			config.New,
		),
		// Domain services
		domain.Module,
		// Infrastructure and handlers
		infrastructure.Module,
		// Application lifecycle
		fx.Invoke(func(params appParams) {
			// Setup startup hook
			params.Lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					params.Logger.Info("Starting application...")

					// Log server configuration
					params.Logger.Info("Server configuration",
						logging.StringField("host", params.Server.Config().App.Host),
						logging.IntField("port", params.Server.Config().App.Port),
						logging.StringField("environment", params.Server.Config().App.Env),
						logging.StringField("server_type", "echo"),
					)

					// Register all handlers
					params.Logger.Info("Registering handlers...")
					for _, h := range params.Handlers {
						h.Register(params.Echo)
					}
					params.Logger.Info("Handlers registered successfully")

					// Setup middleware
					params.Logger.Info("Setting up middleware...")
					params.MiddlewareManager.Setup(params.Echo)
					params.Logger.Info("Middleware setup completed")

					// Start the server after all middleware is registered
					if err := params.Server.Start(); err != nil {
						return fmt.Errorf("failed to start server: %w", err)
					}

					// Log server URL after successful start
					params.Logger.Info("Server started successfully",
						logging.StringField("host", params.Server.Config().App.Host),
						logging.IntField("port", params.Server.Config().App.Port),
						logging.StringField("address", params.Server.Address()),
						logging.StringField("url", params.Server.URL()),
						logging.StringField("environment", params.Server.Config().App.Env),
					)

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
	if err := app.Start(ctx); err != nil {
		log.Fatal("Failed to start application:", err)
	}

	// Wait for interrupt signal
	<-ctx.Done()

	// Create a new context for shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop the application
	if err := app.Stop(shutdownCtx); err != nil {
		log.Fatal("Failed to stop application:", err)
	}
}

func setupLogger() (logging.Logger, error) {
	return logging.NewFactory(logging.FactoryConfig{
		AppName:     "goforms",
		Version:     "1.0.0",
		Environment: "development",
		Fields: map[string]any{
			"version": "1.0.0",
		},
	}).CreateLogger()
}
