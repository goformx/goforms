// Package web provides utilities for handling web assets in the application.
// It supports both development mode (using Vite dev server) and production mode
// (using built assets from the Vite manifest).
package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// AssetManager handles asset path resolution and caching
type AssetManager struct {
	resolver  AssetResolver
	pathCache map[string]string
	mu        sync.RWMutex
	logger    logging.Logger
	config    *config.Config
}

// NewAssetManager creates a new asset manager with proper dependency injection
func NewAssetManager(cfg *config.Config, logger logging.Logger, distFS embed.FS) (*AssetManager, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	manager := &AssetManager{
		pathCache: make(map[string]string),
		config:    cfg,
		logger:    logger,
	}

	// Create appropriate resolver based on environment
	if cfg.App.IsDevelopment() {
		manager.resolver = NewDevelopmentAssetResolver(cfg, logger)
		logger.Info("asset manager initialized in development mode")
	} else {
		// Load manifest for production
		manifest, err := loadManifestFromFS(distFS, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to load manifest: %w", err)
		}
		manager.resolver = NewProductionAssetResolver(manifest, logger)
		logger.Info("asset manager initialized in production mode",
			"manifest_entries", len(manifest),
		)
	}

	return manager, nil
}

// AssetPath returns the resolved asset path for the given input path
func (m *AssetManager) AssetPath(path string) string {
	ctx := context.Background()
	resolvedPath, err := m.ResolveAssetPath(ctx, path)
	if err != nil {
		m.logger.Error("failed to resolve asset path",
			"path", path,
			"error", err,
		)
		return ""
	}
	return resolvedPath
}

// ResolveAssetPath resolves asset paths with context and proper error handling
func (m *AssetManager) ResolveAssetPath(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	m.logger.Debug("asset manager resolving path",
		"input_path", path,
	)

	// Check cache first
	m.mu.RLock()
	if cachedPath, found := m.pathCache[path]; found {
		m.mu.RUnlock()
		m.logger.Debug("asset path found in cache",
			"input_path", path,
			"cached_path", cachedPath,
		)
		return cachedPath, nil
	}
	m.mu.RUnlock()

	// Resolve the path using the appropriate resolver
	resolvedPath, err := m.resolver.ResolveAssetPath(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve asset path: %w", err)
	}

	// Cache the result
	m.mu.Lock()
	m.pathCache[path] = resolvedPath
	m.mu.Unlock()

	m.logger.Debug("asset path resolved",
		"input_path", path,
		"resolved_path", resolvedPath,
		"environment", m.config.App.Env,
	)

	return resolvedPath, nil
}

// GetAssetType returns the type of asset based on its path
func (m *AssetManager) GetAssetType(path string) AssetType {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".mjs":
		return AssetTypeJS
	case ".css", ".scss", ".sass":
		return AssetTypeCSS
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return AssetTypeImage
	case ".woff", ".woff2", ".ttf", ".eot", ".otf":
		return AssetTypeFont
	default:
		return ""
	}
}

// ClearCache clears the asset path cache
func (m *AssetManager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pathCache = make(map[string]string)
	m.logger.Debug("asset path cache cleared")
}

// NewWebModule creates a new web module with proper dependency injection
func NewWebModule(cfg *config.Config, logger logging.Logger, distFS embed.FS) (*WebModule, error) {
	manager, err := NewAssetManager(cfg, logger, distFS)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset manager: %w", err)
	}

	var server AssetServer
	if cfg.App.IsDevelopment() {
		server = NewViteAssetServer(cfg, logger)
	} else {
		// In production, always use embedded filesystem
		server = NewEmbeddedAssetServer(logger, distFS)
	}

	return &WebModule{
		AssetManager: manager,
		AssetServer:  server,
	}, nil
}
