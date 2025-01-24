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
}

// New creates a new server instance
func New(lc fx.Lifecycle, e *echo.Echo, log logging.Logger, cfg *config.Config) *Server {
	srv := &Server{
		echo:         e,
		logger:       log,
		appConfig:    &cfg.App,
		serverConfig: &cfg.Server,
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

	// Start server in background
	go func() {
		s.logger.Info("server listening", logging.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", logging.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	s.logger.Info("shutting down server")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.serverConfig.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("server shutdown error", logging.Error(err))
		return err
	}

	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")
	return s.echo.Shutdown(ctx)
}
