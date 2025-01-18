package server

import (
	"context"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/logger"
)

type Server struct {
	echo   *echo.Echo
	logger logger.Logger
}

func New(logger logger.Logger) *Server {
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
					s.logger.Error("server failed to start", logger.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return s.echo.Shutdown(ctx)
		},
	})
}
