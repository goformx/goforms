package web

import (
	"encoding/json"
	"os"
	"path/filepath"
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

var manifest Manifest

func init() {
	// Load the Vite manifest file
	manifestPath := filepath.Join("static", "dist", ".vite", "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		// In development, files are served directly
		return
	}

	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		// Handle error gracefully in production
		return
	}
}

// GetAssetPath returns the hashed path for a given source file
// If the manifest can't be read or the file isn't found, it returns a default path
func GetAssetPath(src string) string {
	if manifest == nil {
		// Return a default path if manifest can't be read
		switch src {
		case "src/css/main.css":
			return "/static/dist/css/styles.1OrqC9gA.css"
		case "src/js/main.ts":
			return "/static/dist/js/app.ChKofpG5.js"
		case "src/js/login.ts":
			return "/static/dist/js/login.CBkkGgj2.js"
		case "src/js/signup.ts":
			return "/static/dist/js/signup.BFv0nUSD.js"
		case "src/js/validation.ts":
			return "/static/dist/js/validation.Cyw-oOUb.js"
		default:
			return ""
		}
	}

	if entry, ok := manifest[src]; ok {
		return filepath.Join("/static/dist", entry.File)
	}

	return ""
}
