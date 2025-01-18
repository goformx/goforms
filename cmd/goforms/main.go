package main

import (
	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/api"
	"github.com/jonesrussell/goforms/internal/app/server"
	"github.com/jonesrussell/goforms/internal/core"
	"github.com/jonesrussell/goforms/internal/platform"
	"github.com/jonesrussell/goforms/internal/web"
)

func main() {
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
