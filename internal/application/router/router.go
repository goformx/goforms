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
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if strings.HasPrefix(path, cfg.Static.Path) {
				if strings.HasSuffix(path, ".css") {
					c.Response().Header().Set(echo.HeaderContentType, "text/css")
				} else if strings.HasSuffix(path, ".js") {
					c.Response().Header().Set(echo.HeaderContentType, "application/javascript")
				} else if path == "/favicon.ico" {
					c.Response().Header().Set(echo.HeaderContentType, "image/x-icon")
				} else if path == "/robots.txt" {
					c.Response().Header().Set(echo.HeaderContentType, "text/plain")
				}
			}
			return next(c)
		}
	})

	// Configure static files before any other routes
	if cfg.Static.Path != "" && cfg.Static.Root != "" {
		cfg.Logger.Debug("configuring static files",
			logging.String("path", cfg.Static.Path),
			logging.String("root", cfg.Static.Root),
		)
		
		// Create a group for static files that bypasses CSRF
		staticGroup := e.Group("")
		staticGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				// Set a flag to skip CSRF for static files
				c.Set("skip_csrf", true)
				return next(c)
			}
		})
		
		// Configure static routes in the static group
		staticGroup.Static(cfg.Static.Path, cfg.Static.Root)
		staticGroup.Static(cfg.Static.Path+"/dist", cfg.Static.Root+"/dist")
		staticGroup.File("/favicon.ico", cfg.Static.Root+"/favicon.ico")
		staticGroup.File("/robots.txt", cfg.Static.Root+"/robots.txt")
		
		cfg.Logger.Debug("static file configuration complete",
			logging.String("additional_paths", "/dist, /favicon.ico, /robots.txt"))
	}

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
