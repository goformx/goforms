package main

import (
	"github.com/joho/godotenv"
	"github.com/jonesrussell/goforms/internal/app"
	"github.com/jonesrussell/goforms/internal/logger"
	"go.uber.org/fx"
)

func main() {
	// Try to load .env file, but don't panic if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log := logger.GetLogger()
		log.Warn("No .env file found, using environment variables")
	}

	fx.New(
		app.NewModule(),
		fx.Invoke(func(app *app.App) {}),
	).Run()
}
