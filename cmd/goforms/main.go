// Package main is the entry point for the GoFormX application.
// It sets up the application using dependency injection (via fx),
// configures the server, and manages the application lifecycle.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/bootstrap"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/presentation/view"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 5 * time.Second
)

// ShutdownConfig holds configuration for application shutdown
type ShutdownConfig struct {
	Timeout time.Duration `envconfig:"GOFORMS_SHUTDOWN_TIMEOUT" default:"5s"`
}

// provideShutdownConfig creates a new shutdown configuration
func provideShutdownConfig(cfg *config.Config) *ShutdownConfig {
	return &ShutdownConfig{
		Timeout: cfg.Server.ShutdownTimeout,
	}
}

// initializeLogger initializes the application logger
func initializeLogger(logger logging.Logger) logging.Logger {
	logger.Info("Application started")
	return logger
}

// main is the entry point of the application.
func main() {
	// Create logger
	logger, err := logging.NewFactory().CreateLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		return
	}

	// Collect all fx options in a single slice
	options := []fx.Option{
		// Core modules
		infrastructure.RootModule,
		domain.Module,
		view.Module,

		// Bootstrap providers
		fx.Provide(provideShutdownConfig),

		// Invoke functions
		fx.Invoke(
			startServer,
			initializeLogger,
		),

		// Lifecycle hooks
		fx.Invoke(func(lc fx.Lifecycle, logger logging.Logger, cfg *ShutdownConfig) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					logger.Info("Shutting down application")
					return nil
				},
			})
		}),
	}

	// Add bootstrap providers
	options = append(options, bootstrap.Providers()...)
	options = append(options, bootstrap.ServerProviders()...)
	options = append(options, bootstrap.HandlerProviders()...)

	// Create the application with fx
	app := fx.New(options...)

	// Start the application
	if startErr := app.Start(context.Background()); startErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", startErr)
		return
	}

	// Handle shutdown
	handleShutdown(app, logger)
}

// handleShutdown manages the graceful shutdown of the application
func handleShutdown(app *fx.App, logger logging.Logger) {
	// Set up signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	sig := <-signalChan
	logger.Info("Received shutdown signal", logging.String("signal", sig.String()))

	// Create shutdown context with default timeout
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancelShutdown()

	// Start graceful shutdown
	if err := app.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop application: %v\n", err)
		return
	}
}

// ServerParams contains the dependencies required for starting the server.
type ServerParams struct {
	fx.In

	Server            *server.Server
	Config            *config.Config
	Logger            logging.Logger
	WebHandlers       []web.Handler `group:"web_handlers"`
	MiddlewareManager *middleware.Manager
}

// startServer registers all handlers with the server.
func startServer(params ServerParams) error {
	// Register all web handlers with the server
	for _, h := range params.WebHandlers {
		h.Register(params.Server.Echo())
	}

	return nil
}
