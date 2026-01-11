// Package web provides utilities for handling web assets in the application.
// It supports both development mode (using Vite dev server) and production mode
// (using built assets from the Vite manifest).
//
//go:generate mockgen -typed -source=types.go -destination=../../../test/mocks/web/mock_web.go -package=web
package web

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

// AssetType represents the type of asset
type AssetType string

const (
	// AssetTypeJS represents JavaScript files
	AssetTypeJS AssetType = "js"
	// AssetTypeTS represents TypeScript files
	AssetTypeTS AssetType = "ts"
	// AssetTypeCSS represents CSS files
	AssetTypeCSS AssetType = "css"
	// AssetTypeImage represents image files
	AssetTypeImage AssetType = "image"
	// AssetTypeFont represents font files
	AssetTypeFont AssetType = "font"
	// AssetTypeJSON represents JSON files
	AssetTypeJSON AssetType = "json"
	// AssetTypeUnknown represents unknown file types
	AssetTypeUnknown AssetType = "unknown"

	// MaxPathLength represents the maximum allowed path length
	MaxPathLength = 255
	// MinPathLength represents the minimum allowed path length
	MinPathLength = 1
)

// Asset-related errors
var (
	ErrAssetNotFound    = errors.New("asset not found")
	ErrInvalidManifest  = errors.New("invalid manifest")
	ErrInvalidPath      = errors.New("invalid asset path")
	ErrManifestNotFound = errors.New("manifest not found")
	ErrPathTooLong      = errors.New("asset path too long")
	ErrPathTooShort     = errors.New("asset path too short")
	ErrEmptyManifest    = errors.New("manifest is empty")
	ErrResolverNotFound = errors.New("asset resolver not found")
)

// ManifestEntry represents an entry in the Vite manifest file
type ManifestEntry struct {
	File           string   `json:"file"`
	Name           string   `json:"name,omitempty"`
	Src            string   `json:"src,omitempty"`
	IsEntry        bool     `json:"is_entry"`
	CSS            []string `json:"css,omitempty"`
	Assets         []string `json:"assets,omitempty"`
	Imports        []string `json:"imports,omitempty"`
	DynamicImports []string `json:"dynamic_imports,omitempty"`
}

// Manifest represents the Vite manifest file
type Manifest map[string]ManifestEntry

// IsEmpty returns true if the manifest has no entries
func (m Manifest) IsEmpty() bool {
	return len(m) == 0
}

// GetEntry safely retrieves a manifest entry
func (m Manifest) GetEntry(path string) (ManifestEntry, bool) {
	entry, exists := m[path]

	return entry, exists
}

// GetEntryPaths returns all entry point paths from the manifest
func (m Manifest) GetEntryPaths() []string {
	var entries []string

	for path := range m {
		if m[path].IsEntry {
			entries = append(entries, path)
		}
	}

	return entries
}

// AssetResolver defines the interface for resolving asset paths
type AssetResolver interface {
	// ResolveAssetPath resolves an asset path with context and proper error handling
	ResolveAssetPath(ctx context.Context, path string) (string, error)
}

// AssetServer defines the interface for serving assets
type AssetServer interface {
	// RegisterRoutes registers the necessary routes for serving assets
	RegisterRoutes(e *echo.Echo) error
	// Start starts the asset server if needed (e.g., for development mode)
	Start(ctx context.Context) error
	// Stop stops the asset server gracefully
	Stop(ctx context.Context) error
}

// AssetManagerInterface defines the contract for asset management
type AssetManagerInterface interface {
	// AssetPath returns the resolved asset path for the given input path
	// This is a convenience method that uses a background context
	AssetPath(path string) string

	// ResolveAssetPath resolves asset paths with context and proper error handling
	ResolveAssetPath(ctx context.Context, path string) (string, error)

	// GetAssetType returns the type of asset based on its path
	GetAssetType(path string) AssetType

	// ClearCache clears the asset path cache (if applicable)
	ClearCache()

	// ValidatePath validates an asset path
	ValidatePath(path string) error

	// GetBaseURL returns the base URL for assets (useful for CSP headers)
	GetBaseURL() string
}

// AssetPathValidator defines validation rules for asset paths
type AssetPathValidator interface {
	// ValidatePath validates an asset path according to defined rules
	ValidatePath(path string) error

	// IsValidExtension checks if the file extension is allowed
	IsValidExtension(ext string) bool
}

// AssetCache defines the interface for caching resolved asset paths
type AssetCache interface {
	// Get retrieves a cached asset path
	Get(key string) (string, bool)

	// Set stores an asset path in the cache
	Set(key, value string)

	// Clear clears all cached entries
	Clear()

	// Size returns the number of cached entries
	Size() int
}

// Module encapsulates the asset manager and server to eliminate global state
type Module struct {
	AssetManager  AssetManagerInterface
	AssetServer   AssetServer
	AssetResolver AssetResolver
}

// AssetInfo contains metadata about an asset
type AssetInfo struct {
	Path         string
	ResolvedPath string
	Type         AssetType
	Size         int64
	Exists       bool
	IsEntry      bool
	Dependencies []string
}

// Environment represents the deployment environment
type Environment string

const (
	// EnvironmentDevelopment represents development mode
	EnvironmentDevelopment Environment = "development"
	// EnvironmentProduction represents production mode
	EnvironmentProduction Environment = "production"
	// EnvironmentTest represents test mode
	EnvironmentTest Environment = "test"
)

// AssetConfig holds configuration for asset management
type AssetConfig struct {
	Environment       Environment
	BaseURL           string
	ManifestPath      string
	StaticPath        string
	CacheEnabled      bool
	MaxCacheSize      int
	AllowedExtensions []string
	ViteDevServerURL  string
}

// GetAssetTypeFromPath determines the asset type from a file path
func GetAssetTypeFromPath(path string) AssetType {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".js", ".mjs", ".cjs":
		return AssetTypeJS
	case ".ts", ".tsx":
		return AssetTypeTS
	case ".css", ".scss", ".sass", ".less":
		return AssetTypeCSS
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".ico":
		return AssetTypeImage
	case ".woff", ".woff2", ".ttf", ".otf", ".eot":
		return AssetTypeFont
	case ".json":
		return AssetTypeJSON
	default:
		return AssetTypeUnknown
	}
}

// ValidateAssetPath validates an asset path according to standard rules
func ValidateAssetPath(path string) error {
	if len(path) < MinPathLength {
		return ErrPathTooShort
	}

	if len(path) > MaxPathLength {
		return ErrPathTooLong
	}

	if strings.TrimSpace(path) == "" {
		return ErrInvalidPath
	}

	// Check for potentially dangerous path traversal
	if strings.Contains(path, "..") {
		return ErrInvalidPath
	}

	return nil
}

// NormalizePath normalizes an asset path for consistent handling
func NormalizePath(path string) string {
	// Clean the path and ensure forward slashes
	cleaned := filepath.ToSlash(filepath.Clean(path))

	// Remove leading slash for consistency
	cleaned = strings.TrimPrefix(cleaned, "/")

	return cleaned
}
