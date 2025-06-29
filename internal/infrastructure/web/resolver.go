// Package web provides utilities for handling web assets in the application.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// ProductionAssetResolver handles production asset resolution
type ProductionAssetResolver struct {
	manifest Manifest
	logger   logging.Logger
}

// NewProductionAssetResolver creates a new production asset resolver
func NewProductionAssetResolver(manifest Manifest, logger logging.Logger) *ProductionAssetResolver {
	return &ProductionAssetResolver{
		manifest: manifest,
		logger:   logger,
	}
}

// ResolveAssetPath resolves asset paths for production using the manifest
func (r *ProductionAssetResolver) ResolveAssetPath(_ context.Context, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	entry, found := r.manifest[path]
	if !found {
		return "", fmt.Errorf("%w: %s", ErrAssetNotFound, path)
	}

	assetPath := entry.File
	if !strings.HasPrefix(assetPath, "/") {
		assetPath = "/" + assetPath
	}

	return assetPath, nil
}

// DevelopmentAssetResolver handles development asset resolution
type DevelopmentAssetResolver struct {
	config *config.Config
	logger logging.Logger
}

// NewDevelopmentAssetResolver creates a new development asset resolver
func NewDevelopmentAssetResolver(cfg *config.Config, logger logging.Logger) *DevelopmentAssetResolver {
	return &DevelopmentAssetResolver{
		config: cfg,
		logger: logger,
	}
}

// ResolveAssetPath resolves asset paths for development using the Vite dev server
func (r *DevelopmentAssetResolver) ResolveAssetPath(_ context.Context, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	// Always use localhost for the browser, regardless of Vite's binding
	// Vite can bind to 0.0.0.0 for container access, but browser URLs should use localhost
	viteURL := fmt.Sprintf("%s://localhost:%s", r.config.App.Scheme, r.config.App.ViteDevPort)

	var resolvedPath string
	switch {
	case strings.HasPrefix(path, "@vite/") || strings.HasPrefix(path, "@fs/") || strings.HasPrefix(path, "@id/"):
		// Vite-specific paths
		resolvedPath = fmt.Sprintf("%s/%s", viteURL, path)
	case strings.HasPrefix(path, "src/"):
		// If path already starts with src/, use it as-is
		resolvedPath = fmt.Sprintf("%s/%s", viteURL, path)
	case strings.HasSuffix(path, ".css"):
		// For CSS files, check if path already starts with src/
		if strings.HasPrefix(path, "src/") {
			resolvedPath = fmt.Sprintf("%s/%s", viteURL, path)
		} else {
			resolvedPath = fmt.Sprintf("%s/src/css/%s", viteURL, path)
		}
	case strings.HasSuffix(path, ".ts"), strings.HasSuffix(path, ".js"):
		// For TypeScript/JavaScript files, check if path already starts with src/
		if strings.HasPrefix(path, "src/") {
			resolvedPath = fmt.Sprintf("%s/%s", viteURL, path)
		} else {
			baseName := strings.TrimSuffix(strings.TrimSuffix(path, ".js"), ".ts")
			// Special handling for main.ts/js - it's now in pages/
			if baseName == "main" {
				resolvedPath = fmt.Sprintf("%s/src/js/pages/%s.ts", viteURL, baseName)
			} else {
				resolvedPath = fmt.Sprintf("%s/src/js/%s.ts", viteURL, baseName)
			}
		}
	default:
		resolvedPath = fmt.Sprintf("%s/%s", viteURL, path)
	}

	// Add debug logging
	r.logger.Debug("development asset resolved",
		"input", path,
		"output", resolvedPath,
		"vite_url", viteURL,
	)

	return resolvedPath, nil
}

// loadManifestFromFS loads the manifest from the embedded filesystem
func loadManifestFromFS(distFS embed.FS, logger logging.Logger) (Manifest, error) {
	manifestPath := "dist/.vite/manifest.json"

	data, readErr := fs.ReadFile(distFS, manifestPath)
	if readErr != nil {
		return nil, fmt.Errorf("%w: %s", ErrManifestNotFound, readErr.Error())
	}

	var manifest Manifest
	if unmarshalErr := json.Unmarshal(data, &manifest); unmarshalErr != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidManifest, unmarshalErr.Error())
	}

	logger.Info("manifest loaded successfully",
		"entries", len(manifest),
	)
	return manifest, nil
}
