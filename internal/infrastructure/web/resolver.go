// Package web provides utilities for handling web assets in the application.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"sync"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// AssetPathRule defines a rule for transforming asset paths
type AssetPathRule struct {
	Condition func(string) bool
	Transform func(string, string) string
}

// ProductionAssetResolver handles production asset resolution
type ProductionAssetResolver struct {
	manifest Manifest
	logger   logging.Logger
	mu       sync.RWMutex
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
	if err := validateAssetPath(path); err != nil {
		return "", fmt.Errorf("production asset resolution failed: %w", err)
	}

	r.mu.RLock()
	entry, found := r.manifest[path]
	r.mu.RUnlock()

	if !found {
		r.logger.Warn("asset not found in manifest",
			"path", path,
			"available_assets", len(r.manifest))

		return "", fmt.Errorf("%w: %s not found in manifest", ErrAssetNotFound, path)
	}

	assetPath := normalizeAssetPath(entry.File)

	r.logger.Debug("production asset resolved",
		"input", path,
		"output", assetPath,
		"manifest_entry", entry.File)

	return assetPath, nil
}

// DevelopmentAssetResolver handles development asset resolution
type DevelopmentAssetResolver struct {
	config    *config.Config
	logger    logging.Logger
	rules     []AssetPathRule
	baseURL   string
	pathCache map[string]string
	cacheMu   sync.RWMutex
}

// NewDevelopmentAssetResolver creates a new development asset resolver
func NewDevelopmentAssetResolver(cfg *config.Config, logger logging.Logger) *DevelopmentAssetResolver {
	resolver := &DevelopmentAssetResolver{
		config:    cfg,
		logger:    logger,
		pathCache: make(map[string]string),
	}

	resolver.baseURL = resolver.buildBaseURL()
	resolver.rules = resolver.buildPathRules()

	return resolver
}

// ResolveAssetPath resolves asset paths for development using the Vite dev server
func (r *DevelopmentAssetResolver) ResolveAssetPath(ctx context.Context, path string) (string, error) {
	if err := validateAssetPath(path); err != nil {
		return "", fmt.Errorf("development asset resolution failed: %w", err)
	}

	// Check cache first
	r.cacheMu.RLock()

	if cached, exists := r.pathCache[path]; exists {
		r.cacheMu.RUnlock()

		return cached, nil
	}

	r.cacheMu.RUnlock()

	// Apply transformation rules
	resolvedPath := r.applyPathRules(path)

	// Cache the result
	r.cacheMu.Lock()
	r.pathCache[path] = resolvedPath
	r.cacheMu.Unlock()

	r.logger.Debug("development asset resolved",
		"input", path,
		"output", resolvedPath,
		"base_url", r.baseURL,
		"cached", false)

	return resolvedPath, nil
}

// ClearCache clears the path resolution cache
func (r *DevelopmentAssetResolver) ClearCache() {
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()

	r.pathCache = make(map[string]string)
	r.logger.Debug("development asset cache cleared")
}

// buildBaseURL constructs the base URL for the Vite dev server
func (r *DevelopmentAssetResolver) buildBaseURL() string {
	// Use the configured host, fallback to localhost for browser compatibility
	host := r.config.App.ViteDevHost
	if host == "" || host == "0.0.0.0" {
		host = "localhost"
	}

	return fmt.Sprintf("%s://%s:%s",
		r.config.App.Scheme,
		host,
		r.config.App.ViteDevPort)
}

// buildPathRules creates the asset path transformation rules
func (r *DevelopmentAssetResolver) buildPathRules() []AssetPathRule {
	return []AssetPathRule{
		{
			// Vite-specific paths (already correctly formatted)
			Condition: func(path string) bool {
				return strings.HasPrefix(path, "@vite/") ||
					strings.HasPrefix(path, "@fs/") ||
					strings.HasPrefix(path, "@id/")
			},
			Transform: func(path, baseURL string) string {
				return fmt.Sprintf("%s/%s", baseURL, path)
			},
		},
		{
			// Paths already starting with src/
			Condition: func(path string) bool {
				return strings.HasPrefix(path, "src/")
			},
			Transform: func(path, baseURL string) string {
				return fmt.Sprintf("%s/%s", baseURL, path)
			},
		},
		{
			// CSS files
			Condition: func(path string) bool {
				return strings.HasSuffix(path, ".css")
			},
			Transform: func(path, baseURL string) string {
				return fmt.Sprintf("%s/src/css/%s", baseURL, path)
			},
		},
		{
			// TypeScript/JavaScript files with special main.ts handling
			Condition: func(path string) bool {
				return strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".js")
			},
			Transform: func(path, baseURL string) string {
				baseName := strings.TrimSuffix(strings.TrimSuffix(path, ".js"), ".ts")

				return fmt.Sprintf("%s/src/js/%s.ts", baseURL, baseName)
			},
		},
		{
			// Default fallback
			Condition: func(path string) bool {
				return true // Always matches as fallback
			},
			Transform: func(path, baseURL string) string {
				return fmt.Sprintf("%s/%s", baseURL, path)
			},
		},
	}
}

// applyPathRules applies the first matching rule to transform the path
func (r *DevelopmentAssetResolver) applyPathRules(path string) string {
	for _, rule := range r.rules {
		if rule.Condition(path) {
			return rule.Transform(path, r.baseURL)
		}
	}

	// Should never reach here due to fallback rule, but just in case
	return fmt.Sprintf("%s/%s", r.baseURL, path)
}

// validateAssetPath validates that an asset path is not empty
func validateAssetPath(path string) error {
	if path == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	return nil
}

// normalizeAssetPath ensures the path starts with a forward slash
func normalizeAssetPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}

	return path
}

// loadManifestFromFS loads the manifest from the embedded filesystem
func loadManifestFromFS(distFS embed.FS, logger logging.Logger) (Manifest, error) {
	const manifestPath = "dist/.vite/manifest.json"

	data, readErr := fs.ReadFile(distFS, manifestPath)
	if readErr != nil {
		return nil, fmt.Errorf("%w: failed to read %s: %w", ErrManifestNotFound, manifestPath, readErr)
	}

	var manifest Manifest
	if unmarshalErr := json.Unmarshal(data, &manifest); unmarshalErr != nil {
		return nil, fmt.Errorf("%w: failed to parse %s: %w", ErrInvalidManifest, manifestPath, unmarshalErr)
	}

	if len(manifest) == 0 {
		logger.Warn("manifest is empty", "path", manifestPath)
	} else {
		logger.Info("manifest loaded successfully",
			"path", manifestPath,
			"entries", len(manifest))
	}

	return manifest, nil
}

// AssetResolverFactory creates the appropriate resolver based on environment
type AssetResolverFactory struct {
	config *config.Config
	logger logging.Logger
}

// NewAssetResolverFactory creates a new factory
func NewAssetResolverFactory(cfg *config.Config, logger logging.Logger) *AssetResolverFactory {
	return &AssetResolverFactory{
		config: cfg,
		logger: logger,
	}
}

// CreateResolver creates the appropriate resolver for the current environment
func (f *AssetResolverFactory) CreateResolver(distFS embed.FS) (AssetResolver, error) {
	if f.config.App.IsDevelopment() {
		f.logger.Info("creating development asset resolver")

		return NewDevelopmentAssetResolver(f.config, f.logger), nil
	}

	f.logger.Info("creating production asset resolver")

	manifest, err := loadManifestFromFS(distFS, f.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create production resolver: %w", err)
	}

	return NewProductionAssetResolver(manifest, f.logger), nil
}
