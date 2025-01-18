package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jonesrussell/goforms/internal/config/server"
	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// Server handles HTTP server lifecycle
type Server struct {
	echo        *echo.Echo
	logger      logger.Logger
	config      *server.Config
	shutdownCh  chan struct{}
	serverError chan error
}

// New creates a new server instance and registers lifecycle hooks with fx
func New(lc fx.Lifecycle, e *echo.Echo, log logger.Logger, cfg *server.Config) *Server {
	srv := &Server{
		echo:        e,
		logger:      log,
		config:      cfg,
		shutdownCh:  make(chan struct{}),
		serverError: make(chan error, 1),
	}

	lc.Append(fx.Hook{
		OnStart: srv.Start,
		OnStop:  srv.Stop,
	})

	return srv
}

// Start begins the server
func (s *Server) Start(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.logger.Info("server configuration",
		logger.String("bind_address", address),
		logger.String("host", s.config.Host),
		logger.Int("port", s.config.Port),
		logger.Duration("read_timeout", s.config.Timeouts.Read),
		logger.Duration("write_timeout", s.config.Timeouts.Write),
		logger.Duration("idle_timeout", s.config.Timeouts.Idle),
	)

	// Configure server timeouts
	s.echo.Server.ReadTimeout = s.config.Timeouts.Read
	s.echo.Server.WriteTimeout = s.config.Timeouts.Write
	s.echo.Server.IdleTimeout = s.config.Timeouts.Idle

	go func() {
		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			s.serverError <- err
			s.logger.Error("server error",
				logger.Error(err),
				logger.String("bind_address", address),
			)
		}
	}()

	// Monitor for server errors only, not context cancellation
	go func() {
		if err := <-s.serverError; err != nil {
			s.logger.Error("server error detected",
				logger.Error(err),
				logger.String("bind_address", address),
			)
			close(s.shutdownCh)
		}
	}()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("graceful shutdown initiated")

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Wait for in-flight requests to complete
	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("shutdown error", logger.Error(err))
		return err
	}

	close(s.shutdownCh)
	s.logger.Info("server stopped successfully")

	return nil
}
