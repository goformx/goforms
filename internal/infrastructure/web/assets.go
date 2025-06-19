// Package web provides utilities for handling web assets in the application.
// It supports both development mode (using Vite dev server) and production mode
// (using built assets from the Vite manifest).
package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// AssetType represents the type of asset
type AssetType string

const (
	// AssetTypeJS represents JavaScript files
	AssetTypeJS AssetType = "js"
	// AssetTypeCSS represents CSS files
	AssetTypeCSS AssetType = "css"
	// AssetTypeImage represents image files
	AssetTypeImage AssetType = "image"
	// AssetTypeFont represents font files
	AssetTypeFont AssetType = "font"
	MaxPathLength           = 100
)

// ManifestEntry represents an entry in the Vite manifest file
type ManifestEntry struct {
	File    string   `json:"file"`
	Name    string   `json:"name"`
	Src     string   `json:"src"`
	IsEntry bool     `json:"isEntry"`
	CSS     []string `json:"css"`
}

// Manifest represents the Vite manifest file
type Manifest map[string]ManifestEntry

// AssetManager handles asset path resolution and caching
type AssetManager struct {
	manifest       Manifest
	config         *config.Config
	manifestLoaded bool
	pathCache      map[string]string
	mu             sync.RWMutex
	logger         logging.Logger
}

// AssetServer defines the interface for serving assets
type AssetServer interface {
	// RegisterRoutes registers the necessary routes for serving assets
	RegisterRoutes(e *echo.Echo) error
}

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
		if s.logger != nil {
			// Only log the path portion of the URL, limited to 100 characters
			path := r.URL.Path
			if len(path) > MaxPathLength {
				path = path[:MaxPathLength] + "..."
			}
			s.logger.Error("proxy error",
				"error", err,
				"path", path,
			)
		}
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

// StaticAssetServer implements AssetServer for static files in production
type StaticAssetServer struct {
	logger logging.Logger
}

// NewStaticAssetServer creates a new static asset server
func NewStaticAssetServer(logger logging.Logger) *StaticAssetServer {
	return &StaticAssetServer{
		logger: logger,
	}
}

// RegisterRoutes registers the static file serving routes
func (s *StaticAssetServer) RegisterRoutes(e *echo.Echo) error {
	// Add static file headers middleware
	e.Use(setupStaticFileHeaders)

	// Serve static files from dist directory
	e.Static("/assets", "dist/assets")
	e.Static("/fonts", "dist/fonts")
	e.Static("/css", "dist/css")
	e.Static("/js", "dist/js")

	// Serve individual files
	e.File("/robots.txt", "dist/robots.txt")
	e.File("/favicon.ico", "dist/favicon.ico")

	s.logger.Info("static asset server configured", "base_dir", "dist")
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

	s.logger.Info("embedded asset server configured", "base_dir", "dist")
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

var (
	// DefaultManager is the default asset manager instance
	DefaultManager = NewAssetManager()
)

// NewAssetManager creates a new asset manager instance
func NewAssetManager() *AssetManager {
	manager := &AssetManager{
		manifest:  make(Manifest),
		pathCache: make(map[string]string),
	}

	// Try to load manifest immediately
	if err := manager.loadManifest(); err != nil {
		if manager.logger != nil {
			manager.logger.Error("failed to load manifest during initialization", "error", err)
		}
	} else if manager.logger != nil {
		manager.logger.Info("asset manager initialized",
			"manifest_loaded", manager.manifestLoaded,
			"manifest_entries", len(manager.manifest),
		)
	}

	return manager
}

// SetConfig sets the application configuration
func (m *AssetManager) SetConfig(cfg *config.Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
	return nil
}

// SetLogger sets the logger for the asset manager
func (m *AssetManager) SetLogger(logger logging.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger = logger
}

// loadManifest attempts to load the Vite manifest file
func (m *AssetManager) loadManifest() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.manifestLoaded {
		return nil
	}

	// Try primary manifest location
	manifestPath := filepath.Join("dist", ".vite", "manifest.json")
	if m.logger != nil {
		m.logger.Debug("attempting to load manifest", "path", manifestPath)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	// If primary location fails, try alternative location
	if err != nil {
		if m.logger != nil {
			m.logger.Debug("manifest file not found, trying alternative location", "path", manifestPath)
		}
		manifestPath = filepath.Join("dist", "manifest.json")
		data, err = os.ReadFile(manifestPath)
		if err != nil {
			if m.logger != nil {
				m.logger.Debug("alternative manifest file not found", "path", manifestPath, "error", err)
			}
			// Manifest doesn't exist, initialize empty manifest
			m.manifest = make(Manifest)
			m.manifestLoaded = true
			return nil
		}
	}

	if err = json.Unmarshal(data, &m.manifest); err != nil {
		return fmt.Errorf("failed to parse manifest file: %w", err)
	}

	if m.logger != nil {
		m.logger.Debug("manifest loaded successfully", "entries", len(m.manifest))
	}
	m.manifestLoaded = true
	return nil
}

// getDevelopmentAssetPath returns the asset path for development mode
func (m *AssetManager) getDevelopmentAssetPath(path string) string {
	hostPort := net.JoinHostPort(m.config.App.ViteDevHost, m.config.App.ViteDevPort)

	// Debug logging
	if m.logger != nil {
		m.logger.Debug("getting development asset path",
			"path", path,
			"host_port", hostPort,
		)
	}

	// For source files, use the Vite dev server
	if strings.HasPrefix(path, "src/") {
		return fmt.Sprintf("%s://%s/%s", m.config.App.Scheme, hostPort, path)
	}

	// Handle different asset types
	switch {
	case strings.HasSuffix(path, ".css"):
		// For CSS files, use the Vite dev server's CSS endpoint
		return fmt.Sprintf("%s://%s/src/css/%s", m.config.App.Scheme, hostPort, path)
	case strings.HasSuffix(path, ".ts"), strings.HasSuffix(path, ".js"):
		// For TypeScript/JavaScript files, use the Vite dev server's JS endpoint
		// Remove any .js extension since we're using TypeScript
		baseName := strings.TrimSuffix(path, ".js")
		baseName = strings.TrimSuffix(baseName, ".ts")
		return fmt.Sprintf("%s://%s/src/js/%s.ts", m.config.App.Scheme, hostPort, baseName)
	default:
		// For other assets, try to serve them directly
		return fmt.Sprintf("%s://%s/%s", m.config.App.Scheme, hostPort, path)
	}
}

// findAssetBySourcePath looks for an asset in the manifest by its source path
func (m *AssetManager) findAssetBySourcePath(path string) (string, bool) {
	if entry, ok := m.manifest[path]; ok {
		if m.logger != nil {
			m.logger.Debug("found entry in manifest by source path",
				"path", path,
				"file", entry.File,
			)
		}
		return entry.File, true
	}
	return "", false
}

// findAssetByName looks for an asset in the manifest by its name (without extension)
func (m *AssetManager) findAssetByName(path string) (string, bool) {
	baseName := strings.TrimSuffix(path, filepath.Ext(path))
	for _, entry := range m.manifest {
		if entry.Name == baseName {
			if m.logger != nil {
				m.logger.Debug("found entry in manifest by name",
					"path", path,
					"file", entry.File,
				)
			}
			return entry.File, true
		}
	}
	return "", false
}

// findCSSAsset looks for a CSS asset in the manifest's CSS arrays
func (m *AssetManager) findCSSAsset(path string) (string, bool) {
	if !strings.HasSuffix(path, ".css") {
		return "", false
	}

	for _, entry := range m.manifest {
		for _, cssFile := range entry.CSS {
			if strings.HasSuffix(cssFile, path) {
				if m.logger != nil {
					m.logger.Debug("found CSS file in manifest entry",
						"path", path,
						"file", cssFile,
					)
				}
				return cssFile, true
			}
		}
	}
	return "", false
}

// findAssetInManifest looks for an asset in the manifest by various criteria
func (m *AssetManager) findAssetInManifest(path string) (string, bool) {
	// Try each lookup method in sequence
	if assetPath, found := m.findAssetBySourcePath(path); found {
		return assetPath, true
	}

	if assetPath, found := m.findAssetByName(path); found {
		return assetPath, true
	}

	if assetPath, found := m.findCSSAsset(path); found {
		return assetPath, true
	}

	return "", false
}

// constructAssetPath creates a path for an asset based on its type
func (m *AssetManager) constructAssetPath(path string) string {
	baseName := strings.TrimSuffix(path, filepath.Ext(path))
	ext := filepath.Ext(path)
	if ext == "" {
		ext = ".js" // Default to .js if no extension
	}

	var assetPath string
	switch {
	case strings.HasSuffix(path, ".css"):
		assetPath = fmt.Sprintf("/assets/css/%s.css", baseName)
	case strings.HasSuffix(path, ".js"):
		assetPath = fmt.Sprintf("/assets/js/%s.js", baseName)
	default:
		assetPath = fmt.Sprintf("/assets/%s%s", baseName, ext)
	}

	if m.logger != nil {
		m.logger.Debug("constructed asset path",
			"path", path,
			"resolved", assetPath,
		)
	}

	return assetPath
}

// getProductionAssetPath returns the asset path for production mode
func (m *AssetManager) getProductionAssetPath(path string) string {
	if m.logger != nil {
		m.logger.Debug("getting production asset path",
			"path", path,
			"manifest_loaded", m.manifestLoaded,
			"manifest_entries", len(m.manifest),
		)
	}

	// Try to load manifest if not loaded
	if !m.manifestLoaded {
		if err := m.loadManifest(); err != nil {
			if m.logger != nil {
				m.logger.Error("failed to load manifest", "error", err)
			}
		}
	}

	// Try to find the asset in the manifest
	if assetPath, found := m.findAssetInManifest(path); found {
		return assetPath
	}

	// If not found, construct a path
	return m.constructAssetPath(path)
}

// GetAssetPath returns the correct path for an asset based on the environment
func (m *AssetManager) GetAssetPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	if len(path) > MaxPathLength {
		return "", fmt.Errorf("path too long: %s", path)
	}

	if m.logger != nil {
		m.logger.Debug("resolving asset path", "path", path)
	}

	// Check cache first
	m.mu.RLock()
	if cached, ok := m.pathCache[path]; ok {
		m.mu.RUnlock()
		if m.logger != nil {
			m.logger.Debug("cache hit", "path", path, "cached", cached)
		}
		return cached, nil
	}
	m.mu.RUnlock()

	var assetPath string

	// In development mode, use Vite's dev server
	if m.config != nil && m.config.App.IsDevelopment() {
		assetPath = m.getDevelopmentAssetPath(path)
	} else {
		// In production mode, try to use the manifest
		if manifestErr := m.loadManifest(); manifestErr != nil {
			if m.logger != nil {
				m.logger.Error("failed to load manifest", "error", manifestErr)
			}
		}
		assetPath = m.getProductionAssetPath(path)
	}

	// Cache the result
	m.mu.Lock()
	m.pathCache[path] = assetPath
	m.mu.Unlock()

	return assetPath, nil
}

// GetAssetType returns the type of asset based on its path
func (m *AssetManager) GetAssetType(path string) AssetType {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".mjs":
		return AssetTypeJS
	case ".css", ".scss", ".sass":
		return AssetTypeCSS
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return AssetTypeImage
	case ".woff", ".woff2", ".ttf", ".eot", ".otf":
		return AssetTypeFont
	default:
		return ""
	}
}

// ClearCache clears the asset path cache
func (m *AssetManager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pathCache = make(map[string]string)
}

// SetConfig sets the configuration for the default asset manager
func SetConfig(cfg *config.Config) error {
	return DefaultManager.SetConfig(cfg)
}

// SetLogger sets the logger for the default asset manager
func SetLogger(logger logging.Logger) {
	DefaultManager.SetLogger(logger)
}

// GetAssetPath returns the asset path using the default asset manager
func GetAssetPath(path string) (string, error) {
	return DefaultManager.GetAssetPath(path)
}

// GetAssetType returns the asset type using the default asset manager
func GetAssetType(path string) AssetType {
	return DefaultManager.GetAssetType(path)
}
