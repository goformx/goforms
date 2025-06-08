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

	return srv
}

// Echo returns the underlying echo instance
func (s *Server) Echo() *echo.Echo {
	return s.echo
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting server")

	// Create HTTP server
	addr := net.JoinHostPort(s.config.App.Host, fmt.Sprintf("%d", s.config.App.Port))
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.echo,
	}

	// Start server in a goroutine
	go func() {
		s.logger.Info("server listening",
			logging.StringField("address", addr),
			logging.StringField("environment", s.config.App.Env),
		)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("failed to start server",
				logging.ErrorField("error", err),
				logging.StringField("address", addr),
			)
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
		s.logger.Error("server shutdown error", logging.ErrorField("error", err))
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.logger.Info("server stopped")
	return nil
}
