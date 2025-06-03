package web

import (
	"path"
)

// GetAssetPath returns the full path to an asset file
func GetAssetPath(assetPath string) string {
	return path.Join("/assets", assetPath)
}
