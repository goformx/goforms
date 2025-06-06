package bootstrap

import (
	"context"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// provideEcho creates a new Echo server instance
func provideEcho(logger logging.Logger) (*echo.Echo, error) {
	logger.Info("Initializing Echo server")
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = middleware.NewValidator()
	logger.Info("Echo server initialized successfully")
	return e, nil
}

// configureMiddleware sets up the middleware on the Echo instance
func configureMiddleware(e *echo.Echo, mwManager *middleware.Manager, logger logging.Logger) error {
	logger.Info("Configuring middleware")
	mwManager.Setup(e)
	logger.Info("Middleware configuration completed")
	return nil
}

// configureServerLifecycle sets up the server lifecycle hooks
func configureServerLifecycle(lc fx.Lifecycle, srv *server.Server, logger logging.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting server via lifecycle hook")
			return srv.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping server via lifecycle hook")
			return srv.Stop(ctx)
		},
	})
}

// ServerProviders returns all the server-related providers
func ServerProviders() []fx.Option {
	return []fx.Option{
		fx.Provide(
			provideEcho,
		),
		fx.Invoke(configureMiddleware),
		fx.Invoke(configureServerLifecycle),
	}
}
