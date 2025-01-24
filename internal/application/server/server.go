package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Server handles HTTP server lifecycle
type Server struct {
	echo         *echo.Echo
	logger       logging.Logger
	appConfig    *config.AppConfig
	serverConfig *config.ServerConfig
	shutdownCh   chan struct{}
	serverError  chan error
}

// New creates a new server instance
func New(lc fx.Lifecycle, e *echo.Echo, log logging.Logger, cfg *config.Config) *Server {
	srv := &Server{
		echo:         e,
		logger:       log,
		appConfig:    &cfg.App,
		serverConfig: &cfg.Server,
		shutdownCh:   make(chan struct{}),
		serverError:  make(chan error, 1),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return srv.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return srv.Stop(ctx)
		},
	})

	return srv
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting HTTP server",
		logging.String("host", s.appConfig.Host),
		logging.Int("port", s.appConfig.Port),
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.appConfig.Host, s.appConfig.Port),
		Handler:      s.echo,
		ReadTimeout:  s.serverConfig.ReadTimeout,
		WriteTimeout: s.serverConfig.WriteTimeout,
		IdleTimeout:  s.serverConfig.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.serverError <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-s.serverError:
		return fmt.Errorf("server error: %w", err)
	case <-s.shutdownCh:
		return nil
	}
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")
	close(s.shutdownCh)
	return s.echo.Shutdown(ctx)
}
