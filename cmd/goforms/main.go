package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application"
	"github.com/jonesrussell/goforms/internal/application/http"
	"github.com/jonesrussell/goforms/internal/domain"
	"github.com/jonesrussell/goforms/internal/infrastructure"
	"github.com/jonesrussell/goforms/internal/infrastructure/auth"
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
	// Print version information
	fmt.Printf("GoForms %s (%s) built with %s\n", version, gitCommit[:7], goVersion)
	fmt.Printf("Build time: %s\n", buildTime)

	// Try to load .env file
	_ = godotenv.Load()

	app := fx.New(
		// Logging module (must be first to capture other modules' initialization)
		logging.Module,

		// Infrastructure modules
		infrastructure.Module,

		// Domain modules
		domain.Module,

		// Authentication module
		auth.Module,

		// HTTP handlers
		http.Module,

		// Application module
		application.Module,

		// Configure fx to use our logger
		fx.WithLogger(func(log logging.Logger) fx.Printer {
			return &fxLogger{log}
		}),
	)

	app.Run()
}

// fxLogger adapts our logger to fx.Printer
type fxLogger struct {
	log logging.Logger
}

func (l *fxLogger) Printf(format string, args ...interface{}) {
	l.log.Info(fmt.Sprintf(format, args...))
}
