// Package main is the entry point for the GoForms application.
// It sets up the application using dependency injection (via fx),
// configures the server, and manages the application lifecycle.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/application/router"
	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/infrastructure"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/version"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

// main is the entry point of the application.
// It calls run() and handles any fatal errors that occur during startup.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run orchestrates the application startup process.
// It sets up signal handling and starts the application.
func run() error {
	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a goroutine to handle termination signals
	go handleSignals(cancel)

	// Create and run the application
	app := createApp()
	return runApp(ctx, app)
}

// createApp sets up the dependency injection container using fx.
// It provides all necessary dependencies and modules for the application.
func createApp() *fx.App {
	return fx.New(
		// Core dependencies that are required for basic functionality
		fx.Provide(
			func() version.VersionInfo {
				return version.Info()
			},
			logging.NewFactory,
			func(cfg *config.Config, logFactory *logging.Factory) (logging.Logger, error) {
				return logFactory.CreateFromConfig(cfg)
			},
		),
		// Infrastructure module for database, cache, etc.
		infrastructure.Module,
		// Domain module containing business logic
		domain.Module,
		// View module for template rendering
		view.Module,
		// Server setup with Echo framework
		fx.Provide(newServer),
		// Custom logger for fx events
		fx.WithLogger(func(logger logging.Logger) fxevent.Logger {
			return &logging.FxEventLogger{Logger: logger}
		}),
		// Start the server after all dependencies are ready
		fx.Invoke(startServer),
	)
}

// runApp manages the application lifecycle.
// It starts the application, waits for termination signals, and performs graceful shutdown.
func runApp(ctx context.Context, app *fx.App) error {
	// Start the application with the provided context
	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Wait for context cancellation (triggered by termination signals)
	<-ctx.Done()

	// Create a new context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop the application with the shutdown context
	if err := app.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}
	return nil
}

// handleSignals sets up signal handling for graceful shutdown.
// It listens for SIGINT (Ctrl+C) and SIGTERM signals.
func handleSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	cancel()
}

// newServer creates and configures a new Echo server instance.
// It sets up middleware, logging, and security features.
func newServer(cfg *config.Config, logFactory *logging.Factory, userService user.Service) (*echo.Echo, error) {
	// Validate critical configuration settings
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create logger instance
	logger, err := logFactory.CreateFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Set up request validation
	e.Validator = middleware.NewValidator()

	// Configure middleware stack
	middleware.Setup(e, &middleware.Config{
		Logger:      logger,
		JWTSecret:   cfg.Security.JWTSecret,
		UserService: userService,
		EnableCSRF:  cfg.Security.CSRF.Enabled,
	})

	return e, nil
}

// validateConfig checks critical configuration settings.
// It ensures required security settings are present and valid.
func validateConfig(cfg *config.Config) error {
	if cfg.Security.JWTSecret == "" {
		return errors.New("JWT secret is required")
	}
	if cfg.Security.CSRF.Enabled && cfg.Security.CSRF.Secret == "" {
		return errors.New("CSRF secret is required when CSRF is enabled")
	}
	return nil
}

// ServerParams contains the dependencies required for starting the server.
// It uses fx.In to automatically inject dependencies.
type ServerParams struct {
	fx.In

	Server   *server.Server
	Config   *config.Config
	Logger   logging.Logger
	Handlers []handlers.Handler `group:"handlers"`
}

// startServer configures and starts the HTTP server.
// It sets up static file serving and routes.
func startServer(p ServerParams) error {
	// Log handler types for debugging
	handlerTypes := make([]string, len(p.Handlers))
	for i, h := range p.Handlers {
		handlerTypes[i] = fmt.Sprintf("%T", h)
	}

	p.Logger.Debug("starting server with handlers",
		logging.Int("handler_count", len(p.Handlers)),
		logging.String("handler_types", fmt.Sprintf("%v", handlerTypes)),
	)

	// Register static file routes
	registerStaticFiles(p.Server.Echo())

	// Configure application routes
	if err := router.Setup(p.Server.Echo(), &router.Config{
		Handlers: p.Handlers,
		Static: router.StaticConfig{
			Path: "/static",
			Root: "static",
		},
		Logger: p.Logger,
	}); err != nil {
		return fmt.Errorf("failed to setup router: %w", err)
	}

	return nil
}

// registerStaticFiles sets up static file serving for the application.
// It configures routes for static assets, favicon, and robots.txt.
func registerStaticFiles(e *echo.Echo) {
	e.Static("/static", "static")
	e.Static("/static/dist", "static/dist")
	e.File("/favicon.ico", "static/favicon.ico")
	e.File("/robots.txt", "static/robots.txt")
}
