package access_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
)

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
	manager := access.NewManager(config, access.DefaultRules())

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
			expected: access.Authenticated,
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
	rules := access.DefaultRules()

	// Test that essential rules are present
	essentialPaths := map[string]access.Level{
		constants.PathHome:      access.Public,
		constants.PathLogin:     access.Public,
		constants.PathSignup:    access.Public,
		constants.PathDashboard: access.Authenticated,
		constants.PathAdmin:     access.Admin,
		constants.PathForms:     access.Authenticated,
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
		assert.True(t, found, "Path %s should be in default rules", path)
	}
}
