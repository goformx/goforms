package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Server handles HTTP server lifecycle and configuration
type Server struct {
	echo   *echo.Echo
	logger logging.Logger
	config *config.Config
	server *http.Server
}

// New creates a new server instance with the provided dependencies
func New(lc fx.Lifecycle, logger logging.Logger, config *config.Config) *Server {
	e := echo.New()
	srv := &Server{
		echo:   e,
		logger: logger,
		config: config,
	}

	// Setup server lifecycle hooks
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

// Echo returns the underlying echo instance
func (s *Server) Echo() *echo.Echo {
	return s.echo
}

// Start initializes and starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.App.Host, s.config.App.Port)
	s.logger.Info("starting server",
		logging.String("addr", addr),
		logging.String("env", s.config.App.Env),
	)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.echo,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	// Create a listener first
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Start server in background
	go func() {
		s.logger.Info("server listening", logging.String("addr", addr))
		if err := s.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error", logging.Error(err))
		}
	}()

	// Signal that we're ready
	s.logger.Info("server ready")

	// Wait for context cancellation
	<-ctx.Done()
	return s.Stop(context.Background())
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.logger.Info("shutting down server")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.Server.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("server shutdown error", logging.Error(err))
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.logger.Info("server stopped")
	return nil
}
