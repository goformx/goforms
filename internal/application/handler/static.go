package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// StaticHandler handles serving static files
// It is the single source of truth for static file serving in the application.
type StaticHandler struct {
	logger logging.Logger
	config *config.Config
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(logger logging.Logger, cfg *config.Config) *StaticHandler {
	if cfg == nil {
		panic("config is required for StaticHandler")
	}
	if logger == nil {
		panic("logger is required for StaticHandler")
	}
	return &StaticHandler{
		logger: logger,
		config: cfg,
	}
}

// IsStaticFile checks if the given path is a static file
func (h *StaticHandler) IsStaticFile(path string) bool {
	// Skip TypeScript files in development mode
	if strings.HasSuffix(path, ".ts") {
		return false
	}

	// Skip Vite dev server paths
	if strings.HasPrefix(path, "/@vite/") {
		return false
	}

	// Skip source files in development mode
	if strings.HasPrefix(path, "/src/") {
		return false
	}

	// Skip assets in development mode
	if strings.HasPrefix(path, "/assets/") {
		return false
	}

	return strings.HasPrefix(path, "/public/") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".css")
}

// setMIMEType sets the appropriate Content-Type header based on the file extension
func (h *StaticHandler) setMIMEType(c echo.Context, path string) {
	switch {
	case strings.HasSuffix(path, ".css"):
		c.Response().Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(path, ".js"):
		c.Response().Header().Set("Content-Type", "application/javascript")
	case strings.HasSuffix(path, ".ts"):
		c.Response().Header().Set("Content-Type", "application/javascript")
	case strings.HasSuffix(path, ".mjs"):
		c.Response().Header().Set("Content-Type", "application/javascript")
	case strings.HasSuffix(path, ".ico"):
		c.Response().Header().Set("Content-Type", "image/x-icon")
	case strings.HasSuffix(path, ".txt"):
		c.Response().Header().Set("Content-Type", "text/plain")
	case strings.HasPrefix(path, "/@vite/"):
		c.Response().Header().Set("Content-Type", "application/javascript")
	}
}

// Register sets up routes for static file serving
func (h *StaticHandler) Register(e *echo.Echo) {
	// Handle Chrome DevTools well-known route only in development
	if h.config.App.IsDevelopment() {
		e.GET("/.well-known/appspecific/com.chrome.devtools.json", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]any{
				"devtoolsFrontendUrl":  "",
				"faviconUrl":           "/favicon.ico",
				"id":                   "goforms",
				"title":                "GoFormX",
				"type":                 "node",
				"url":                  "/",
				"webSocketDebuggerUrl": "",
			})
		})
	}

	// Serve static files using Echo's built-in middleware
	distDir := h.config.Static.DistDir
	if distDir == "" {
		h.logger.Error("static directory not configured",
			logging.StringField("config_key", "Static.DistDir"),
		)
		return
	}

	// Ensure the path is absolute
	absPath, err := filepath.Abs(distDir)
	if err != nil {
		h.logger.Error("failed to resolve static directory path",
			logging.StringField("dir", distDir),
			logging.ErrorField("error", err),
		)
		return
	}

	if stat, err := os.Stat(absPath); err == nil && stat.IsDir() {
		h.logger.Info("serving static files from dist directory",
			logging.StringField("dir", absPath),
		)

		// Add MIME type middleware for static files
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				path := c.Request().URL.Path
				if h.IsStaticFile(path) {
					h.setMIMEType(c, path)
				}
				return next(c)
			}
		})

		e.Static("/", absPath)
	} else {
		h.logger.Error("static directory not found or inaccessible",
			logging.StringField("dir", absPath),
			logging.ErrorField("error", err),
		)
	}
}
