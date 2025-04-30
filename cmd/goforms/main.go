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

	"github.com/joho/godotenv"
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

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := loadEnvironment(); err != nil {
		return err
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle termination signals
	go handleSignals(cancel)

	app := createApp()
	return runApp(ctx, app)
}

func handleSignals(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	cancel()
}

func loadEnvironment() error {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v\n", err)
	}
	return nil
}

func createApp() *fx.App {
	return fx.New(
		// Core dependencies
		fx.Provide(
			func() version.VersionInfo {
				return version.Info()
			},
			logging.NewFactory,
			func(cfg *config.Config, logFactory *logging.Factory) (logging.Logger, error) {
				return logFactory.CreateFromConfig(cfg)
			},
		),
		// Infrastructure module
		infrastructure.Module,
		// Domain module
		domain.Module,
		// View module
		view.Module,
		// Server setup
		fx.Provide(newServer),
		// Logger setup
		fx.WithLogger(func(logger logging.Logger) fxevent.Logger {
			return &logging.FxEventLogger{Logger: logger}
		}),
		// Start server
		fx.Invoke(startServer),
	)
}

func runApp(ctx context.Context, app *fx.App) error {
	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Create a new context for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}
	return nil
}

func newServer(cfg *config.Config, logFactory *logging.Factory, userService user.Service) (*echo.Echo, error) {
	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create logger
	logger, err := logFactory.CreateFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Register validator
	e.Validator = middleware.NewValidator()

	// Configure middleware
	middleware.Setup(e, &middleware.Config{
		Logger:      logger,
		JWTSecret:   cfg.Security.JWTSecret,
		UserService: userService,
		EnableCSRF:  cfg.Security.CSRF.Enabled,
	})

	return e, nil
}

func validateConfig(cfg *config.Config) error {
	if cfg.Security.JWTSecret == "" {
		return errors.New("JWT secret is required")
	}
	if cfg.Security.CSRF.Enabled && cfg.Security.CSRF.Secret == "" {
		return errors.New("CSRF secret is required when CSRF is enabled")
	}
	return nil
}

// ServerParams contains the dependencies for starting the server
type ServerParams struct {
	fx.In

	Server   *server.Server
	Config   *config.Config
	Logger   logging.Logger
	Handlers []handlers.Handler `group:"handlers"`
}

func startServer(p ServerParams) error {
	// Get handler types for logging
	handlerTypes := make([]string, len(p.Handlers))
	for i, h := range p.Handlers {
		handlerTypes[i] = fmt.Sprintf("%T", h)
	}

	p.Logger.Debug("starting server with handlers",
		logging.Int("handler_count", len(p.Handlers)),
		logging.String("handler_types", fmt.Sprintf("%v", handlerTypes)),
	)

	// Register static files
	registerStaticFiles(p.Server.Echo())

	// Configure routes
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

func registerStaticFiles(e *echo.Echo) {
	e.Static("/static", "static")
	e.Static("/static/dist", "static/dist")
	e.File("/favicon.ico", "static/favicon.ico")
	e.File("/robots.txt", "static/robots.txt")
}
