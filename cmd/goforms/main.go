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

type appParams struct {
	fx.In

	Lifecycle         fx.Lifecycle
	Echo              *echo.Echo
	Server            *server.Server
	Logger            logging.Logger
	Handlers          []web.Handler `group:"handlers"`
	MiddlewareManager *appmiddleware.Manager
}

// setupServer configures and starts the server
func setupServer(params appParams) error {
	// Log server configuration
	logServerConfig(params.Logger, params.Server)

	// Register all handlers
	registerHandlers(params.Logger, params.Echo, params.Handlers)

	// Setup middleware
	setupMiddleware(params.Logger, params.Echo, params.MiddlewareManager)

	// Start the server
	if err := params.Server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Log server URL after successful start
	logServerStart(params.Logger, params.Server)

	return nil
}

// logServerConfig logs the server configuration details
func logServerConfig(logger logging.Logger, srv *server.Server) {
	logger.Info("Server configuration",
		logging.StringField("host", srv.Config().App.Host),
		logging.IntField("port", srv.Config().App.Port),
		logging.StringField("environment", srv.Config().App.Env),
		logging.StringField("server_type", "echo"),
	)
}

// registerHandlers registers all application handlers
func registerHandlers(logger logging.Logger, e *echo.Echo, handlers []web.Handler) {
	logger.Info("Registering handlers...")
	for _, h := range handlers {
		h.Register(e)
	}
	logger.Info("Handlers registered successfully")
}

// setupMiddleware configures all middleware
func setupMiddleware(logger logging.Logger, e *echo.Echo, manager *appmiddleware.Manager) {
	logger.Info("Setting up middleware...")
	manager.Setup(e)
	logger.Info("Middleware setup completed")
}

// logServerStart logs the server start information
func logServerStart(logger logging.Logger, srv *server.Server) {
	logger.Info("Server started successfully",
		logging.StringField("host", srv.Config().App.Host),
		logging.IntField("port", srv.Config().App.Port),
		logging.StringField("address", srv.Address()),
		logging.StringField("url", srv.URL()),
		logging.StringField("environment", srv.Config().App.Env),
	)
}

// createLifecycleHooks creates the application lifecycle hooks
func createLifecycleHooks(params appParams) fx.Hook {
	return fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.Logger.Info("Starting application...")
			return setupServer(params)
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Shutting down application...")
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
			echo.New,
			config.New,
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

func setupLogger() (logging.Logger, error) {
	return logging.NewFactory(logging.FactoryConfig{
		AppName:     "goforms",
		Version:     "1.0.0",
		Environment: "development",
		Fields:      map[string]any{},
	}).CreateLogger()
}
