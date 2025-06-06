// Package main is the entry point for the GoFormX application.
// It sets up the application using dependency injection (via fx),
// configures the server, and manages the application lifecycle.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
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

// provideEcho creates a new Echo server instance
func provideEcho(logger logging.Logger) (*echo.Echo, error) {
	logger.Info("Initializing Echo server")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = middleware.NewValidator()

	// Add basic health check route
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Configure Echo
	e.Debug = true // Enable debug mode for development

	logger.Info("Echo server initialized successfully")
	return e, nil
}

// configureMiddleware sets up the middleware on the Echo instance
func configureMiddleware(e *echo.Echo, mwManager *middleware.Manager, logger logging.Logger) error {
	logger.Info("Configuring middleware")
	mwManager.Setup(e)
	logger.Info("Middleware configuration completed")
	return nil
}

// configureServerLifecycle sets up the server lifecycle hooks
func configureServerLifecycle(lc fx.Lifecycle, e *echo.Echo, cfg *config.Config, logger logging.Logger) {
	logger.Info("Configuring server lifecycle")

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
			logger.Info("Starting server",
				logging.StringField("addr", addr),
				logging.StringField("host", cfg.Server.Host),
				logging.IntField("port", cfg.Server.Port),
				logging.StringField("env", cfg.App.Env),
			)

			// Start server in a goroutine
			go func() {
				logger.Info("Server starting to listen", logging.StringField("addr", addr))
				if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("Server error",
						logging.ErrorField("error", err),
						logging.StringField("addr", addr),
					)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down server")
			shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
			defer cancel()
			return e.Shutdown(shutdownCtx)
		},
	})

	logger.Info("Server lifecycle configured")
}

// main is the entry point of the application.
func main() {
	// Phase 1: Initialize logging
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "goforms",
		Version:     "1.0.0",
		Environment: "development",
		Fields: map[string]any{
			"version": "1.0.0",
		},
	})
	startupLogger, err := factory.CreateLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create startup logger: %v\n", err)
		os.Exit(1)
	}
	startupLogger.Info("Starting application initialization...")

	// Phase 2: Create fx application
	startupLogger.Info("Creating fx application...")

	app := fx.New(
		// Core infrastructure
		fx.Provide(
			func() (logging.Logger, error) {
				return factory.CreateLogger()
			},
			provideShutdownConfig,
			provideEcho,
		),
		// Infrastructure modules
		infrastructure.RootModule,
		// Domain modules
		domain.Module,
		// Presentation modules
		view.Module,
		// Lifecycle hooks
		fx.Invoke(
			initializeLogger,
			configureMiddleware,
			configureServerLifecycle,
		),
	)

	if appErr := app.Err(); appErr != nil {
		startupLogger.Error("Failed to create fx application",
			logging.ErrorField("error", appErr))
		os.Exit(1)
	}

	// Phase 3: Start application
	startupLogger.Info("Starting fx application...")
	startCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	if startErr := app.Start(startCtx); startErr != nil {
		startupLogger.Error("Failed to start application",
			logging.ErrorField("error", startErr))
		os.Exit(1)
	}
	startupLogger.Info("Fx application started successfully")

	// Phase 4: Handle shutdown
	handleShutdown(app, startupLogger)
}

// handleShutdown manages the graceful shutdown of the application
func handleShutdown(app *fx.App, logger logging.Logger) {
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutdown signal received")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	// Stop the application
	if err := app.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown",
			logging.ErrorField("error", err))
		os.Exit(1)
	}

	logger.Info("Application shutdown complete")
}
