package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/jonesrussell/goforms/internal/app"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/database"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
	"go.uber.org/fx"
)

func main() {
	// Try to load .env file, but don't panic if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log := logger.GetLogger()
		log.Warn("No .env file found, using environment variables")
	}

	fx.New(
		fx.Provide(
			logger.GetLogger,
			config.New,
			app.NewEcho,
			app.NewTemplateProvider,
			database.New,
			func(db *sqlx.DB) models.DB { return db },
			models.NewSubscriptionStore,
			models.NewContactStore,
			handlers.NewMarketingHandler,
			handlers.NewSubscriptionHandler,
			handlers.NewHealthHandler,
			handlers.NewContactHandler,
			app.NewApp,
		),
		fx.Invoke(func(app *app.App) {}),
	).Run()
}
