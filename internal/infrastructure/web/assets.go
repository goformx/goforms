package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
)

// ManifestEntry represents a single entry in the Vite manifest file
type ManifestEntry struct {
	File    string   `json:"file"`
	Src     string   `json:"src,omitempty"`
	CSS     []string `json:"css,omitempty"`
	Imports []string `json:"imports,omitempty"`
}

// ViteManifest represents the entire Vite manifest file
type ViteManifest map[string]ManifestEntry

var (
	// Manifest contains the webpack manifest data for asset versioning
	Manifest ViteManifest
	cfg      *config.Config
)

func init() {
	var err error
	cfg, err = config.New()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Only load manifest in production mode
	if !cfg.App.IsDevelopment() {
		// Load the Vite manifest file
		manifestPath := filepath.Join("dist", ".vite", "manifest.json")
		log.Printf("Attempting to load manifest from: %s", manifestPath)

		manifestData, readErr := os.ReadFile(manifestPath)
		if readErr != nil {
			log.Printf("Warning: Could not read manifest file at %s: %v", manifestPath, readErr)
			return
		}

		if unmarshalErr := json.Unmarshal(manifestData, &Manifest); unmarshalErr != nil {
			log.Printf("Warning: Could not parse manifest file: %v", unmarshalErr)
			return
		}

		log.Printf("Successfully loaded manifest with %d entries", len(Manifest))
		for key, entry := range Manifest {
			log.Printf("Manifest entry: %s -> %s", key, entry.File)
		}
	}
}

// GetAssetPath returns the path to an asset, handling development and production environments
func GetAssetPath(asset string) string {
	if cfg.App.Env == "development" {
		hostPort := net.JoinHostPort(cfg.App.ViteDevHost, cfg.App.ViteDevPort)
		return fmt.Sprintf("http://%s/%s", hostPort, asset)
	}
	return fmt.Sprintf("/assets/%s", asset)
}

// GetManifestPath returns the path to an asset from the manifest
func GetManifestPath(asset string) string {
	if cfg.App.Env == "development" {
		hostPort := net.JoinHostPort(cfg.App.ViteDevHost, cfg.App.ViteDevPort)
		return fmt.Sprintf("http://%s/%s", hostPort, asset)
	}
	if path, ok := Manifest[asset]; ok {
		return fmt.Sprintf("/assets/%s", path.File)
	}
	return fmt.Sprintf("/assets/%s", asset)
}
