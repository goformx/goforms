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
	addr   string
}

// New creates a new server instance with the provided dependencies
func New(lc fx.Lifecycle, logger logging.Logger, cfg *config.Config, e *echo.Echo) *Server {
	srv := &Server{
		echo:   e,
		logger: logger,
		config: cfg,
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
	s.addr = fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.logger.Info("starting server",
		logging.String("addr", s.addr),
		logging.String("env", s.config.App.Env),
	)

	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.echo,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	ln, listenErr := net.Listen("tcp", s.addr)
	if listenErr != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, listenErr)
	}

	go func() {
		s.logger.Info("server listening", logging.String("addr", s.addr))
		if serveErr := s.server.Serve(ln); serveErr != nil && serveErr != http.ErrServerClosed {
			s.logger.Error("server error", logging.Error(serveErr))
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.Server.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("server shutdown error", logging.Error(err))
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.logger.Info("server stopped")
	return nil
}
