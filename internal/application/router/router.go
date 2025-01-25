package router

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/application/handler"
)

// Config holds router configuration
type Config struct {
	Handlers []handler.Handler
	Static   StaticConfig
}

// StaticConfig configures static file serving
type StaticConfig struct {
	Path string
	Root string
}

// Setup configures all routes for an Echo instance
func Setup(e *echo.Echo, cfg *Config) {
	// Register API handlers
	for _, h := range cfg.Handlers {
		h.Register(e)
	}

	// Configure static files
	if cfg.Static.Path != "" {
		e.Static(cfg.Static.Path, cfg.Static.Root)
	}
}
