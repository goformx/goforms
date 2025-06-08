package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Server handles HTTP server lifecycle and configuration
type Server struct {
	echo   *echo.Echo
	logger logging.Logger
	config *config.Config
	server *http.Server
}

// Address returns the server's address in host:port format
func (s *Server) Address() string {
	return net.JoinHostPort(s.config.App.Host, fmt.Sprintf("%d", s.config.App.Port))
}

// URL returns the server's full HTTP URL
func (s *Server) URL() string {
	return fmt.Sprintf("http://%s", s.Address())
}

// Start starts the server and returns when it's ready to accept connections
func (s *Server) Start() error {
	addr := s.Address()
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.echo,
	}

	// Create a channel to signal when the server is ready
	ready := make(chan struct{})

	// Start server in a goroutine
	go func() {
		// Signal that the server is ready to accept connections
		close(ready)

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("failed to start server",
				logging.ErrorField("error", err),
				logging.StringField("address", addr),
			)
		}
	}()

	// Wait for the server to be ready
	<-ready

	return nil
}

// New creates a new server instance with the provided dependencies
func New(
	lc fx.Lifecycle,
	logger logging.Logger,
	cfg *config.Config,
	e *echo.Echo,
	middlewareManager *middleware.Manager,
) *Server {
	srv := &Server{
		echo:   e,
		logger: logger,
		config: cfg,
	}

	// Log server configuration
	logger.Info("initializing server",
		logging.StringField("host", cfg.App.Host),
		logging.IntField("port", cfg.App.Port),
		logging.StringField("environment", cfg.App.Env),
		logging.StringField("server_type", "echo"),
	)

	// Serve static files from public directory
	e.Static("/", "public")

	// Vite dev server proxy for /src and /@vite in development mode
	if cfg.App.IsDevelopment() {
		viteURL := &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(cfg.App.ViteDevHost, cfg.App.ViteDevPort),
		}
		viteProxy := httputil.NewSingleHostReverseProxy(viteURL)

		// Configure proxy to handle WebSocket connections
		viteProxy.ModifyResponse = func(resp *http.Response) error {
			resp.Header.Set("Access-Control-Allow-Origin", "*")
			return nil
		}

		// Proxy all static asset requests to Vite dev server
		e.Group("/src").Any("/*", echo.WrapHandler(viteProxy))
		e.Group("/@vite").Any("/*", echo.WrapHandler(viteProxy))
		e.Group("/assets").Any("/*", echo.WrapHandler(viteProxy))
		e.Group("/node_modules").Any("/*", echo.WrapHandler(viteProxy))
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil // Server will be started after middleware is registered
		},
		OnStop: func(ctx context.Context) error {
			if srv.server == nil {
				return nil
			}

			srv.logger.Info("shutting down server")

			shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
			defer cancel()

			if err := srv.server.Shutdown(shutdownCtx); err != nil {
				srv.logger.Error("server shutdown error", logging.ErrorField("error", err))
				return fmt.Errorf("server shutdown error: %w", err)
			}

			srv.logger.Info("server stopped")
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
