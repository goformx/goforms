package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jonesrussell/goforms/internal/config/server"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Server handles HTTP server lifecycle
type Server struct {
	echo        *echo.Echo
	logger      *zap.Logger
	config      *server.Config
	serverError chan error
}

// New creates a new server instance
func New(e *echo.Echo, logger *zap.Logger, cfg *server.Config) *Server {
	return &Server{
		echo:        e,
		logger:      logger,
		config:      cfg,
		serverError: make(chan error, 1),
	}
}

// Start begins the server
func (s *Server) Start(ctx context.Context) error {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.logger.Info("server configuration",
		zap.String("bind_address", address),
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port),
		zap.Duration("read_timeout", s.config.Timeouts.Read),
		zap.Duration("write_timeout", s.config.Timeouts.Write),
		zap.Duration("idle_timeout", s.config.Timeouts.Idle),
	)

	go func() {
		if err := s.echo.Start(address); err != nil {
			if err != http.ErrServerClosed {
				s.serverError <- err
				s.logger.Error("server error",
					zap.Error(err),
					zap.String("bind_address", address),
				)
			}
		}
	}()

	// Monitor for server errors
	go func() {
		select {
		case err := <-s.serverError:
			s.logger.Error("server error detected",
				zap.Error(err),
				zap.String("bind_address", address),
			)
		case <-ctx.Done():
			s.logger.Info("server shutdown initiated",
				zap.String("bind_address", address),
			)
			return
		}
	}()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.echo.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("shutdown error", zap.Error(err))
		return err
	}

	return nil
}
