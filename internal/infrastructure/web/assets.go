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
	distFS         embed.FS // Add embedded filesystem for production
}

// AssetServer defines the interface for serving assets
type AssetServer interface {
	// RegisterRoutes registers the necessary routes for serving assets
	RegisterRoutes(e *echo.Echo) error
}

// WebModule encapsulates the asset manager and server to eliminate global state
type WebModule struct {
	AssetManager *AssetManager
	AssetServer  AssetServer
}

// NewWebModule creates a new web module with proper dependency injection
func NewWebModule(cfg *config.Config, logger logging.Logger, distFS embed.FS) *WebModule {
	manager := NewAssetManager(cfg, logger, distFS)

	var server AssetServer
	if cfg.App.IsDevelopment() {
		server = NewViteAssetServer(cfg, logger)
	} else {
		// In production, always use embedded filesystem
		server = NewEmbeddedAssetServer(logger, distFS)
	}

	return &WebModule{
		AssetManager: manager,
		AssetServer:  server,
	}
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

// NewAssetManager creates a new asset manager instance
func NewAssetManager(cfg *config.Config, logger logging.Logger, distFS ...embed.FS) *AssetManager {
	manager := &AssetManager{
		manifest:  make(Manifest),
		pathCache: make(map[string]string),
		config:    cfg,
		logger:    logger,
	}

	// Set embedded filesystem if provided
	if len(distFS) > 0 {
		manager.distFS = distFS[0]
	}

	// Try to load manifest immediately
	if err := manager.loadManifest(); err != nil {
		manager.logger.Error("failed to load manifest during initialization", "error", err)
	} else {
		manager.logger.Info("asset manager initialized",
			"manifest_loaded", manager.manifestLoaded,
			"manifest_entries", len(manager.manifest),
		)
	}

	return manager
}

// loadManifest attempts to load the Vite manifest file
func (m *AssetManager) loadManifest() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.manifestLoaded {
		return nil
	}

	var data []byte
	var err error

	// Try to load from embedded filesystem
	manifestPath := "dist/.vite/manifest.json"
	m.logger.Debug("attempting to load manifest from embedded filesystem", "path", manifestPath)

	data, err = fs.ReadFile(m.distFS, manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest from embedded filesystem: %w", err)
	}

	// Parse the manifest
	if unmarshalErr := json.Unmarshal(data, &m.manifest); unmarshalErr != nil {
		return fmt.Errorf("failed to parse manifest: %w", unmarshalErr)
	}

	m.manifestLoaded = true

	m.logger.Info("manifest loaded successfully",
		"entries", len(m.manifest),
	)

	return nil
}

// getDevelopmentAssetPath returns the asset path for development mode
func (m *AssetManager) getDevelopmentAssetPath(path string) string {
	hostPort := net.JoinHostPort(m.config.App.ViteDevHost, m.config.App.ViteDevPort)

	// Debug logging
	m.logger.Debug("getting development asset path",
		"path", path,
		"host_port", hostPort,
	)

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

// getManifestKeys returns a slice of all keys in the manifest for debugging
func getManifestKeys(manifest Manifest) []string {
	keys := make([]string, 0, len(manifest))
	for key := range manifest {
		keys = append(keys, key)
	}
	return keys
}

// findAssetInManifest searches for an asset in the loaded manifest
func (m *AssetManager) findAssetInManifest(path string) (string, bool) {
	m.logger.Debug("searching manifest for asset",
		"path", path,
		"manifest_entries", len(m.manifest),
	)

	// Only allow exact match
	if entry, found := m.manifest[path]; found {
		m.logger.Debug("exact match found in manifest",
			"path", path,
			"manifest_file", entry.File,
		)
		return entry.File, true
	}

	m.logger.Debug("asset not found in manifest",
		"path", path,
		"available_keys", getManifestKeys(m.manifest),
	)

	return "", false
}

// getProductionAssetPath returns the asset path for production mode
func (m *AssetManager) getProductionAssetPath(path string) string {
	m.logger.Debug("getting production asset path",
		"path", path,
		"manifest_loaded", m.manifestLoaded,
		"manifest_entries", len(m.manifest),
	)

	// Try to load manifest if not loaded
	if !m.manifestLoaded {
		if err := m.loadManifest(); err != nil {
			m.logger.Error("failed to load manifest", "error", err)
		}
	}

	// Try to find the asset in the manifest
	if assetPath, found := m.findAssetInManifest(path); found {
		// Ensure the path starts with a slash for proper URL construction
		if !strings.HasPrefix(assetPath, "/") {
			assetPath = "/" + assetPath
		}
		m.logger.Debug("asset found in manifest",
			"input_path", path,
			"manifest_path", assetPath,
		)
		return assetPath
	}

	m.logger.Debug("asset not found in manifest",
		"input_path", path,
	)
	return ""
}

// AssetPath returns the resolved asset path for the given input path
func (m *AssetManager) AssetPath(path string) string {
	m.logger.Debug("asset manager resolving path",
		"input_path", path,
		"manifest_loaded", m.manifestLoaded,
		"manifest_entries", len(m.manifest),
	)

	// Check cache first
	m.mu.RLock()
	if cachedPath, found := m.pathCache[path]; found {
		m.mu.RUnlock()
		m.logger.Debug("asset path found in cache", "input_path", path, "cached_path", cachedPath)
		return cachedPath
	}
	m.mu.RUnlock()

	// Resolve the path
	var resolvedPath string
	if m.config.App.Env == "production" {
		resolvedPath = m.getProductionAssetPath(path)
	} else {
		resolvedPath = m.getDevelopmentAssetPath(path)
	}

	// Cache the result
	m.mu.Lock()
	m.pathCache[path] = resolvedPath
	m.mu.Unlock()

	m.logger.Debug("asset path resolved",
		"input_path", path,
		"resolved_path", resolvedPath,
		"environment", m.config.App.Env,
	)

	return resolvedPath
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
