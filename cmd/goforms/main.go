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

// setupLogger creates a new logger instance
func setupLogger(cfg *config.Config) (logging.Logger, error) {
	if cfg == nil {
		return nil, errors.New("config is required for logger setup")
	}
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
	if err := registerHandlers(params.Logger, params.Echo, params.Handlers); err != nil {
		return fmt.Errorf("failed to register handlers: %w", err)
	}

	// Setup middleware
	params.MiddlewareManager.Setup(params.Echo)

	// Start the server
	if err := params.Server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// registerHandlers registers all application handlers
func registerHandlers(logger logging.Logger, e *echo.Echo, handlers []web.Handler) error {
	logger.Info("registering handlers")

	for _, h := range handlers {
		if h == nil {
			return errors.New("nil handler encountered during registration")
		}
		h.Register(e)
	}

	logger.Info("handlers registered successfully")
	return nil
}

// createFallbackLogger creates a basic logger for emergency error reporting
func createFallbackLogger() logging.Logger {
	logger, _ := logging.NewFactory(logging.FactoryConfig{
		AppName:     "goforms",
		Version:     "1.0.0",
		Environment: "production",
		Fields:      map[string]any{},
	}).CreateLogger()
	return logger
}

func main() {
	// Load configuration first
	cfg, err := config.New()
	if err != nil {
		fallbackLogger := createFallbackLogger()
		fallbackLogger.Fatal("Failed to load configuration",
			logging.ErrorField("error", err),
		)
	}

	// Create logger first
	logger, err := setupLogger(cfg)
	if err != nil {
		fallbackLogger := createFallbackLogger()
		fallbackLogger.Fatal("Failed to setup logger",
			logging.ErrorField("error", err),
		)
	}

	// Create the application with all dependencies
	app := fx.New(
		// Core infrastructure
		fx.Provide(
			func() logging.Logger { return logger }, // Provide the already created logger
			func() *config.Config { return cfg },    // Provide the already loaded config
			echo.New,
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
		logger.Fatal("Failed to start application",
			logging.ErrorField("error", startErr),
		)
	}

	// Wait for interrupt signal
	<-ctx.Done()

	// Create a new context for shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	// Stop the application
	if stopErr := app.Stop(shutdownCtx); stopErr != nil {
		stop() // Ensure signal handler is stopped
		if shutdownCtx.Err() == context.DeadlineExceeded {
			logger.Fatal("Application shutdown timed out",
				logging.Duration("timeout", DefaultShutdownTimeout),
			)
		} else {
			logger.Fatal("Failed to stop application",
				logging.ErrorField("error", stopErr),
			)
		}
	}

	logger.Info("Application shutdown completed successfully")
}
