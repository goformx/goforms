package router

import (
	"fmt"

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

	// Configure static files
	if cfg.Static.Path != "" && cfg.Static.Root != "" {
		cfg.Logger.Debug("configuring static files",
			logging.String("path", cfg.Static.Path),
			logging.String("root", cfg.Static.Root),
		)
		e.Static(cfg.Static.Path, cfg.Static.Root)
	}

	return nil
}
