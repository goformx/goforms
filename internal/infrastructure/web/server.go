// Package web provides utilities for handling web assets in the application.
package web

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// DevelopmentAssetServer implements AssetServer for development mode
// It serves static files from the public directory but doesn't proxy Vite routes
type DevelopmentAssetServer struct {
	config *config.Config
	logger logging.Logger
}

// NewDevelopmentAssetServer creates a new development asset server
func NewDevelopmentAssetServer(cfg *config.Config, logger logging.Logger) *DevelopmentAssetServer {
	return &DevelopmentAssetServer{
		config: cfg,
		logger: logger,
	}
}

// RegisterRoutes registers routes for static files in development
func (s *DevelopmentAssetServer) RegisterRoutes(e *echo.Echo) error {
	if s.config == nil {
		return errors.New("config is required")
	}

	// Add static file headers middleware
	e.Use(setupStaticFileHeaders)

	// Serve static files from the public directory
	publicDir := "public"
	if _, err := os.Stat(publicDir); err != nil {
		return fmt.Errorf("public directory not found: %w", err)
	}

	// Create file server for public directory
	fileServer := http.FileServer(http.Dir(publicDir))

	// Serve favicon.ico
	e.GET("/favicon.ico", echo.WrapHandler(fileServer))

	// Serve robots.txt
	e.GET("/robots.txt", echo.WrapHandler(fileServer))

	// Serve fonts
	e.GET("/fonts/*", echo.WrapHandler(
		http.StripPrefix("/fonts/", http.FileServer(http.Dir(filepath.Join(publicDir, "fonts")))),
	))

	// Serve Form.io fonts from the expected path to fix 404 errors
	e.GET("/node_modules/@formio/js/dist/fonts/*", echo.WrapHandler(
		http.StripPrefix(
			"/node_modules/@formio/js/dist/fonts/",
			http.FileServer(http.Dir(filepath.Join(publicDir, "fonts"))),
		),
	))

	s.logger.Info("development asset server configured",
		"public_dir", publicDir,
	)

	return nil
}

// EmbeddedAssetServer implements AssetServer for embedded static files in production
type EmbeddedAssetServer struct {
	logger logging.Logger
	distFS embed.FS
}

// NewEmbeddedAssetServer creates a new embedded asset server
func NewEmbeddedAssetServer(logger logging.Logger, distFS embed.FS) *EmbeddedAssetServer {
	return &EmbeddedAssetServer{
		logger: logger,
		distFS: distFS,
	}
}

// RegisterRoutes registers the embedded static file serving routes
func (s *EmbeddedAssetServer) RegisterRoutes(e *echo.Echo) error {
	// Add static file headers middleware
	e.Use(setupStaticFileHeaders)

	// Create a sub-filesystem for the dist directory
	distSubFS, err := fs.Sub(s.distFS, "dist")
	if err != nil {
		return fmt.Errorf("failed to create dist sub-filesystem: %w", err)
	}

	// Create a sub-filesystem for the assets directory
	assetsSubFS, err := fs.Sub(distSubFS, "assets")
	if err != nil {
		return fmt.Errorf("failed to create assets sub-filesystem: %w", err)
	}

	// Create a sub-filesystem for the fonts directory
	fontsSubFS, err := fs.Sub(distSubFS, "fonts")
	if err != nil {
		return fmt.Errorf("failed to create fonts sub-filesystem: %w", err)
	}

	// Create file server for embedded assets
	assetHandler := http.FileServer(http.FS(assetsSubFS))
	fontHandler := http.FileServer(http.FS(fontsSubFS))

	// Serve assets using the file server - strip the /assets prefix and serve from assets directory
	e.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", assetHandler)))

	// Serve fonts using the file server - strip the /assets/fonts prefix and serve from fonts directory
	e.GET("/assets/fonts/*", echo.WrapHandler(http.StripPrefix("/assets/fonts/", fontHandler)))

	// Serve Form.io fonts from the expected path to fix 404 errors
	e.GET("/node_modules/@formio/js/dist/fonts/*", echo.WrapHandler(
		http.StripPrefix("/node_modules/@formio/js/dist/fonts/", fontHandler),
	))

	// Serve individual files from embedded filesystem
	e.GET("/robots.txt", func(c echo.Context) error {
		data, readErr := fs.ReadFile(distSubFS, "robots.txt")
		if readErr != nil {
			return c.NoContent(http.StatusNotFound)
		}

		c.Response().Header().Set("Content-Type", "text/plain")

		return c.Blob(http.StatusOK, "text/plain", data)
	})
	e.GET("/favicon.ico", func(c echo.Context) error {
		data, readErr := fs.ReadFile(distSubFS, "favicon.ico")
		if readErr != nil {
			return c.NoContent(http.StatusNotFound)
		}

		c.Response().Header().Set("Content-Type", "image/x-icon")

		return c.Blob(http.StatusOK, "image/x-icon", data)
	})

	s.logger.Info("embedded asset server configured",
		"base_dir", "dist",
	)

	return nil
}

// setupStaticFileHeaders adds security headers for static files
func setupStaticFileHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Set security headers for static files
		c.Response().Header().Set("X-Content-Type-Options", "nosniff")
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000")

		return next(c)
	}
}
