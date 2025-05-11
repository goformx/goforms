// Package main is the entry point for the GoForms application.
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

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/application/middleware"
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

const (
	// ShutdownTimeout is the maximum time to wait for graceful shutdown
	ShutdownTimeout = 5 * time.Second
)

// main is the entry point of the application.
// It calls run() and handles any fatal errors that occur during startup.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run orchestrates the application startup process.
// It sets up signal handling and starts the application.
func run() error {
	// Create a temporary logger for startup
	var logger *zap.Logger
	var err error

	// Load .env file
	if loadErr := godotenv.Load(); loadErr != nil {
		// Initialize a basic logger for startup errors
		logger, err = zap.NewDevelopment()
		if err != nil {
			return fmt.Errorf("failed to create startup logger: %w", err)
		}
		logger.Warn("failed to load .env file", zap.Error(loadErr))
	} else {
		// Initialize logger based on environment
		if os.Getenv("GOFORMS_APP_ENV") == "development" {
			logger, err = zap.NewDevelopment()
		} else {
			logger, err = zap.NewProduction()
		}
		if err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}
	}
	defer logger.Sync()

	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a goroutine to handle termination signals
	go handleSignals(cancel)

	// Create and run the application
	app := createApp(logger)
	return runApp(ctx, app)
}

// createApp sets up the dependency injection container using fx.
// It provides all necessary dependencies and modules for the application.
func createApp(logger *zap.Logger) *fx.App {
	return fx.New(
		// Core dependencies that are required for basic functionality
		fx.Provide(
			GetVersion,
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
		fx.Provide(
			newServer,
		),
		// Custom logger for fx events
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		// Start the server using fx.Invoke
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
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
func newServer(
	cfg *config.Config,
	logFactory *logging.Factory,
	userService user.Service,
) (
	*echo.Echo,
	*middleware.Manager,
	error,
) {
	// Create logger instance
	logger, err := logFactory.CreateFromConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Set up request validation
	e.Validator = middleware.NewValidator()

	// Configure middleware stack using Manager pattern
	mwManager := middleware.New(&middleware.ManagerConfig{
		Logger:      logger,
		UserService: userService,
		Security:    &cfg.Security,
	})
	mwManager.Setup(e)

	return e, mwManager, nil
}

// ServerParams contains the dependencies required for starting the server.
// It uses fx.In to automatically inject dependencies.
type ServerParams struct {
	fx.In

	Server            *server.Server
	Config            *config.Config
	Logger            logging.Logger
	Handlers          []handlers.Handler `group:"handlers"`
	MiddlewareManager *middleware.Manager
}

// startServer registers all handlers with the server.
// It uses fx.In to automatically inject dependencies.
func startServer(params ServerParams) error {
	// Register all handlers with the middleware manager
	for _, h := range params.Handlers {
		if webHandler, ok := h.(*handler.WebHandler); ok {
			handler.WithMiddlewareManager(params.MiddlewareManager)(webHandler)
		}
		h.Register(params.Server.Echo())
	}

	return nil
}

// GetVersion returns the version information
func GetVersion() version.Info {
	return version.GetInfo()
}
