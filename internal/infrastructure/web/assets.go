// Package web provides utilities for handling web assets in the application.
// It supports both development mode (using Vite dev server) and production mode
// (using built assets from the Vite manifest).
package web

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goformx/goforms/internal/infrastructure/config"
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
)

// ManifestEntry represents an entry in the Vite manifest file
type ManifestEntry struct {
	File    string   `json:"file"`
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
}

var (
	// DefaultManager is the default asset manager instance
	DefaultManager = NewAssetManager()
)

// NewAssetManager creates a new asset manager instance
func NewAssetManager() *AssetManager {
	return &AssetManager{
		manifest:  make(Manifest),
		pathCache: make(map[string]string),
	}
}

// SetConfig sets the application configuration
func (m *AssetManager) SetConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
	return nil
}

// loadManifest attempts to load the Vite manifest file
func (m *AssetManager) loadManifest() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.manifestLoaded {
		return nil
	}

	manifestPath := filepath.Join("dist", ".vite", "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Manifest doesn't exist, initialize empty manifest
			m.manifest = make(Manifest)
			m.manifestLoaded = true
			return nil
		}
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	if err = json.Unmarshal(data, &m.manifest); err != nil {
		return fmt.Errorf("failed to parse manifest file: %w", err)
	}

	m.manifestLoaded = true
	return nil
}

// GetAssetPath returns the correct path for an asset based on the environment
func (m *AssetManager) GetAssetPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check cache first
	m.mu.RLock()
	if cached, ok := m.pathCache[path]; ok {
		m.mu.RUnlock()
		return cached, nil
	}
	m.mu.RUnlock()

	// In development mode, use Vite's dev server
	if m.config != nil && m.config.App.IsDevelopment() {
		hostPort := net.JoinHostPort(m.config.App.ViteDevHost, m.config.App.ViteDevPort)
		var assetPath string

		// For source files, use the Vite dev server
		if strings.HasPrefix(path, "src/") {
			assetPath = fmt.Sprintf("http://%s/%s", hostPort, path)
		} else {
			// For built assets, use the Vite dev server with the original path
			assetPath = fmt.Sprintf("http://%s/assets/%s", hostPort, path)
		}

		// Cache the result
		m.mu.Lock()
		m.pathCache[path] = assetPath
		m.mu.Unlock()

		return assetPath, nil
	}

	// In production mode, try to use the manifest
	if err := m.loadManifest(); err == nil {
		if entry, ok := m.manifest[path]; ok {
			assetPath := "/" + entry.File

			// Cache the result
			m.mu.Lock()
			m.pathCache[path] = assetPath
			m.mu.Unlock()

			return assetPath, nil
		}
	}

	// Fallback to direct asset path if manifest loading failed or entry not found
	assetPath := "/assets/" + path

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

// GetAssetPath returns the asset path using the default asset manager
func GetAssetPath(path string) (string, error) {
	return DefaultManager.GetAssetPath(path)
}

// GetAssetType returns the asset type using the default asset manager
func GetAssetType(path string) AssetType {
	return DefaultManager.GetAssetType(path)
}
