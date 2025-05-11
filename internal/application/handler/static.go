package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/web"
)

// StaticHandler handles serving static files
type StaticHandler struct {
	logger    logging.Logger
	config    *config.Config
	fileIndex map[string]string // base name -> full path
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(logger logging.Logger, cfg *config.Config) *StaticHandler {
	handler := &StaticHandler{
		logger:    logger,
		config:    cfg,
		fileIndex: make(map[string]string),
	}
	// Build file index for dist directory
	distDir := cfg.Static.DistDir
	if _, statErr := os.Stat(distDir); statErr == nil {
		walkErr := filepath.Walk(distDir, func(walkPath string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !info.IsDir() {
				baseName := strings.Split(info.Name(), ".")[0]
				handler.fileIndex[baseName] = walkPath
			}
			return nil
		})
		if walkErr != nil {
			logger.Error("error walking distDir", logging.ErrorField("error", walkErr))
		}
	}
	return handler
}

// Register sets up routes for static file serving
func (h *StaticHandler) Register(e *echo.Echo) {
	// Handle Chrome DevTools well-known route
	e.GET("/.well-known/appspecific/com.chrome.devtools.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"devtoolsFrontendUrl":  "",
			"faviconUrl":           "/favicon.ico",
			"id":                   "goforms",
			"title":                "GoForms",
			"type":                 "node",
			"url":                  "/",
			"webSocketDebuggerUrl": "",
		})
	})

	// In production, serve from the dist directory
	distDir := h.config.Static.DistDir
	if _, err := os.Stat(distDir); err == nil {
		h.logger.Info("serving static files from dist directory",
			logging.StringField("dir", distDir),
		)
		// Use a wildcard route to handle hashed filenames
		prefix := "/" + distDir + "/*"
		e.GET(prefix, h.HandleStatic)
	}
}

// HandleStatic serves static files
func (h *StaticHandler) HandleStatic(c echo.Context) error {
	path := c.Param("*")
	if path == "" {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}

	// In production, serve from the dist directory
	requestBaseName := strings.Split(filepath.Base(path), ".")[0]
	foundFile, ok := h.fileIndex[requestBaseName]
	if !ok {
		h.logger.Error("file not found",
			logging.StringField("path", path),
			logging.StringField("distDir", h.config.Static.DistDir),
		)
		accept := c.Request().Header.Get("Accept")
		if strings.Contains(accept, "text/html") {
			// Try to serve a custom 404 page
			notFoundPage := filepath.Join(h.config.Static.DistDir, "404.html")
			if _, err := os.Stat(notFoundPage); err == nil {
				return c.File(notFoundPage)
			}
		}
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	// Set appropriate content type based on file extension
	ext := filepath.Ext(foundFile)
	switch ext {
	case ".css":
		c.Response().Header().Set("Content-Type", "text/css")
	case ".js":
		c.Response().Header().Set("Content-Type", "application/javascript")
	case ".map":
		c.Response().Header().Set("Content-Type", "application/json")
	}

	// Set cache headers based on manifest presence
	if _, manifestOk := web.Manifest[path]; manifestOk {
		// File is in the manifest, safe to cache long-term
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		// Not in manifest, use a short cache
		c.Response().Header().Set("Cache-Control", "public, max-age=3600")
	}

	h.logger.Debug("serving static file",
		logging.StringField("path", path),
		logging.StringField("distDir", h.config.Static.DistDir),
	)

	return c.File(foundFile)
}
