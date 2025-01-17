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
	"go.uber.org/fx/fxevent"
)

func main() {
	log := logger.GetLogger()

	// Try to load .env file, but don't panic if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found, using environment variables")
	}

	app := fx.New(
		fx.Provide(
			logger.GetLogger,
			config.New,
			app.NewEcho,
			app.NewTemplateProvider,
			database.New,
			func(db *sqlx.DB) models.DB { return db },
			func(db *sqlx.DB) handlers.PingContexter { return db },
			models.NewSubscriptionStore,
			models.NewContactStore,
			handlers.NewMarketingHandler,
			handlers.NewSubscriptionHandler,
			handlers.NewHealthHandler,
			handlers.NewContactHandler,
			app.NewApp,
		),
		fx.Invoke(func(a *app.App) {
			log.Info("Application started successfully")
		}),
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}
