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

// GetAssetPath returns the hashed filename for a given asset
func GetAssetPath(assetName string) string {
	manifestPath := filepath.Join("static", "dist", "manifest.json")
	
	// Read manifest file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return assetName // Fallback to original name if manifest not found
	}

	var manifest AssetManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return assetName // Fallback to original name if manifest is invalid
	}

	// Return the hashed filename based on the asset name
	switch assetName {
	case "main.js":
		return filepath.Join("dist", manifest.Main.File)
	case "validation.js":
		return filepath.Join("dist", manifest.Validation.File)
	default:
		return assetName
	}
} 