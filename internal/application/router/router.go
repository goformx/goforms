package router

import (
	"errors"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/handlers"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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

// isStaticFile determines if a path corresponds to a static file
func isStaticFile(path, distDir string) bool {
	// Ensure distDir always starts with a slash for prefix check
	prefix := distDir
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix += "/"
	return strings.HasPrefix(path, prefix) ||
		path == "/favicon.ico" ||
		path == "/robots.txt"
}

// setupMIMETypeMiddleware creates middleware for setting appropriate MIME types
func setupMIMETypeMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			switch {
			case strings.HasSuffix(path, ".css"):
				c.Response().Header().Set("Content-Type", "text/css")
			case strings.HasSuffix(path, ".js"):
				c.Response().Header().Set("Content-Type", "application/javascript")
			case path == "/favicon.ico":
				c.Response().Header().Set("Content-Type", "image/x-icon")
			case path == "/robots.txt":
				c.Response().Header().Set("Content-Type", "text/plain")
			}
			return next(c)
		}
	}
}

// logHandlerRegistration logs handler registration details
func logHandlerRegistration(logger logging.Logger, index int, handlerType string) {
	logger.Debug("registering handler",
		logging.IntField("index", index),
		logging.StringField("type", handlerType))
}

// setupStaticRoutes configures static file routes
func setupStaticRoutes(group interface {
	Static(prefix, root string)
	File(path, file string)
}, distDir string) {
	group.Static("/public", "public")
	group.Static("/dist", distDir)
	group.File("/favicon.ico", "./public/favicon.ico")
	group.File("/robots.txt", "./public/robots.txt")
}

// registerHandlers registers all API handlers
func registerHandlers(e *echo.Echo, handlerList []handlers.Handler, logger logging.Logger) {
	for i, h := range handlerList {
		logHandlerRegistration(logger, i, fmt.Sprintf("%T", h))
		h.Register(e)
		logger.Debug("handler registered",
			logging.IntField("index", i),
			logging.StringField("type", fmt.Sprintf("%T", h)),
		)
	}
}

// validateConfig checks if the configuration is valid
func validateConfig(cfg *Config) error {
	if cfg.Static.Path == "" || cfg.Static.Root == "" {
		return errors.New("static config must include both path and root")
	}
	return nil
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

	if err := validateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cfg.Logger.Debug("setting up routes",
		logging.IntField("handler_count", len(cfg.Handlers)),
		logging.StringField("static_path", cfg.Static.Path),
		logging.StringField("static_root", cfg.Static.Root),
	)

	// Setup middleware
	e.Pre(setupMIMETypeMiddleware())

	// Create static group that sets skip_csrf flag
	staticGroup := e.Group("")
	staticGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if isStaticFile(c.Request().URL.Path, cfg.Static.Root) {
				c.Set("skip_csrf", true)
				c.Set("skip_auth", true)
			}
			return next(c)
		}
	})

	// Setup routes
	// Use cfg.Static.Root for distDir
	distDir := cfg.Static.Root
	if distDir == "" {
		distDir = "dist"
	}
	setupStaticRoutes(staticGroup, distDir)

	// Register API handlers
	registerHandlers(e, cfg.Handlers, cfg.Logger)

	return nil
}
