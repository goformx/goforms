package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/jonesrussell/goforms/internal/domain"
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
	// godotenv loads .env into os.Environ()
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v\n", err)
		// Continue even if .env is missing - we might be in production
		// with real environment variables
	}

	// envconfig (used in config.New()) reads from os.Environ()
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logging.NewLogger(cfg)

	app := fx.New(
		// Core modules
		infrastructure.Module,
		domain.Module,

		// Configure logging
		fx.WithLogger(func() fxevent.Logger {
			return &logging.FxEventLogger{Logger: logger}
		}),

		// Invoke server startup
		fx.Invoke(func(lifecycle fx.Lifecycle, log logging.Logger) {
			log.Info("Starting GoForms",
				logging.String("version", version),
				logging.String("commit", gitCommit),
				logging.String("buildTime", buildTime),
				logging.String("goVersion", goVersion),
			)
		}),
	)

	if err := app.Start(context.Background()); err != nil {
		log.Printf("Failed to start application: %v\n", err)
		os.Exit(1)
	}

	// Block until context is done
	<-app.Done()

	// Graceful shutdown
	if err := app.Stop(context.Background()); err != nil {
		log.Printf("Failed to stop application: %v\n", err)
		os.Exit(1)
	}
}
