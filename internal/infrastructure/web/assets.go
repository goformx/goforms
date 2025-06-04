package web

import (
	"encoding/json"
	"fmt"
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

func init() {
	// Read the manifest file
	manifestPath := filepath.Join("dist", ".vite", "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		panic(err)
	}

	// Parse the manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		panic(err)
	}
}

// SetConfig sets the application configuration
func SetConfig(cfg *config.Config) {
	appConfig = cfg
}

// GetAssetPath returns the correct path for an asset based on the environment
func GetAssetPath(path string) string {
	// In development mode, use Vite's dev server
	if appConfig != nil && appConfig.App.IsDevelopment() {
		// For source files, use the Vite dev server
		if strings.HasPrefix(path, "src/") {
			return fmt.Sprintf("http://%s:%s/%s", appConfig.App.ViteDevHost, appConfig.App.ViteDevPort, path)
		}
		// For built assets, use the Vite dev server with the original path
		return fmt.Sprintf("http://%s:%s/assets/%s", appConfig.App.ViteDevHost, appConfig.App.ViteDevPort, path)
	}

	// In production mode, use the manifest
	if entry, ok := manifest[path]; ok {
		return "/" + entry.File
	}
	return "/assets/" + path
}
