package web

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/goformx/goforms/internal/infrastructure/config"
)

type ManifestEntry struct {
	File    string   `json:"file"`
	Src     string   `json:"src"`
	IsEntry bool     `json:"isEntry"`
	CSS     []string `json:"css"`
}

type Manifest map[string]ManifestEntry

var manifest Manifest
var appConfig *config.Config
var manifestLoaded bool

// loadManifest attempts to load the Vite manifest file
func loadManifest() error {
	if manifestLoaded {
		return nil
	}

	manifestPath := filepath.Join("dist", ".vite", "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Manifest doesn't exist, initialize empty manifest
			manifest = make(Manifest)
			manifestLoaded = true
			return nil
		}
		return fmt.Errorf("failed to read manifest file: %w", err)
	}

	if err = json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest file: %w", err)
	}

	manifestLoaded = true
	return nil
}

// SetConfig sets the application configuration
func SetConfig(cfg *config.Config) {
	appConfig = cfg
}

// GetAssetPath returns the correct path for an asset based on the environment
func GetAssetPath(path string) string {
	// In development mode, use Vite's dev server
	if appConfig != nil && appConfig.App.IsDevelopment() {
		hostPort := net.JoinHostPort(appConfig.App.ViteDevHost, appConfig.App.ViteDevPort)
		// For source files, use the Vite dev server
		if strings.HasPrefix(path, "src/") {
			return fmt.Sprintf("http://%s/%s", hostPort, path)
		}
		// For built assets, use the Vite dev server with the original path
		return fmt.Sprintf("http://%s/assets/%s", hostPort, path)
	}

	// In production mode, try to use the manifest
	if err := loadManifest(); err == nil {
		if entry, ok := manifest[path]; ok {
			return "/" + entry.File
		}
	}

	// Fallback to direct asset path if manifest loading failed or entry not found
	return "/assets/" + path
}
