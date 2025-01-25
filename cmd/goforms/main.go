package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/jonesrussell/goforms/internal/application"
	"github.com/jonesrussell/goforms/internal/application/http"
	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/infrastructure"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	view "github.com/jonesrussell/goforms/internal/presentation/templates"
)

//nolint:gochecknoglobals // These variables are populated by -ldflags at build time
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
	goVersion = "unknown"
)

func main() {
	// Print version information
	fmt.Printf("GoForms %s (%s) built with %s\n", version, gitCommit[:7], goVersion)
	fmt.Printf("Build time: %s\n", buildTime)

	// Try to load .env file
	_ = godotenv.Load()

	app := fx.New(
		// Infrastructure modules (must be first to provide config)
		infrastructure.Module,

		// Logging module
		logging.Module,

		// Domain modules
		domain.Module,

		// Application modules
		application.Module,
		http.Module,

		// Presentation modules
		view.Module,

		// Configure logging
		fx.WithLogger(func(logger logging.Logger) fxevent.Logger {
			return &logging.FxEventLogger{Logger: logger}
		}),
	)

	// Create a context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the application
	if err := app.Start(ctx); err != nil {
		os.Exit(1)
	}

	// Wait for interrupt signal
	<-ctx.Done()

	// Stop the application gracefully
	if err := app.Stop(ctx); err != nil {
		os.Exit(1)
	}
}
