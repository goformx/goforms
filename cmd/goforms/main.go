// Package main is the entry point for the GoFormX application.
// It sets up the application using dependency injection (via fx),
// configures the server, and manages the application lifecycle.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"
	"go.uber.org/zap"

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
			params.Lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					params.Logger.Info("Starting application...")

					// Register all handlers
					params.Logger.Info("Registering handlers...")
					for _, h := range params.Handlers {
						h.Register(params.Echo)
					}

					// Setup middleware
					params.Logger.Info("Setting up middleware...")
					params.MiddlewareManager.Setup(params.Echo)

					// Start server in a goroutine
					go func() {
						if err := params.Server.Start(ctx); err != nil {
							params.Logger.Fatal("Failed to start server", zap.Error(err))
						}
					}()

					// Wait for interrupt signal
					quit := make(chan os.Signal, 1)
					signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
					<-quit

					return nil
				},
				OnStop: func(ctx context.Context) error {
					params.Logger.Info("Shutting down application...")
					return params.Server.Stop(ctx)
				},
			})
		}),
	)

	// Start the application
	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Stop the application
	if err := app.Stop(context.Background()); err != nil {
		log.Fatal(err)
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
