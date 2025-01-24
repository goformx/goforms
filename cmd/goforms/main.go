package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/jonesrussell/goforms/internal/api"
	"github.com/jonesrussell/goforms/internal/application"
	"github.com/jonesrussell/goforms/internal/auth"
	"github.com/jonesrussell/goforms/internal/core"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/platform"
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

	// Initialize logger
	log := logger.GetLogger()

	app := fx.New(
		// Platform modules
		platform.Module,

		// Core business logic
		core.Module,

		// Authentication module
		auth.Module,

		// API handlers
		api.Module,

		// App configuration
		application.Module,

		// Configure fx to use our logger
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger.UnderlyingZap(log)}
		}),
	)

	app.Run()
}
