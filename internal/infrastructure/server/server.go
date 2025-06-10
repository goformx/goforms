package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

const (
	// DefaultReadHeaderTimeout is the default timeout for reading request headers
	DefaultReadHeaderTimeout = 5 * time.Second
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 10 * time.Second
	// DefaultStartupTimeout is the default timeout for server startup
	DefaultStartupTimeout = 5 * time.Second
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
	return net.JoinHostPort(s.config.App.Host, strconv.Itoa(s.config.App.Port))
}

// URL returns the server's full HTTP URL
func (s *Server) URL() string {
	return fmt.Sprintf("http://%s", s.Address())
}

// Start starts the server and returns when it's ready to accept connections
func (s *Server) Start() error {
	addr := s.Address()
	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.echo,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
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
		s.logger.Info("server started",
			"host", s.config.App.Host,
			"port", s.config.App.Port,
			"environment", s.config.App.Env,
			"version", s.config.App.Version)
		return nil
	case <-time.After(DefaultStartupTimeout):
		return fmt.Errorf("server startup timed out after %v", DefaultStartupTimeout)
	}
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
		"host", cfg.App.Host,
		"port", cfg.App.Port,
		"environment", cfg.App.Env,
		"server_type", "echo")

	// Add health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Serve static files from public directory with proper security headers
	e.Static("/static", "public")
	e.Static("/assets", "public/assets")

	// Vite dev server proxy for development mode
	if cfg.App.IsDevelopment() {
		viteURL, err := url.Parse(fmt.Sprintf("http://%s",
			net.JoinHostPort(cfg.App.ViteDevHost, cfg.App.ViteDevPort),
		))
		if err != nil {
			logger.Error("failed to parse Vite dev server URL",
				"error", err,
				"host", cfg.App.ViteDevHost,
				"port", cfg.App.ViteDevPort)
		} else {
			viteProxy := httputil.NewSingleHostReverseProxy(viteURL)

			// Configure proxy to handle WebSocket connections and CORS
			viteProxy.ModifyResponse = func(resp *http.Response) error {
				resp.Header.Set("Access-Control-Allow-Origin", "*")
				resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				return nil
			}

			// Proxy all asset requests to Vite dev server
			e.Group("/assets").Any("/*", echo.WrapHandler(viteProxy))
			e.Group("/src").Any("/*", echo.WrapHandler(viteProxy))
			e.Group("/@vite").Any("/*", echo.WrapHandler(viteProxy))
			e.Group("/node_modules").Any("/*", echo.WrapHandler(viteProxy))
			e.Group("/js").Any("/*", echo.WrapHandler(viteProxy))

			logger.Info("Vite dev server proxy configured", "url", viteURL.String())
		}
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

			shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
			defer cancel()

			if err := srv.server.Shutdown(shutdownCtx); err != nil {
				srv.logger.Error("server shutdown error", "error", err, "timeout", DefaultShutdownTimeout)
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
