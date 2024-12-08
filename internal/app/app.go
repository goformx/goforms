package app

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/handlers"
)

type App struct {
	logger *zap.Logger
	db     *sqlx.DB
	echo   *echo.Echo
}

func New(cfg *config.Config, logger *zap.Logger, db *sqlx.DB) *App {
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(cfg.RateLimit.Rate))))

	// Custom error handler
	e.HTTPErrorHandler = customErrorHandler(logger)

	return &App{
		logger: logger,
		db:     db,
		echo:   e,
	}
}

func (a *App) Start(address string) error {
	return a.echo.Start(address)
}

func customErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := 500
		message := "Internal Server Error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = he.Message.(string)
		}

		logger.Error("request error",
			zap.Int("status", code),
			zap.String("message", message),
			zap.Error(err),
		)

		if !c.Response().Committed {
			if err := c.JSON(code, map[string]string{
				"error": message,
			}); err != nil {
				logger.Error("error sending error response", zap.Error(err))
			}
		}
	}
}

func (a *App) RegisterHandlers() {
	subscriptionHandler := handlers.NewSubscriptionHandler(a.db, a.logger)
	subscriptionHandler.Register(a.echo)
}
