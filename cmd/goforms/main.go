package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/handler"
	"github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/application/router"
	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
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
		infrastructure.Module,
		domain.Module,
		fx.Provide(
			logging.NewFactory,
			newServer,
			// Provide version info
			fx.Annotated{
				Name: "version_info",
				Target: func() handler.VersionInfo {
					return versionInfo
				},
			},
		),
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
	logger := logFactory.CreateFromConfig(cfg)

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Configure middleware
	middleware.Setup(e, &middleware.Config{
		Logger:      logger,
		JWTSecret:   cfg.Security.JWTSecret,
		UserService: userService,
		EnableCSRF:  cfg.Security.CSRF.Enabled,
	})

	return e, nil
}

func startServer(e *echo.Echo, handlers []handler.Handler, cfg *config.Config, logger logging.Logger) error {
	// Configure routes
	router.Setup(e, &router.Config{
		Handlers: handlers,
		Static: router.StaticConfig{
			Path: "/static",
			Root: "static",
		},
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Starting server",
		logging.String("addr", addr),
		logging.String("env", cfg.App.Env),
		logging.String("version", version),
		logging.String("gitCommit", gitCommit),
	)

	return e.Start(addr)
}
