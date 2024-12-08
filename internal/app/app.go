package app

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/handlers"
)

type App struct {
	logger   *zap.Logger
	echo     *echo.Echo
	config   *config.Config
	handlers *handlers.SubscriptionHandler
}

func NewApp(
	lc fx.Lifecycle,
	logger *zap.Logger,
	echo *echo.Echo,
	cfg *config.Config,
	handler *handlers.SubscriptionHandler,
) *App {
	app := &App{
		logger:   logger,
		echo:     echo,
		config:   cfg,
		handlers: handler,
	}

	app.setupMiddleware()
	app.registerHandlers()

	lc.Append(fx.Hook{
		OnStart: app.start,
		OnStop:  app.stop,
	})

	return app
}

func (a *App) setupMiddleware() {
	a.echo.Use(middleware.Recover())
	a.echo.Use(middleware.RequestID())
	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	}))
	a.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(a.config.RateLimit.Rate))))

	a.echo.HTTPErrorHandler = a.customErrorHandler()
}

func (a *App) registerHandlers() {
	a.handlers.Register(a.echo)
}

func (a *App) start(_ context.Context) error {
	address := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
	a.logger.Info("starting server", zap.String("address", address))

	return a.echo.Start(address)
}

func (a *App) stop(ctx context.Context) error {
	return a.echo.Shutdown(ctx)
}

func (a *App) customErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := 500
		message := "Internal Server Error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = he.Message.(string)
		}

		a.logger.Error("request error",
			zap.Int("status", code),
			zap.String("message", message),
			zap.Error(err),
		)

		if !c.Response().Committed {
			if err := c.JSON(code, map[string]string{
				"error": message,
			}); err != nil {
				a.logger.Error("error sending error response", zap.Error(err))
			}
		}
	}
}

// Module returns the application fx module
func Module() fx.Option {
	return NewModule()
}
