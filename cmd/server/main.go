package main

import (
	"github.com/joho/godotenv"
	"github.com/jonesrussell/goforms/internal/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	// Try to load .env file, but don't panic if it doesn't exist
	if err := godotenv.Load(); err != nil {
		logger, _ := zap.NewDevelopment()
		logger.Warn("No .env file found, using environment variables")
	}

	fx.New(
		app.NewModule(),
	).Run()
}
