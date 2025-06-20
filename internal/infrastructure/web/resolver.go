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

	r.logger.Debug("resolving production asset path",
		"path", path,
		"manifest_entries", len(r.manifest),
	)

	entry, found := r.manifest[path]
	if !found {
		r.logger.Debug("asset not found in manifest",
			"path", path,
			"available_keys", strings.Join(getManifestKeys(r.manifest), ", "),
		)
		return "", fmt.Errorf("%w: %s", ErrAssetNotFound, path)
	}

	assetPath := entry.File
	if !strings.HasPrefix(assetPath, "/") {
		assetPath = "/" + assetPath
	}

	r.logger.Debug("production asset resolved",
		"input", path,
		"output", assetPath,
	)
	return assetPath, nil
}

// DevelopmentAssetResolver handles development asset resolution
type DevelopmentAssetResolver struct {
	config *config.Config
	logger logging.Logger
}

// NewDevelopmentAssetResolver creates a new development asset resolver
func NewDevelopmentAssetResolver(config *config.Config, logger logging.Logger) *DevelopmentAssetResolver {
	return &DevelopmentAssetResolver{
		config: config,
		logger: logger,
	}
}

// ResolveAssetPath resolves asset paths for development using the Vite dev server
func (r *DevelopmentAssetResolver) ResolveAssetPath(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	hostPort := net.JoinHostPort(r.config.App.ViteDevHost, r.config.App.ViteDevPort)

	r.logger.Debug("resolving development asset path",
		"path", path,
		"host_port", hostPort,
	)

	var resolvedPath string
	switch {
	case strings.HasPrefix(path, "src/"):
		resolvedPath = fmt.Sprintf("%s://%s/%s", r.config.App.Scheme, hostPort, path)
	case strings.HasSuffix(path, ".css"):
		resolvedPath = fmt.Sprintf("%s://%s/src/css/%s", r.config.App.Scheme, hostPort, path)
	case strings.HasSuffix(path, ".ts"), strings.HasSuffix(path, ".js"):
		baseName := strings.TrimSuffix(strings.TrimSuffix(path, ".js"), ".ts")
		resolvedPath = fmt.Sprintf("%s://%s/src/js/%s.ts", r.config.App.Scheme, hostPort, baseName)
	default:
		resolvedPath = fmt.Sprintf("%s://%s/%s", r.config.App.Scheme, hostPort, path)
	}

	r.logger.Debug("development asset resolved",
		"input", path,
		"output", resolvedPath,
	)
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
	logger.Debug("loading manifest from embedded filesystem",
		"path", manifestPath,
	)

	data, err := fs.ReadFile(distFS, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrManifestNotFound, err.Error())
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidManifest, err.Error())
	}

	logger.Info("manifest loaded successfully",
		"entries", len(manifest),
	)
	return manifest, nil
}
