package web_test

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/web"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
)

func TestDevelopmentAssetResolver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		App: config.AppConfig{
			Scheme:      "http",
			ViteDevHost: "localhost",
			ViteDevPort: "5173",
		},
	}
	mockLogger := mocklogging.NewMockLogger(ctrl)

	mockLogger.EXPECT().Debug(
		"development asset resolved",
		"input", gomock.Any(),
		"output", gomock.Any(),
		"base_url", gomock.Any(),
		"cached", gomock.Any(),
	).AnyTimes()

	resolver := web.NewDevelopmentAssetResolver(cfg, mockLogger)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "src file",
			path:     "src/js/main.ts",
			expected: "http://localhost:5173/src/js/main.ts",
		},
		{
			name:     "css file",
			path:     "main.css",
			expected: "http://localhost:5173/src/css/main.css",
		},
		{
			name:     "js file",
			path:     "main.js",
			expected: "http://localhost:5173/src/js/main.ts",
		},
		{
			name:     "ts file",
			path:     "main.ts",
			expected: "http://localhost:5173/src/js/main.ts",
		},
		{
			name:     "other file",
			path:     "image.png",
			expected: "http://localhost:5173/image.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveAssetPath(t.Context(), tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProductionAssetResolver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manifest := web.Manifest{
		"main.js": {
			File: "assets/main.abc123.js",
		},
		"style.css": {
			File: "assets/style.def456.css",
		},
	}
	mockLogger := mocklogging.NewMockLogger(ctrl)

	mockLogger.EXPECT().Debug(
		"production asset resolved",
		"input", gomock.Any(),
		"output", gomock.Any(),
		"manifest_entry", gomock.Any(),
	).AnyTimes()

	mockLogger.EXPECT().Warn(
		"asset not found in manifest",
		"path", gomock.Any(),
		"available_assets", gomock.Any(),
	).AnyTimes()

	resolver := web.NewProductionAssetResolver(manifest, mockLogger)

	tests := []struct {
		name        string
		path        string
		expected    string
		expectError bool
	}{
		{
			name:     "existing js file",
			path:     "main.js",
			expected: "/assets/main.abc123.js",
		},
		{
			name:     "existing css file",
			path:     "style.css",
			expected: "/assets/style.def456.css",
		},
		{
			name:        "non-existent file",
			path:        "missing.js",
			expectError: true,
		},
		{
			name:        "empty path",
			path:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.ResolveAssetPath(t.Context(), tt.path)
			if tt.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAssetManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "development",
			Scheme:      "http",
			ViteDevHost: "localhost",
			ViteDevPort: "5173",
		},
	}
	mockLogger := mocklogging.NewMockLogger(ctrl)

	// Expect debug calls from DevelopmentAssetResolver
	mockLogger.EXPECT().Debug(
		"development asset resolved",
		"input", gomock.Any(),
		"output", gomock.Any(),
		"base_url", gomock.Any(),
		"cached", gomock.Any(),
	).AnyTimes()

	// Expect debug calls from AssetManager
	mockLogger.EXPECT().Debug(
		"asset resolved", "asset_path", gomock.Any(), "resolved", gomock.Any(),
	).AnyTimes()

	// Expect debug calls for cache clearing
	mockLogger.EXPECT().Debug(
		"asset cache cleared",
		"previous_size", gomock.Any(),
	).AnyTimes()

	var distFS embed.FS

	manager, err := web.NewAssetManager(cfg, mockLogger, distFS)
	require.NoError(t, err)
	assert.NotNil(t, manager)

	path := manager.AssetPath("main.js")
	assert.Equal(t, "http://localhost:5173/src/js/main.ts", path)

	path2 := manager.AssetPath("main.js")
	assert.Equal(t, path, path2)

	assetType := manager.GetAssetType("main.js")
	assert.Equal(t, web.AssetTypeJS, assetType)

	assetType = manager.GetAssetType("style.css")
	assert.Equal(t, web.AssetTypeCSS, assetType)

	manager.ClearCache()
}

func TestAssetManager_ProductionMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "production",
		},
	}
	mockLogger := mocklogging.NewMockLogger(ctrl)

	var distFS embed.FS

	manager, err := web.NewAssetManager(cfg, mockLogger, distFS)
	require.Error(t, err)
	require.Nil(t, manager)
}

func TestAssetManager_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocklogging.NewMockLogger(ctrl)

	manager, err := web.NewAssetManager(nil, mockLogger, embed.FS{})
	require.Error(t, err)
	require.Nil(t, manager)
	require.Contains(t, err.Error(), "config is required")

	cfg := &config.Config{
		App: config.AppConfig{
			Environment: "development",
		},
	}
	manager, err = web.NewAssetManager(cfg, nil, embed.FS{})
	require.Error(t, err)
	require.Nil(t, manager)
	require.Contains(t, err.Error(), "logger is required")
}
