package web

import (
	"encoding/json"
	"log"
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

// Manifest represents the entire Vite manifest file
type Manifest map[string]ManifestEntry

var (
	manifest Manifest
	cfg      *config.Config
)

func init() {
	var err error
	cfg, err = config.New()
	if err != nil {
		log.Printf("Warning: Could not load config: %v", err)
		return
	}

	// Only load manifest in production mode
	if !cfg.App.IsDevelopment() {
		// Load the Vite manifest file
		manifestPath := filepath.Join("static", "dist", ".vite", "manifest.json")
		log.Printf("Attempting to load manifest from: %s", manifestPath)

		manifestData, readErr := os.ReadFile(manifestPath)
		if readErr != nil {
			log.Printf("Warning: Could not read manifest file at %s: %v", manifestPath, readErr)
			return
		}

		if unmarshalErr := json.Unmarshal(manifestData, &manifest); unmarshalErr != nil {
			log.Printf("Warning: Could not parse manifest file: %v", unmarshalErr)
			return
		}

		log.Printf("Successfully loaded manifest with %d entries", len(manifest))
		for key, entry := range manifest {
			log.Printf("Manifest entry: %s -> %s", key, entry.File)
		}
	}
}

// GetAssetPath returns the path for a given source file
// In development mode, it returns the Vite dev server URL
// In production mode, it returns the hashed path from the manifest
func GetAssetPath(src string) string {
	log.Printf("Getting asset path for: %s", src)

	if cfg != nil && cfg.App.IsDevelopment() {
		// In development, return the Vite dev server URL
		return "http://localhost:3000/" + src
	}

	if entry, ok := manifest[src]; ok {
		path := filepath.Join("static", "dist", entry.File)
		log.Printf("Found manifest entry: %s -> %s", src, path)
		return path
	}

	log.Printf("No manifest entry found for: %s", src)
	return ""
}
