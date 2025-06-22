// Package web provides utilities for handling web assets in the application.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
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
func (r *ProductionAssetResolver) ResolveAssetPath(ctx context.Context, path string) (string, error) {
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
func (r *DevelopmentAssetResolver) ResolveAssetPath(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	hostPort := net.JoinHostPort(r.config.App.ViteDevHost, r.config.App.ViteDevPort)

	var resolvedPath string
	switch {
	case strings.HasPrefix(path, "src/"):
		// If path already starts with src/, use it as-is
		resolvedPath = fmt.Sprintf("%s://%s/%s", r.config.App.Scheme, hostPort, path)
	case strings.HasSuffix(path, ".css"):
		resolvedPath = fmt.Sprintf("%s://%s/src/css/%s", r.config.App.Scheme, hostPort, path)
	case strings.HasSuffix(path, ".ts"), strings.HasSuffix(path, ".js"):
		baseName := strings.TrimSuffix(strings.TrimSuffix(path, ".js"), ".ts")
		// Special handling for main.ts/js - it's now in pages/
		if baseName == "main" {
			resolvedPath = fmt.Sprintf("%s://%s/src/js/pages/%s.ts", r.config.App.Scheme, hostPort, baseName)
		} else {
			resolvedPath = fmt.Sprintf("%s://%s/src/js/%s.ts", r.config.App.Scheme, hostPort, baseName)
		}
	default:
		resolvedPath = fmt.Sprintf("%s://%s/%s", r.config.App.Scheme, hostPort, path)
	}

	return resolvedPath, nil
}

// getManifestKeys returns a slice of all keys in the manifest for debugging
func getManifestKeys(manifest Manifest) []string {
	keys := make([]string, 0, len(manifest))
	for key := range manifest {
		keys = append(keys, key)
	}
	return keys
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
