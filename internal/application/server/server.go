package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// TimeoutConfig holds server timeout settings
type TimeoutConfig struct {
	Read  time.Duration
	Write time.Duration
	Idle  time.Duration
}

// Config holds server configuration
type Config struct {
	Host     string
	Port     int
	Timeouts TimeoutConfig
}

// Server handles HTTP server lifecycle
type Server struct {
	echo        *echo.Echo
	logger      logging.Logger
	config      *Config
	shutdownCh  chan struct{}
	serverError chan error
}

// New creates a new server instance and registers lifecycle hooks with fx
func New(lc fx.Lifecycle, e *echo.Echo, log logging.Logger, cfg *Config) *Server {
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
		logging.String("bind_address", address),
		logging.String("host", s.config.Host),
		logging.Int("port", s.config.Port),
		logging.Duration("read_timeout", s.config.Timeouts.Read),
		logging.Duration("write_timeout", s.config.Timeouts.Write),
		logging.Duration("idle_timeout", s.config.Timeouts.Idle),
	)

	// Configure server timeouts
	s.echo.Server.ReadTimeout = s.config.Timeouts.Read
	s.echo.Server.WriteTimeout = s.config.Timeouts.Write
	s.echo.Server.IdleTimeout = s.config.Timeouts.Idle

	go func() {
		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			s.serverError <- err
			s.logger.Error("server error",
				logging.Error(err),
				logging.String("bind_address", address),
			)
		}
	}()

	// Monitor for server errors only, not context cancellation
	go func() {
		if err := <-s.serverError; err != nil {
			s.logger.Error("server error detected",
				logging.Error(err),
				logging.String("bind_address", address),
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
		s.logger.Error("shutdown error", logging.Error(err))
		return err
	}

	close(s.shutdownCh)
	s.logger.Info("server stopped successfully")

	return nil
}

// Start is used by fx.Invoke to create and start the server
func Start(e *echo.Echo, log logging.Logger, cfg *Config, lc fx.Lifecycle) {
	_ = New(lc, e, log, cfg)
}
