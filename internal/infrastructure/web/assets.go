// Package web provides utilities for handling web assets in the application.
// It supports both development mode (using Vite dev server) and production mode
// (using built assets from the Vite manifest).
package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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
	logger         logging.Logger
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

	manifestPath := filepath.Join("dist", ".vite", "manifest.json")
	if m.logger != nil {
		m.logger.Debug("attempting to load manifest", "path", manifestPath)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			if m.logger != nil {
				m.logger.Debug("manifest file not found, trying alternative location", "path", manifestPath)
			}
			// Try alternative manifest location
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
		} else {
			return fmt.Errorf("failed to read manifest file: %w", err)
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

// GetAssetPath returns the correct path for an asset based on the environment
func (m *AssetManager) GetAssetPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
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

	// In development mode, use Vite's dev server
	if m.config != nil && m.config.App.IsDevelopment() {
		hostPort := net.JoinHostPort(m.config.App.ViteDevHost, m.config.App.ViteDevPort)
		var assetPath string

		// For source files, use the Vite dev server
		if strings.HasPrefix(path, "src/") {
			assetPath = fmt.Sprintf("%s://%s/%s", m.config.App.Scheme, hostPort, path)
		} else {
			// In development, we need to use the entry points directly
			// Remove any file extension for the entry point
			entryPoint := strings.TrimSuffix(path, filepath.Ext(path))

			// Handle CSS files differently
			if strings.HasSuffix(path, ".css") {
				assetPath = fmt.Sprintf("%s://%s/src/css/%s.css", m.config.App.Scheme, hostPort, entryPoint)
			} else {
				assetPath = fmt.Sprintf("%s://%s/src/js/%s.ts", m.config.App.Scheme, hostPort, entryPoint)
			}
		}

		if m.logger != nil {
			m.logger.Debug("development mode asset path", "path", path, "resolved", assetPath)
		}

		// Cache the result
		m.mu.Lock()
		m.pathCache[path] = assetPath
		m.mu.Unlock()

		return assetPath, nil
	}

	// In production mode, try to use the manifest
	if err := m.loadManifest(); err != nil {
		if m.logger != nil {
			m.logger.Error("failed to load manifest", "error", err)
		}
	} else {
		if m.logger != nil {
			m.logger.Debug("manifest loaded successfully")
		}

		// Try to find the entry in the manifest
		if entry, ok := m.manifest[path]; ok {
			if m.logger != nil {
				m.logger.Debug("found direct match in manifest", "path", path)
			}

			// For entry points, use the file directly
			if entry.IsEntry {
				assetPath := "/" + entry.File
				if m.logger != nil {
					m.logger.Debug("entry point found", "path", path, "resolved", assetPath)
				}

				// Cache the result
				m.mu.Lock()
				m.pathCache[path] = assetPath
				m.mu.Unlock()

				return assetPath, nil
			}

			// For CSS files referenced by entries
			if len(entry.CSS) > 0 {
				assetPath := "/" + entry.CSS[0]
				if m.logger != nil {
					m.logger.Debug("CSS file found", "path", path, "resolved", assetPath)
				}

				// Cache the result
				m.mu.Lock()
				m.pathCache[path] = assetPath
				m.mu.Unlock()

				return assetPath, nil
			}

			// For other files
			assetPath := "/" + entry.File
			if m.logger != nil {
				m.logger.Debug("other file found", "path", path, "resolved", assetPath)
			}

			// Cache the result
			m.mu.Lock()
			m.pathCache[path] = assetPath
			m.mu.Unlock()

			return assetPath, nil
		}

		// Try to find the entry by its base name (without extension)
		baseName := strings.TrimSuffix(path, filepath.Ext(path))
		if m.logger != nil {
			m.logger.Debug("trying to match by base name", "base_name", baseName)
		}

		for key, entry := range m.manifest {
			if strings.TrimSuffix(key, filepath.Ext(key)) == baseName {
				if m.logger != nil {
					m.logger.Debug("found match by base name", "key", key, "file", entry.File)
				}

				if entry.IsEntry {
					assetPath := "/" + entry.File
					m.mu.Lock()
					m.pathCache[path] = assetPath
					m.mu.Unlock()
					return assetPath, nil
				}
				if len(entry.CSS) > 0 {
					assetPath := "/" + entry.CSS[0]
					m.mu.Lock()
					m.pathCache[path] = assetPath
					m.mu.Unlock()
					return assetPath, nil
				}
				assetPath := "/" + entry.File
				m.mu.Lock()
				m.pathCache[path] = assetPath
				m.mu.Unlock()
				return assetPath, nil
			}
		}

		// For CSS files, look in the CSS arrays of all entries
		if strings.HasSuffix(path, ".css") {
			if m.logger != nil {
				m.logger.Debug("searching for CSS file", "path", path)
			}
			for _, entry := range m.manifest {
				if entry.CSS != nil {
					for _, cssFile := range entry.CSS {
						if strings.HasSuffix(cssFile, filepath.Base(path)) {
							assetPath := "/" + cssFile
							if m.logger != nil {
								m.logger.Debug("found CSS file in manifest", "path", path, "resolved", assetPath)
							}
							m.mu.Lock()
							m.pathCache[path] = assetPath
							m.mu.Unlock()
							return assetPath, nil
						}
					}
				}
			}
		}
	}

	// Fallback to direct asset path if manifest loading failed or entry not found
	ext := filepath.Ext(path)
	var assetPath string
	switch ext {
	case ".js":
		assetPath = "/assets/js/" + filepath.Base(path)
	case ".css":
		assetPath = "/assets/css/" + filepath.Base(path)
	default:
		assetPath = "/assets/" + path
	}

	if m.logger != nil {
		m.logger.Debug("using fallback path", "path", path, "resolved", assetPath)
	}

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
