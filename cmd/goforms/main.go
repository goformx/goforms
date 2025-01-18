package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/api"
	"github.com/jonesrussell/goforms/internal/app/server"
	"github.com/jonesrussell/goforms/internal/core"
	"github.com/jonesrussell/goforms/internal/platform"
	"github.com/jonesrussell/goforms/internal/web"
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
		// Platform modules
		platform.Module,

		// Core business logic
		core.Module,

		// API handlers
		api.Module,

		// Web handlers
		web.Module,

		fx.Invoke(server.Start),
	)

	app.Run()
}
