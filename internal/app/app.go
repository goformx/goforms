package app

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/app/server"
	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/middleware"
)

type App struct {
	server     *server.Server
	middleware *middleware.Manager
	handlers   *handlers.SubscriptionHandler
}

func NewApp(
	lc fx.Lifecycle,
	logger *zap.Logger,
	echo *echo.Echo,
	cfg *config.Config,
	handler *handlers.SubscriptionHandler,
	healthHandler *handlers.HealthHandler,
) *App {
	mw := middleware.New(logger, cfg)
	srv := server.New(echo, logger, &cfg.Server)

	app := &App{
		server:     srv,
		middleware: mw,
		handlers:   handler,
	}

	// Setup order: middleware -> handlers -> lifecycle hooks
	mw.Setup(echo)
	handler.Register(echo)
	healthHandler.Register(echo)

	lc.Append(fx.Hook{
		OnStart: srv.Start,
		OnStop:  srv.Stop,
	})

	return app
}
