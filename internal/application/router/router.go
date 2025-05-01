package router

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Config holds router configuration
type Config struct {
	Handlers []handlers.Handler
	Static   StaticConfig
	Logger   logging.Logger
}

// StaticConfig configures static file serving
type StaticConfig struct {
	Path string
	Root string
}

// Setup configures all routes for an Echo instance
func Setup(e *echo.Echo, cfg *Config) error {
	if cfg.Logger == nil {
		logger, err := logging.NewTestLogger()
		if err != nil {
			return fmt.Errorf("failed to create test logger: %w", err)
		}
		cfg.Logger = logger
	}

	cfg.Logger.Debug("setting up routes",
		logging.Int("handler_count", len(cfg.Handlers)),
		logging.String("static_path", cfg.Static.Path),
		logging.String("static_root", cfg.Static.Root),
	)

	// Add MIME type middleware first
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if strings.HasSuffix(path, ".css") {
				c.Response().Header().Set("Content-Type", "text/css")
			} else if strings.HasSuffix(path, ".js") {
				c.Response().Header().Set("Content-Type", "application/javascript")
			} else if path == "/favicon.ico" {
				c.Response().Header().Set("Content-Type", "image/x-icon")
			} else if path == "/robots.txt" {
				c.Response().Header().Set("Content-Type", "text/plain")
			}
			return next(c)
		}
	})

	// Create static group that sets skip_csrf flag
	staticGroup := e.Group("")
	staticGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if strings.HasPrefix(path, "/static/") || 
			   path == "/favicon.ico" ||
			   path == "/robots.txt" {
				c.Set("skip_csrf", true)
				c.Set("skip_auth", true)
			}
			return next(c)
		}
	})

	// Configure static routes
	staticGroup.Static(cfg.Static.Path, cfg.Static.Root)
	staticGroup.Static("/static/dist", "./static/dist")
	staticGroup.File("/favicon.ico", "./static/favicon.ico")
	staticGroup.File("/robots.txt", "./static/robots.txt")

	// Create form group that ensures CSRF tokens are generated
	formGroup := e.Group("")
	formGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if strings.HasPrefix(path, "/login") || 
			   strings.HasPrefix(path, "/signup") || 
			   strings.HasPrefix(path, "/forgot-password") ||
			   strings.HasPrefix(path, "/contact") ||
			   strings.HasPrefix(path, "/demo") {
				// Ensure CSRF tokens are generated for form pages
				c.Set("skip_csrf", false)
			}
			return next(c)
		}
	})

	// Register API handlers
	for i, h := range cfg.Handlers {
		cfg.Logger.Debug("registering handler",
			logging.Int("index", i),
			logging.String("type", fmt.Sprintf("%T", h)),
		)
		h.Register(e)
		cfg.Logger.Debug("handler registered",
			logging.Int("index", i),
			logging.String("type", fmt.Sprintf("%T", h)),
		)
	}

	return nil
}
