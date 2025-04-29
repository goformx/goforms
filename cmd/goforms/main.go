package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/application/router"
	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

//nolint:gochecknoglobals // These variables are populated by -ldflags at build time
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
	goVersion = "unknown"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// Create version info
	versionInfo := handler.VersionInfo{
		Version:   version,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		GoVersion: goVersion,
	}

	// Create app with DI
	app := fx.New(
		// Version info first
		fx.Provide(
			func() handler.VersionInfo {
				return versionInfo
			},
		),
		// Core infrastructure next (config, logging, database)
		infrastructure.Module,
		// Domain services next
		domain.Module,
		// View module for rendering
		view.Module,
		// Local providers last
		fx.Provide(
			logging.NewFactory,
			newServer,
		),
		fx.WithLogger(func(log logging.Logger) fxevent.Logger {
			return &logging.FxEventLogger{Logger: log}
		}),
		// Add debug logging for dependency injection
		fx.Invoke(func(log logging.Logger) {
			log.Debug("checking module initialization")
		}),
		fx.Invoke(func(p infrastructure.HandlerParams) {
			p.Logger.Debug("handler dependencies available",
				logging.Bool("renderer_available", p.Renderer != nil),
				logging.Bool("contact_service_available", p.ContactService != nil),
				logging.Bool("subscription_service_available", p.SubscriptionService != nil),
				logging.Bool("user_service_available", p.UserService != nil),
			)
		}),
		// Start server last
		fx.Invoke(startServer),
	)

	// Run app
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	<-app.Done()
	if err := app.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}
	return nil
}

func newServer(cfg *config.Config, logFactory *logging.Factory, userService user.Service) (*echo.Echo, error) {
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

	// Add CSRF token middleware if CSRF is enabled
	if cfg.Security.CSRF.Enabled {
		e.Use(middleware.CSRFToken())
	}

	return e, nil
}

// ServerParams contains the dependencies for starting the server
type ServerParams struct {
	fx.In

	Echo     *echo.Echo
	Config   *config.Config
	Logger   logging.Logger
	Handlers []handler.Handler `group:"handlers"`
}

func startServer(p ServerParams) error {
	p.Logger.Debug("starting server with handlers",
		logging.Int("handler_count", len(p.Handlers)),
	)

	for i, h := range p.Handlers {
		p.Logger.Debug("handler available",
			logging.Int("index", i),
			logging.String("type", fmt.Sprintf("%T", h)),
		)
	}

	// Configure routes
	if err := router.Setup(p.Echo, &router.Config{
		Handlers: p.Handlers,
		Static: router.StaticConfig{
			Path: "/static",
			Root: "static",
		},
		Logger: p.Logger,
	}); err != nil {
		return fmt.Errorf("failed to setup router: %w", err)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", p.Config.Server.Host, p.Config.Server.Port)
	if p.Config.Server.Port == 0 {
		addr = fmt.Sprintf("%s:8090", p.Config.Server.Host) // Default to 8090 if port is not set
	}

	p.Logger.Info("Starting server",
		logging.String("addr", addr),
		logging.String("env", p.Config.App.Env),
		logging.String("version", version),
		logging.String("gitCommit", gitCommit),
	)

	return p.Echo.Start(addr)
}
