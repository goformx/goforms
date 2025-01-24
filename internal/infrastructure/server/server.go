package server

import (
	"context"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

type Server struct {
	echo   *echo.Echo
	logger logging.Logger
}

func New(logger logging.Logger) *Server {
	return &Server{
		echo:   echo.New(),
		logger: logger,
	}
}

func Start(lc fx.Lifecycle, s *Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := s.echo.Start(":8080"); err != nil {
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
