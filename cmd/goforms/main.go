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
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/infrastructure"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
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
		// Core dependencies
		fx.Provide(
			func() handler.VersionInfo {
				return versionInfo
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

	return e, nil
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
	p.Logger.Debug("starting server with handlers",
		logging.Int("handler_count", len(p.Handlers)),
	)

	for i, h := range p.Handlers {
		p.Logger.Debug("handler available",
			logging.Int("index", i),
			logging.String("type", fmt.Sprintf("%T", h)),
		)
	}

	// Register static files first
	p.Server.Echo().Static("/static", "static")
	p.Server.Echo().Static("/static/dist", "static/dist")
	p.Server.Echo().File("/favicon.ico", "static/favicon.ico")
	p.Server.Echo().File("/robots.txt", "static/robots.txt")

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
