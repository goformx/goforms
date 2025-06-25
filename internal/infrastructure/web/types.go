// Package web provides utilities for handling web assets in the application.
// It supports both development mode (using Vite dev server) and production mode
// (using built assets from the Vite manifest).
//
//go:generate mockgen -typed -source=types.go -destination=../../../test/mocks/web/mock_web.go -package=web
package web

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"
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
	// MaxPathLength represents the maximum allowed path length
	MaxPathLength = 100
)

// Asset-related errors
var (
	ErrAssetNotFound    = errors.New("asset not found")
	ErrInvalidManifest  = errors.New("invalid manifest")
	ErrInvalidPath      = errors.New("invalid asset path")
	ErrManifestNotFound = errors.New("manifest not found")
)

// ManifestEntry represents an entry in the Vite manifest file
type ManifestEntry struct {
	File    string   `json:"file"`
	Name    string   `json:"name"`
	Src     string   `json:"src"`
	IsEntry bool     `json:"is_entry"`
	CSS     []string `json:"css"`
}

// Manifest represents the Vite manifest file
type Manifest map[string]ManifestEntry

// AssetResolver interface separates resolution logic from management
type AssetResolver interface {
	ResolveAssetPath(ctx context.Context, path string) (string, error)
}

// AssetServer defines the interface for serving assets
type AssetServer interface {
	// RegisterRoutes registers the necessary routes for serving assets
	RegisterRoutes(e *echo.Echo) error
}

// AssetManagerInterface defines the contract for asset management
type AssetManagerInterface interface {
	// AssetPath returns the resolved asset path for the given input path
	AssetPath(path string) string
	// ResolveAssetPath resolves asset paths with context and proper error handling
	ResolveAssetPath(ctx context.Context, path string) (string, error)
	// GetAssetType returns the type of asset based on its path
	GetAssetType(path string) AssetType
	// ClearCache clears the asset path cache
	ClearCache()
}

// Module encapsulates the asset manager and server to eliminate global state
type Module struct {
	AssetManager AssetManagerInterface
	AssetServer  AssetServer
}
