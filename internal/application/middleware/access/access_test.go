package access_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
)

// generateTestRules creates access rules for testing purposes
func generateTestRules() []access.Rule {
	return []access.Rule{
		{Path: constants.PathHome, AccessLevel: access.Public},
		{Path: constants.PathLogin, AccessLevel: access.Public},
		{Path: constants.PathSignup, AccessLevel: access.Public},
		{Path: constants.PathDashboard, AccessLevel: access.Authenticated},
		{Path: constants.PathAdmin, AccessLevel: access.Admin},
		{Path: constants.PathForms, AccessLevel: access.Authenticated},
		{Path: "/forms/new", AccessLevel: access.Authenticated},
		{Path: "/forms/:id", AccessLevel: access.Authenticated},
		{Path: constants.PathProfile, AccessLevel: access.Authenticated},
		{Path: constants.PathSettings, AccessLevel: access.Authenticated},
		{Path: constants.PathDemo, AccessLevel: access.Public},
		{Path: constants.PathHealth, AccessLevel: access.Public},
		{Path: constants.PathMetrics, AccessLevel: access.Public},
		{Path: constants.PathAssets, AccessLevel: access.Public},
		{Path: constants.PathFonts, AccessLevel: access.Public},
		{Path: constants.PathCSS, AccessLevel: access.Public},
		{Path: constants.PathJS, AccessLevel: access.Public},
		{Path: constants.PathImages, AccessLevel: access.Public},
		{Path: constants.PathStatic, AccessLevel: access.Public},
		{Path: constants.PathFavicon, AccessLevel: access.Public},
		{Path: "/api/v1/validation/login", AccessLevel: access.Public},
		{Path: "/api/v1/forms", AccessLevel: access.Authenticated},
	}
}

func TestManager_IsPublicPath(t *testing.T) {
	config := access.DefaultConfig()
	manager := access.NewManager(config, nil)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "exact public path match",
			path:     "/login",
			expected: true,
		},
		{
			name:     "public path prefix match",
			path:     "/assets/images/logo.png",
			expected: true,
		},
		{
			name:     "non-public path",
			path:     "/dashboard",
			expected: false,
		},
		{
			name:     "root path",
			path:     "/",
			expected: false, // Not in default public paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.IsPublicPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_IsAdminPath(t *testing.T) {
	config := access.DefaultConfig()
	manager := access.NewManager(config, nil)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "exact admin path match",
			path:     "/admin",
			expected: true,
		},
		{
			name:     "admin path prefix match",
			path:     "/admin/users",
			expected: true,
		},
		{
			name:     "non-admin path",
			path:     constants.PathDashboard,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.IsAdminPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_GetRequiredAccess(t *testing.T) {
	config := access.DefaultConfig()
	manager := access.NewManager(config, generateTestRules())

	tests := []struct {
		name     string
		path     string
		method   string
		expected access.Level
	}{
		{
			name:     "public path",
			path:     constants.PathLogin,
			method:   "GET",
			expected: access.Public,
		},
		{
			name:     "authenticated path",
			path:     constants.PathDashboard,
			method:   "GET",
			expected: access.Authenticated,
		},
		{
			name:     "admin path",
			path:     constants.PathAdmin,
			method:   "GET",
			expected: access.Admin,
		},
		{
			name:     "unknown path defaults to authenticated",
			path:     "/unknown",
			method:   "GET",
			expected: access.Authenticated,
		},
		{
			name:     "public API validation endpoint",
			path:     "/api/v1/validation/login",
			method:   "GET",
			expected: access.Public,
		},
		{
			name:     "authenticated API endpoint",
			path:     "/api/v1/forms",
			method:   "GET",
			expected: access.Authenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.GetRequiredAccess(tt.path, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_AddRule(t *testing.T) {
	config := access.DefaultConfig()
	manager := access.NewManager(config, nil)

	// Add a custom rule
	rule := access.Rule{
		Path:        "/custom",
		AccessLevel: access.Admin,
		Methods:     []string{"GET", "POST"},
	}
	manager.AddRule(rule)

	// Test the added rule
	tests := []struct {
		name     string
		path     string
		method   string
		expected access.Level
	}{
		{
			name:     "custom path with allowed method",
			path:     "/custom",
			method:   "GET",
			expected: access.Admin,
		},
		{
			name:     "custom path with another allowed method",
			path:     "/custom",
			method:   "POST",
			expected: access.Admin,
		},
		{
			name:     "custom path with disallowed method",
			path:     "/custom",
			method:   "PUT",
			expected: access.Authenticated, // Default access
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.GetRequiredAccess(tt.path, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *access.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &access.Config{
				DefaultAccess: access.Authenticated,
				PublicPaths:   []string{constants.PathLogin},
				AdminPaths:    []string{constants.PathAdmin},
			},
			expectError: false,
		},
		{
			name: "invalid default access level",
			config: &access.Config{
				DefaultAccess: 999, // Invalid access level
				PublicPaths:   []string{constants.PathLogin},
				AdminPaths:    []string{constants.PathAdmin},
			},
			expectError: true,
		},
		{
			name: "valid public access level",
			config: &access.Config{
				DefaultAccess: access.Public,
				PublicPaths:   []string{constants.PathLogin},
				AdminPaths:    []string{constants.PathAdmin},
			},
			expectError: false,
		},
		{
			name: "valid admin access level",
			config: &access.Config{
				DefaultAccess: access.Admin,
				PublicPaths:   []string{constants.PathLogin},
				AdminPaths:    []string{constants.PathAdmin},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultRules(t *testing.T) {
	rules := generateTestRules()

	// Test that essential rules are present
	essentialPaths := map[string]access.Level{
		constants.PathHome:      access.Public,
		constants.PathLogin:     access.Public,
		constants.PathSignup:    access.Public,
		constants.PathDashboard: access.Authenticated,
		constants.PathAdmin:     access.Admin,
		constants.PathForms:     access.Authenticated,
		"/forms/new":            access.Authenticated,
		"/forms/:id":            access.Authenticated,
		constants.PathProfile:   access.Authenticated,
		constants.PathSettings:  access.Authenticated,
		constants.PathDemo:      access.Public,
		constants.PathHealth:    access.Public,
		constants.PathMetrics:   access.Public,
		constants.PathAssets:    access.Public,
		constants.PathFonts:     access.Public,
		constants.PathCSS:       access.Public,
		constants.PathJS:        access.Public,
		constants.PathImages:    access.Public,
		constants.PathStatic:    access.Public,
		constants.PathFavicon:   access.Public,
	}

	for path, expectedLevel := range essentialPaths {
		found := false

		for _, rule := range rules {
			if rule.Path == path {
				assert.Equal(t, expectedLevel, rule.AccessLevel, "Path %s should have access level %v", path, expectedLevel)

				found = true

				break
			}
		}

		assert.True(t, found, "Path %s should be in test rules", path)
	}
}
