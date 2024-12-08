package main

import (
	"github.com/joho/godotenv"
	"github.com/jonesrussell/goforms/internal/app"
	"go.uber.org/fx"
)

func main() {
	// Load .env file before fx initialization
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file: " + err.Error())
	}

	fx.New(
		app.Module(),
	).Run()
}
