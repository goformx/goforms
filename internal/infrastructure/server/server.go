// Package server provides HTTP server setup and lifecycle management for the application.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/version"
	"github.com/goformx/goforms/internal/infrastructure/web"
)

const (
	// StartupTimeout is the timeout for server startup
	StartupTimeout = 5 * time.Second
	// ShutdownTimeout is the timeout for graceful shutdown
	ShutdownTimeout = 10 * time.Second
)

// Server handles HTTP server lifecycle and configuration
// Implements ServerInterface
type Server struct {
	echo   *echo.Echo
	logger logging.Logger
	config *config.Config
	server *http.Server
}

// Ensure Server implements ServerInterface
var _ ServerInterface = (*Server)(nil)

// URL returns the server's full HTTP URL
func (s *Server) URL() string {
	return s.config.App.GetServerURL()
}

// Start starts the server and returns when it's ready to accept connections
func (s *Server) Start() error {
	// Extract host and port from the URL for the HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.App.Host, s.config.App.Port)

	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.echo,
		ReadTimeout:       s.config.App.ReadTimeout,
		WriteTimeout:      s.config.App.WriteTimeout,
		IdleTimeout:       s.config.App.IdleTimeout,
		ReadHeaderTimeout: s.config.App.ReadTimeout,
	}

	// Create channels for server startup coordination
	started := make(chan struct{})
	errored := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		// Create a listener to check if the server can bind to the port
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			errored <- fmt.Errorf("failed to create listener: %w", err)

			return
		}

		// Signal that the server is ready to accept connections
		close(started)

		// Start serving
		if serveErr := s.server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			errored <- fmt.Errorf("server error: %w", serveErr)
		}
	}()

	// Wait for the server to be ready or fail
	select {
	case err := <-errored:
		return fmt.Errorf("server failed to start: %w", err)
	case <-started:
		versionInfo := version.GetInfo()
		s.logger.Info("server started",
			"host", s.config.App.Host,
			"port", s.config.App.Port,
			"environment", s.config.App.Environment,
			"version", versionInfo.Version,
			"build_time", versionInfo.BuildTime,
			"git_commit", versionInfo.GitCommit)

		return nil
	case <-time.After(StartupTimeout):
		return errors.New("server startup timed out after 5 seconds")
	}
}

// Deps contains the dependencies for creating a server
type Deps struct {
	fx.In
	Lifecycle   fx.Lifecycle
	Logger      logging.Logger
	Config      *config.Config
	Echo        *echo.Echo
	AssetServer web.AssetServer
}

// New creates a new server instance with the provided dependencies
func New(deps Deps) ServerInterface {
	srv := &Server{
		echo:   deps.Echo,
		logger: deps.Logger,
		config: deps.Config,
	}

	// Log server configuration
	deps.Logger.Info("initializing server",
		"url", srv.URL(),
		"environment", deps.Config.App.Environment,
		"server_type", "echo")

	// Register asset routes
	if err := deps.AssetServer.RegisterRoutes(deps.Echo); err != nil {
		deps.Logger.Error("failed to register asset routes", "error", err)
	}

	// Register lifecycle hooks
	deps.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			return nil // Server will be started after middleware is registered
		},
		OnStop: func(ctx context.Context) error {
			if srv.server == nil {
				return nil
			}

			srv.logger.Info("shutting down server")

			// Use configurable shutdown timeout if available, otherwise use default
			shutdownTimeout := ShutdownTimeout
			if srv.config.App.ShutdownTimeout > 0 {
				shutdownTimeout = srv.config.App.ShutdownTimeout
			}

			shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
			defer cancel()

			if err := srv.server.Shutdown(shutdownCtx); err != nil {
				srv.logger.Error("server shutdown error", "error", err, "timeout", shutdownTimeout)

				return fmt.Errorf("server shutdown error: %w", err)
			}

			srv.logger.Info("server stopped gracefully")

			return nil
		},
	})

	return srv
}

// Echo returns the underlying echo instance
func (s *Server) Echo() *echo.Echo {
	return s.echo
}

// Config returns the server configuration
func (s *Server) Config() *config.Config {
	return s.config
}
