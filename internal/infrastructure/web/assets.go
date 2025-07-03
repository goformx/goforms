// Package web provides utilities for handling web assets in the application.
// It supports both development mode (using Vite dev server) and production mode
// (using built assets from the Vite manifest).
package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
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
	} else {
		// Load manifest for production
		manifest, err := loadManifestFromFS(distFS, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to load manifest: %w", err)
		}

		manager.resolver = NewProductionAssetResolver(manifest, logger)
	}

	return manager, nil
}

// AssetPath returns the resolved asset path for the given input path
// This is a convenience method that uses a background context
func (m *AssetManager) AssetPath(path string) string {
	ctx := context.Background()

	resolvedPath, err := m.ResolveAssetPath(ctx, path)
	if err != nil {
		m.logger.Error("failed to resolve asset path",
			"asset_path", path,
			"error", err,
		)

		// Return original path as fallback instead of empty string
		return path
	}

	m.logger.Debug("asset resolved", "asset_path", path, "resolved", resolvedPath)

	return resolvedPath
}

// ResolveAssetPath resolves asset paths with context and proper error handling
func (m *AssetManager) ResolveAssetPath(ctx context.Context, path string) (string, error) {
	if err := m.ValidatePath(path); err != nil {
		return "", fmt.Errorf("path validation failed: %w", err)
	}

	// Check cache first
	m.mu.RLock()

	if cachedPath, found := m.pathCache[path]; found {
		m.mu.RUnlock()

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

	return resolvedPath, nil
}

// GetAssetType returns the type of asset based on its path
func (m *AssetManager) GetAssetType(path string) AssetType {
	return GetAssetTypeFromPath(path)
}

// ClearCache clears the asset path cache
func (m *AssetManager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldSize := len(m.pathCache)
	m.pathCache = make(map[string]string)

	m.logger.Debug("asset cache cleared", "previous_size", oldSize)
}

// ValidatePath validates an asset path
func (m *AssetManager) ValidatePath(path string) error {
	return ValidateAssetPath(path)
}

// GetBaseURL returns the base URL for assets (useful for CSP headers)
func (m *AssetManager) GetBaseURL() string {
	if m.config.App.IsDevelopment() {
		// For development, return the Vite dev server URL
		host := m.config.App.ViteDevHost
		if host == "" || host == "0.0.0.0" {
			host = "localhost"
		}

		return strings.TrimSuffix(
			m.config.App.Scheme+"://"+host+":"+m.config.App.ViteDevPort,
			"/",
		)
	}

	// For production, return the configured server URL or default
	if m.config.App.URL != "" {
		return strings.TrimSuffix(m.config.App.URL, "/")
	}

	// Default fallback - construct from scheme and host
	if m.config.App.Scheme != "" && m.config.App.Host != "" {
		baseURL := m.config.App.Scheme + "://" + m.config.App.Host
		if m.config.App.Port != 80 && m.config.App.Port != 443 {
			baseURL += fmt.Sprintf(":%d", m.config.App.Port)
		}

		return baseURL
	}

	// Final fallback
	return ""
}

// GetCacheSize returns the current number of cached entries
func (m *AssetManager) GetCacheSize() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.pathCache)
}

// GetConfig returns the configuration (useful for testing)
func (m *AssetManager) GetConfig() *config.Config {
	return m.config
}

// GetResolver returns the underlying resolver (useful for testing)
func (m *AssetManager) GetResolver() AssetResolver {
	return m.resolver
}

// IsPathCached checks if a path is already cached
func (m *AssetManager) IsPathCached(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.pathCache[path]

	return exists
}

// NewModule creates a new web module with proper dependency injection
func NewModule(cfg *config.Config, logger logging.Logger, distFS embed.FS) (*Module, error) {
	// Create asset manager
	manager, err := NewAssetManager(cfg, logger, distFS)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset manager: %w", err)
	}

	// Create asset server
	var server AssetServer
	if cfg.App.IsDevelopment() {
		// In development, use development asset server for static files
		server = NewDevelopmentAssetServer(cfg, logger)
	} else {
		// In production, use embedded filesystem
		server = NewEmbeddedAssetServer(logger, distFS)
	}

	// Create asset resolver for the module
	var resolver AssetResolver
	if cfg.App.IsDevelopment() {
		resolver = NewDevelopmentAssetResolver(cfg, logger)
	} else {
		// Load manifest for production resolver
		manifest, manifestErr := loadManifestFromFS(distFS, logger)
		if manifestErr != nil {
			return nil, fmt.Errorf("failed to load manifest for resolver: %w", manifestErr)
		}

		resolver = NewProductionAssetResolver(manifest, logger)
	}

	return &Module{
		AssetManager:  manager,
		AssetServer:   server,
		AssetResolver: resolver,
	}, nil
}

// AssetManagerFactory creates asset managers with different configurations
type AssetManagerFactory struct {
	config *config.Config
	logger logging.Logger
}

// NewAssetManagerFactory creates a new asset manager factory
func NewAssetManagerFactory(cfg *config.Config, logger logging.Logger) *AssetManagerFactory {
	return &AssetManagerFactory{
		config: cfg,
		logger: logger,
	}
}

// CreateManager creates an asset manager for the current environment
func (f *AssetManagerFactory) CreateManager(distFS embed.FS) (*AssetManager, error) {
	return NewAssetManager(f.config, f.logger, distFS)
}

// CreateManagerWithResolver creates an asset manager with a specific resolver
func (f *AssetManagerFactory) CreateManagerWithResolver(resolver AssetResolver) *AssetManager {
	return &AssetManager{
		resolver:  resolver,
		pathCache: make(map[string]string),
		config:    f.config,
		logger:    f.logger,
	}
}
