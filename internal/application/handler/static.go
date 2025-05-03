package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// StaticHandler handles serving static files
type StaticHandler struct {
	logger logging.Logger
	config *config.Config
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(logger logging.Logger, cfg *config.Config) *StaticHandler {
	return &StaticHandler{
		logger: logger,
		config: cfg,
	}
}

// Register sets up routes for static file serving
func (h *StaticHandler) Register(e *echo.Echo) {
	// In development mode, proxy static requests to Vite dev server
	if h.config.App.IsDevelopment() {
		h.logger.Info("development mode: proxying static files to Vite dev server")
		e.GET("/static/dist/*", func(c echo.Context) error {
			path := c.Param("*")
			url := "http://localhost:3000/static/dist/" + path
			resp, err := http.Get(url)
			if err != nil {
				return echo.NewHTTPError(http.StatusNotFound, "file not found")
			}
			defer resp.Body.Close()

			// Set appropriate content type based on file extension
			ext := filepath.Ext(path)
			switch ext {
			case ".css":
				c.Response().Header().Set("Content-Type", "text/css")
			case ".js":
				c.Response().Header().Set("Content-Type", "application/javascript")
			case ".map":
				c.Response().Header().Set("Content-Type", "application/json")
			default:
				// For other files, copy the content type from Vite
				if contentType := resp.Header.Get("Content-Type"); contentType != "" {
					c.Response().Header().Set("Content-Type", contentType)
				}
			}

			return c.Stream(resp.StatusCode, c.Response().Header().Get("Content-Type"), resp.Body)
		})
		return
	}

	// In production, serve from the dist directory
	distDir := filepath.Join("static", "dist")
	if _, err := os.Stat(distDir); err == nil {
		h.logger.Info("serving static files from dist directory",
			logging.String("dir", distDir),
		)
		// Use a wildcard route to handle hashed filenames
		e.GET("/static/dist/*", h.HandleStatic)
	}

	// Always serve static files from the static directory
	e.Static("/static", "static")
}

// HandleStatic serves static files
func (h *StaticHandler) HandleStatic(c echo.Context) error {
	path := c.Param("*")
	if path == "" {
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
	}

	// In production, serve from the dist directory
	distDir := filepath.Join("static", "dist")

	// Walk the dist directory to find the file with the matching base name
	var foundFile string
	err := filepath.Walk(distDir, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// Check if the file matches the requested path (ignoring the hash)
			baseName := strings.Split(info.Name(), ".")[0]
			requestBaseName := strings.Split(filepath.Base(path), ".")[0]
			if baseName == requestBaseName {
				foundFile = walkPath
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil || foundFile == "" {
		h.logger.Error("file not found",
			logging.String("path", path),
			logging.String("distDir", distDir),
		)
		return echo.NewHTTPError(http.StatusNotFound, "file not found")
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

	return c.File(foundFile)
}
