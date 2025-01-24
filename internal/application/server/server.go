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
		logging.String("env", s.appConfig.Env),
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.appConfig.Host, s.appConfig.Port),
		Handler:      s.echo,
		ReadTimeout:  s.serverConfig.ReadTimeout,
		WriteTimeout: s.serverConfig.WriteTimeout,
		IdleTimeout:  s.serverConfig.IdleTimeout,
	}

	go func() {
		s.logger.Info("server listening", logging.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", logging.Error(err))
			s.serverError <- err
		}
	}()

	// Don't block on context.Done() - let the server run
	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")
	close(s.shutdownCh)
	return s.echo.Shutdown(ctx)
}
