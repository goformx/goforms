// Package web provides utilities for handling web assets in the application.
package web

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// ViteAssetServer implements AssetServer for Vite development server
type ViteAssetServer struct {
	config *config.Config
	logger logging.Logger
}

// NewViteAssetServer creates a new Vite asset server
func NewViteAssetServer(cfg *config.Config, logger logging.Logger) *ViteAssetServer {
	return &ViteAssetServer{
		config: cfg,
		logger: logger,
	}
}

// RegisterRoutes registers the Vite dev server proxy routes
func (s *ViteAssetServer) RegisterRoutes(e *echo.Echo) error {
	if s.config == nil {
		return errors.New("config is required")
	}

	viteURL := fmt.Sprintf("%s://%s:%s", s.config.App.Scheme, s.config.App.ViteDevHost, s.config.App.ViteDevPort)
	parsedURL, err := url.Parse(viteURL)
	if err != nil {
		return fmt.Errorf("invalid Vite dev server URL: %w", err)
	}

	// Create a proxy for the Vite dev server
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// Only log the path portion of the URL, limited to 100 characters
		path := r.URL.Path
		if len(path) > MaxPathLength {
			path = path[:MaxPathLength] + "..."
		}
		s.logger.Error("proxy error",
			"error", err,
			"path", path,
		)
		http.Error(w, "Proxy Error", http.StatusBadGateway)
	}

	// Register routes for the Vite dev server
	e.Any("/src/*", echo.WrapHandler(proxy))
	e.Any("/@vite/*", echo.WrapHandler(proxy))
	e.Any("/@fs/*", echo.WrapHandler(proxy))
	e.Any("/@id/*", echo.WrapHandler(proxy))
	e.Any("/favicon.ico", echo.WrapHandler(proxy))

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
