package main

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/jonesrussell/goforms/internal/app"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/database"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
)

func main() {
	// Try to load .env file
	_ = godotenv.Load()

	fx.New(
		fx.Provide(
			logger.GetLogger,
			config.New,
			database.New,
			echo.New,
			func(db *database.DB) handlers.PingContexter {
				return db
			},
			func(db *database.DB) models.DB {
				return db
			},
			func(db *database.DB) *sqlx.DB {
				return db.DB
			},
			models.NewSubscriptionStore,
			models.NewContactStore,
			handlers.NewMarketingHandler,
			handlers.NewSubscriptionHandler,
			handlers.NewHealthHandler,
			handlers.NewContactHandler,
			app.NewApp,
		),
		fx.Invoke(app.RegisterHooks),
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger.UnderlyingZap(logger.GetLogger())}
		}),
	).Run()

	fmt.Println("Server stopped")
	os.Exit(0)
}
