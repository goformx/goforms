package web

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AssetManifest represents the structure of Vite's manifest.json
type AssetManifest struct {
	Main struct {
		File    string   `json:"file"`
		CSS     []string `json:"css,omitempty"`
		Imports []string `json:"imports,omitempty"`
	} `json:"main"`
	Validation struct {
		File    string   `json:"file"`
		CSS     []string `json:"css,omitempty"`
		Imports []string `json:"imports,omitempty"`
	} `json:"validation"`
}

type ViteManifest map[string]struct {
	File    string   `json:"file"`
	Src     string   `json:"src,omitempty"`
	CSS     []string `json:"css,omitempty"`
	Imports []string `json:"imports,omitempty"`
}

var manifest ViteManifest

func init() {
	// Load the Vite manifest file
	manifestPath := filepath.Join("static", "dist", ".vite", "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		// In development, files are served directly
		return
	}

	unmarshalErr := json.Unmarshal(manifestData, &manifest)
	if unmarshalErr != nil {
		// Handle error gracefully in production
		return
	}
}

// GetAssetPath returns the hashed filename for a given asset
func GetAssetPath(assetName string) string {
	return GetViteAssetPath(assetName)
}

// GetViteAssetPath returns the correct path for a Vite asset
func GetViteAssetPath(path string) string {
	if manifest == nil {
		// In development, return the path as is
		return path
	}

	// In production, get the hashed filename from the manifest
	if entry, ok := manifest[path]; ok {
		return entry.File
	}

	return path
}
