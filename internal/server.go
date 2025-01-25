package internal

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/http"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

func registerServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger logging.Logger,
	handlers *http.Handlers,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Register all routes
	handlers.Register(e)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				addr := fmt.Sprintf(":%d", cfg.Server.Port)
				if err := e.Start(addr); err != nil {
					logger.Error("failed to start server", logging.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return e.Shutdown(ctx)
		},
	})

	return e
}
