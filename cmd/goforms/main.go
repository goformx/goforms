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
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	webhandler "github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
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
func configureServerLifecycle(lc fx.Lifecycle, srv *server.Server, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting server via lifecycle hook")
			return srv.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping server via lifecycle hook")
			return srv.Stop(ctx)
		},
	})
}

// main is the entry point of the application.
func main() {
	// Collect all fx options in a single slice
	options := []fx.Option{
		// Core modules
		infrastructure.RootModule,
		domain.Module,
		view.Module,

		// Logger setup
		fx.Provide(
			func() (logging.Logger, error) {
				return logging.NewFactory().CreateLogger()
			},
		),

		// Server setup
		fx.Provide(
			provideEcho,
			provideShutdownConfig,
		),
		fx.Invoke(
			initializeLogger,
			configureMiddleware,
			configureServerLifecycle,
		),

		// Web handlers
		fx.Provide(
			func(formService form.Service, logger logging.Logger) *webhandler.BaseHandler {
				return webhandler.NewBaseHandler(formService, logger)
			},
			func(deps webhandler.HandlerDeps) (*webhandler.AuthHandler, error) {
				return webhandler.NewAuthHandler(deps)
			},
			func(deps webhandler.HandlerDeps, formService form.Service) (*webhandler.PageHandler, error) {
				return webhandler.NewPageHandler(deps, formService)
			},
			func(deps webhandler.HandlerDeps) (*webhandler.WebHandler, error) {
				return webhandler.NewWebHandler(deps)
			},
			func(deps webhandler.HandlerDeps) *webhandler.DemoHandler {
				return webhandler.NewDemoHandler(deps)
			},
			func(deps webhandler.HandlerDeps, formService form.Service) *webhandler.FormHandler {
				return webhandler.NewFormHandler(deps, formService)
			},
		),

		// Group handlers
		fx.Provide(
			fx.Annotate(
				func(deps webhandler.HandlerDeps) webhandler.Handler {
					deps.Logger.Info("Registering demo handler in web_handlers group")
					return webhandler.NewDemoHandler(deps)
				},
				fx.ResultTags(`group:"web_handlers"`),
			),
			fx.Annotate(
				func(deps webhandler.HandlerDeps, formService form.Service) webhandler.Handler {
					deps.Logger.Info("Registering form handler in web_handlers group")
					return webhandler.NewFormHandler(deps, formService)
				},
				fx.ResultTags(`group:"web_handlers"`),
			),
		),

		// Add fx logger
		fx.WithLogger(func(logger logging.Logger) fxevent.Logger {
			if zapLogger, ok := logger.(*logging.ZapLogger); ok {
				return &fxevent.ZapLogger{Logger: zapLogger.GetZapLogger()}
			}
			devLogger, _ := zap.NewDevelopment()
			return &fxevent.ZapLogger{Logger: devLogger}
		}),
	}

	// Create the application with fx
	app := fx.New(options...)

	// Start the application
	if startErr := app.Start(context.Background()); startErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to start application: %v\n", startErr)
		return
	}

	// Handle shutdown
	handleShutdown(app)
}

// handleShutdown manages the graceful shutdown of the application
func handleShutdown(app *fx.App) {
	// Set up signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	fmt.Println("Waiting for shutdown signal...")
	sig := <-signalChan
	fmt.Printf("Received shutdown signal: %s\n", sig.String())

	// Create shutdown context with default timeout
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancelShutdown()

	// Start graceful shutdown
	fmt.Println("Starting graceful shutdown...")
	if err := app.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop application: %v\n", err)
		return
	}
	fmt.Println("Application shutdown completed successfully")
}
