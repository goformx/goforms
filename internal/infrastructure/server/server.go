package server

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

type Server struct {
	echo   *echo.Echo
	logger logging.Logger
	config *config.Config
}

func New(logger logging.Logger, cfg *config.Config) *Server {
	return &Server{
		echo:   echo.New(),
		logger: logger,
		config: cfg,
	}
}

func Start(lc fx.Lifecycle, s *Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				addr := fmt.Sprintf("%s:%d", s.config.App.Host, s.config.App.Port)
				if err := s.echo.Start(addr); err != nil {
					s.logger.Error("server failed to start", logging.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return s.echo.Shutdown(ctx)
		},
	})
}
