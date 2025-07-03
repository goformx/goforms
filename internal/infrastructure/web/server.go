// Package web provides utilities for handling web assets in the application.
package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// AssetServerConfig holds configuration for asset servers
type AssetServerConfig struct {
	PublicDir       string
	MaxAge          time.Duration
	EnableGzip      bool
	EnableBrotli    bool
	SecurityHeaders map[string]string
	CustomMimeTypes map[string]string
}

// DefaultAssetServerConfig returns default configuration for asset servers
func DefaultAssetServerConfig() AssetServerConfig {
	return AssetServerConfig{
		PublicDir:    "public",
		MaxAge:       365 * 24 * time.Hour, // 1 year
		EnableGzip:   true,
		EnableBrotli: true,
		SecurityHeaders: map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "DENY",
			"Referrer-Policy":        "strict-origin-when-cross-origin",
		},
		CustomMimeTypes: map[string]string{
			".woff2": "font/woff2",
			".woff":  "font/woff",
			".ttf":   "font/ttf",
			".otf":   "font/otf",
		},
	}
}

// DevelopmentAssetServer implements AssetServer for development mode
type DevelopmentAssetServer struct {
	config       *config.Config
	logger       logging.Logger
	serverConfig AssetServerConfig
	isRunning    bool
}

// NewDevelopmentAssetServer creates a new development asset server
func NewDevelopmentAssetServer(cfg *config.Config, logger logging.Logger) *DevelopmentAssetServer {
	return &DevelopmentAssetServer{
		config:       cfg,
		logger:       logger,
		serverConfig: DefaultAssetServerConfig(),
		isRunning:    false,
	}
}

// WithConfig allows customizing the server configuration
func (s *DevelopmentAssetServer) WithConfig(cfg AssetServerConfig) *DevelopmentAssetServer {
	s.serverConfig = cfg

	return s
}

// RegisterRoutes registers routes for static files in development
func (s *DevelopmentAssetServer) RegisterRoutes(e *echo.Echo) error {
	if s.config == nil {
		return errors.New("config is required")
	}

	// Validate public directory exists
	if err := s.validatePublicDirectory(); err != nil {
		return fmt.Errorf("public directory validation failed: %w", err)
	}

	// Register security headers middleware
	e.Use(s.createSecurityHeadersMiddleware())

	// Register static file routes
	s.registerStaticRoutes(e)
	s.registerSpecialFileRoutes(e)
	s.registerFormioRoutes(e)

	s.logger.Info("development asset server configured",
		"public_dir", s.serverConfig.PublicDir,
		"max_age", s.serverConfig.MaxAge,
		"security_headers", len(s.serverConfig.SecurityHeaders))

	return nil
}

// Start initializes the development asset server
func (s *DevelopmentAssetServer) Start(ctx context.Context) error {
	s.logger.Info("starting development asset server")
	s.isRunning = true

	return nil
}

// Stop gracefully stops the development asset server
func (s *DevelopmentAssetServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping development asset server")
	s.isRunning = false

	return nil
}

// validatePublicDirectory ensures the public directory exists and is accessible
func (s *DevelopmentAssetServer) validatePublicDirectory() error {
	if s.serverConfig.PublicDir == "" {
		return errors.New("public directory path cannot be empty")
	}

	stat, err := os.Stat(s.serverConfig.PublicDir)
	if err != nil {
		return fmt.Errorf("public directory not accessible: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", s.serverConfig.PublicDir)
	}

	return nil
}

// registerStaticRoutes registers standard static file routes
func (s *DevelopmentAssetServer) registerStaticRoutes(e *echo.Echo) {
	// Fonts directory
	fontsPath := filepath.Join(s.serverConfig.PublicDir, "fonts")
	if _, err := os.Stat(fontsPath); err == nil {
		e.GET("/fonts/*", s.createFileHandler(
			http.StripPrefix("/fonts/", http.FileServer(http.Dir(fontsPath))),
		))
	}

	// Assets directory (if it exists in development)
	assetsPath := filepath.Join(s.serverConfig.PublicDir, "assets")
	if _, err := os.Stat(assetsPath); err == nil {
		e.GET("/assets/*", s.createFileHandler(
			http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsPath))),
		))
	}
}

// registerSpecialFileRoutes registers routes for special files like favicon and robots.txt
func (s *DevelopmentAssetServer) registerSpecialFileRoutes(e *echo.Echo) {
	specialFiles := map[string]string{
		"/favicon.ico": "favicon.ico",
		"/robots.txt":  "robots.txt",
	}

	for route, filename := range specialFiles {
		fullPath := filepath.Join(s.serverConfig.PublicDir, filename)
		if _, err := os.Stat(fullPath); err == nil {
			e.GET(route, s.createSpecialFileHandler(filename))
		}
	}
}

// registerFormioRoutes registers routes for Form.io compatibility
func (s *DevelopmentAssetServer) registerFormioRoutes(e *echo.Echo) {
	fontsPath := filepath.Join(s.serverConfig.PublicDir, "fonts")
	if _, err := os.Stat(fontsPath); err == nil {
		e.GET("/node_modules/@formio/js/dist/fonts/*", s.createFileHandler(
			http.StripPrefix("/node_modules/@formio/js/dist/fonts/",
				http.FileServer(http.Dir(fontsPath))),
		))
	}
}

// createFileHandler creates a handler that wraps an http.Handler with Echo
func (s *DevelopmentAssetServer) createFileHandler(handler http.Handler) echo.HandlerFunc {
	return echo.WrapHandler(handler)
}

// createSpecialFileHandler creates a handler for special files with proper MIME types
func (s *DevelopmentAssetServer) createSpecialFileHandler(filename string) echo.HandlerFunc {
	return func(c echo.Context) error {
		fullPath := filepath.Join(s.serverConfig.PublicDir, filename)

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}

		mimeType := s.detectMimeType(filename)
		c.Response().Header().Set("Content-Type", mimeType)

		return c.Blob(http.StatusOK, mimeType, data)
	}
}

// EmbeddedAssetServer implements AssetServer for embedded static files in production
type EmbeddedAssetServer struct {
	logger         logging.Logger
	distFS         embed.FS
	serverConfig   AssetServerConfig
	subFileSystems map[string]fs.FS
	isRunning      bool
}

// NewEmbeddedAssetServer creates a new embedded asset server
func NewEmbeddedAssetServer(logger logging.Logger, distFS embed.FS) *EmbeddedAssetServer {
	return &EmbeddedAssetServer{
		logger:         logger,
		distFS:         distFS,
		serverConfig:   DefaultAssetServerConfig(),
		subFileSystems: make(map[string]fs.FS),
		isRunning:      false,
	}
}

// WithConfig allows customizing the server configuration
func (s *EmbeddedAssetServer) WithConfig(cfg AssetServerConfig) *EmbeddedAssetServer {
	s.serverConfig = cfg

	return s
}

// RegisterRoutes registers the embedded static file serving routes
func (s *EmbeddedAssetServer) RegisterRoutes(e *echo.Echo) error {
	// Create sub-filesystems
	if err := s.createSubFileSystems(); err != nil {
		return fmt.Errorf("failed to create sub-filesystems: %w", err)
	}

	// Register security headers middleware
	e.Use(s.createSecurityHeadersMiddleware())

	// Register asset routes
	s.registerAssetRoutes(e)
	s.registerSpecialFileRoutes(e)
	s.registerFormioRoutes(e)

	s.logger.Info("embedded asset server configured",
		"filesystems", len(s.subFileSystems),
		"max_age", s.serverConfig.MaxAge)

	return nil
}

// Start initializes the embedded asset server
func (s *EmbeddedAssetServer) Start(ctx context.Context) error {
	s.logger.Info("starting embedded asset server")
	s.isRunning = true

	return nil
}

// Stop gracefully stops the embedded asset server
func (s *EmbeddedAssetServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping embedded asset server")
	s.isRunning = false

	return nil
}

// createSubFileSystems creates sub-filesystems for different asset types
func (s *EmbeddedAssetServer) createSubFileSystems() error {
	// Create dist sub-filesystem
	distSubFS, err := fs.Sub(s.distFS, "dist")
	if err != nil {
		return fmt.Errorf("failed to create dist sub-filesystem: %w", err)
	}

	s.subFileSystems["dist"] = distSubFS

	// Create assets sub-filesystem if it exists
	if assetsSubFS, assetsErr := fs.Sub(distSubFS, "assets"); assetsErr == nil {
		s.subFileSystems["assets"] = assetsSubFS
	}

	// Create fonts sub-filesystem if it exists
	if fontsSubFS, fontsErr := fs.Sub(distSubFS, "fonts"); fontsErr == nil {
		s.subFileSystems["fonts"] = fontsSubFS
	}

	return nil
}

// registerAssetRoutes registers routes for embedded assets
func (s *EmbeddedAssetServer) registerAssetRoutes(e *echo.Echo) {
	if assetsFS, exists := s.subFileSystems["assets"]; exists {
		assetHandler := http.FileServer(http.FS(assetsFS))
		e.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", assetHandler)))
	}

	if fontsFS, exists := s.subFileSystems["fonts"]; exists {
		fontHandler := http.FileServer(http.FS(fontsFS))
		e.GET("/assets/fonts/*", echo.WrapHandler(http.StripPrefix("/assets/fonts/", fontHandler)))
	}
}

// registerSpecialFileRoutes registers routes for special embedded files
func (s *EmbeddedAssetServer) registerSpecialFileRoutes(e *echo.Echo) {
	distFS := s.subFileSystems["dist"]

	specialFiles := map[string]string{
		"/robots.txt":  "robots.txt",
		"/favicon.ico": "favicon.ico",
	}

	for route, filename := range specialFiles {
		e.GET(route, s.createEmbeddedFileHandler(distFS, filename))
	}
}

// registerFormioRoutes registers Form.io compatibility routes
func (s *EmbeddedAssetServer) registerFormioRoutes(e *echo.Echo) {
	if fontsFS, exists := s.subFileSystems["fonts"]; exists {
		fontHandler := http.FileServer(http.FS(fontsFS))
		e.GET("/node_modules/@formio/js/dist/fonts/*", echo.WrapHandler(
			http.StripPrefix("/node_modules/@formio/js/dist/fonts/", fontHandler),
		))
	}
}

// createEmbeddedFileHandler creates a handler for embedded files
func (s *EmbeddedAssetServer) createEmbeddedFileHandler(filesystem fs.FS, filename string) echo.HandlerFunc {
	return func(c echo.Context) error {
		data, err := fs.ReadFile(filesystem, filename)
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}

		mimeType := s.detectMimeType(filename)
		c.Response().Header().Set("Content-Type", mimeType)

		return c.Blob(http.StatusOK, mimeType, data)
	}
}

// createSecurityHeadersMiddleware creates middleware for setting security headers
func (s *DevelopmentAssetServer) createSecurityHeadersMiddleware() echo.MiddlewareFunc {
	return s.setupStaticFileHeaders()
}

// createSecurityHeadersMiddleware creates middleware for setting security headers
func (s *EmbeddedAssetServer) createSecurityHeadersMiddleware() echo.MiddlewareFunc {
	return s.setupStaticFileHeaders()
}

// setupStaticFileHeaders adds security and caching headers for static files
func (s *DevelopmentAssetServer) setupStaticFileHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set security headers
			for header, value := range s.serverConfig.SecurityHeaders {
				c.Response().Header().Set(header, value)
			}

			// Set cache headers
			maxAge := int(s.serverConfig.MaxAge.Seconds())
			c.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))

			return next(c)
		}
	}
}

// setupStaticFileHeaders adds security and caching headers for static files
func (s *EmbeddedAssetServer) setupStaticFileHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set security headers
			for header, value := range s.serverConfig.SecurityHeaders {
				c.Response().Header().Set(header, value)
			}

			// Set cache headers with longer cache for production
			maxAge := int(s.serverConfig.MaxAge.Seconds())
			c.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", maxAge))

			return next(c)
		}
	}
}

// detectMimeType detects MIME type for a file
func (s *DevelopmentAssetServer) detectMimeType(filename string) string {
	return detectMimeType(filename, s.serverConfig.CustomMimeTypes)
}

// detectMimeType detects MIME type for a file
func (s *EmbeddedAssetServer) detectMimeType(filename string) string {
	return detectMimeType(filename, s.serverConfig.CustomMimeTypes)
}

// detectMimeType detects MIME type for a file with custom types support
func detectMimeType(filename string, customTypes map[string]string) string {
	ext := filepath.Ext(filename)

	// Check custom MIME types first
	if mimeType, exists := customTypes[ext]; exists {
		return mimeType
	}

	// Use standard library
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		return mimeType
	}

	// Fallback for common web files
	switch ext {
	case ".ico":
		return "image/x-icon"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

// AssetServerFactory creates the appropriate asset server based on environment
type AssetServerFactory struct {
	config *config.Config
	logger logging.Logger
}

// NewAssetServerFactory creates a new asset server factory
func NewAssetServerFactory(cfg *config.Config, logger logging.Logger) *AssetServerFactory {
	return &AssetServerFactory{
		config: cfg,
		logger: logger,
	}
}

// CreateServer creates the appropriate asset server for the current environment
func (f *AssetServerFactory) CreateServer(distFS embed.FS) AssetServer {
	if f.config.App.IsDevelopment() {
		f.logger.Info("creating development asset server")

		return NewDevelopmentAssetServer(f.config, f.logger)
	}

	f.logger.Info("creating embedded asset server")

	return NewEmbeddedAssetServer(f.logger, distFS)
}
