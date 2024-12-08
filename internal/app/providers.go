package app

import (
	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/database"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module exports all app providers
var Module = fx.Options(
	fx.Provide(
		config.Load,
		NewLogger,
		database.New,
		NewEcho,
		AsModelsDB,
		models.NewSubscriptionStore,
		handlers.NewSubscriptionHandler,
		NewApp,
	),
	fx.Invoke(func(app *App) {}),
)

func NewLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func NewEcho() *echo.Echo {
	return echo.New()
}

func AsModelsDB(db *sqlx.DB) models.DB {
	return db
}
