package application

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// App represents the application
type App struct {
	echo   *echo.Echo
	logger logging.Logger
}

// NewApp creates a new application instance
func NewApp(e *echo.Echo, log logging.Logger) *App {
	return &App{
		echo:   e,
		logger: log,
	}
}

// RegisterHooks sets up the application hooks
func RegisterHooks(app *App) {
	app.logger.Info("Application started successfully")
}
