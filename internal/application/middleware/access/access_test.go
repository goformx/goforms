package access_test

import (
	"testing"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessManager_IsPublicPath(t *testing.T) {
	config := access.DefaultConfig()
	manager := access.NewAccessManager(config, nil)

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

func TestAccessManager_IsAdminPath(t *testing.T) {
	config := access.DefaultConfig()
	manager := access.NewAccessManager(config, nil)

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
			path:     "/dashboard",
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

func TestAccessManager_GetRequiredAccess(t *testing.T) {
	config := access.DefaultConfig()
	manager := access.NewAccessManager(config, access.DefaultRules())

	tests := []struct {
		name     string
		path     string
		method   string
		expected access.AccessLevel
	}{
		{
			name:     "public path",
			path:     "/login",
			method:   "GET",
			expected: access.PublicAccess,
		},
		{
			name:     "authenticated path",
			path:     "/dashboard",
			method:   "GET",
			expected: access.AuthenticatedAccess,
		},
		{
			name:     "admin path",
			path:     "/admin/users",
			method:   "GET",
			expected: access.AdminAccess,
		},
		{
			name:     "unknown path defaults to authenticated",
			path:     "/unknown",
			method:   "GET",
			expected: access.AuthenticatedAccess,
		},
		{
			name:     "public API validation endpoint",
			path:     "/api/v1/validation/login",
			method:   "GET",
			expected: access.PublicAccess,
		},
		{
			name:     "authenticated API endpoint",
			path:     "/api/v1/forms",
			method:   "GET",
			expected: access.AuthenticatedAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.GetRequiredAccess(tt.path, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAccessManager_AddRule(t *testing.T) {
	config := access.DefaultConfig()
	manager := access.NewAccessManager(config, nil)

	// Add a custom rule
	rule := access.AccessRule{
		Path:        "/custom",
		AccessLevel: access.AdminAccess,
		Methods:     []string{"GET", "POST"},
	}
	manager.AddRule(rule)

	// Test the added rule
	tests := []struct {
		name     string
		path     string
		method   string
		expected access.AccessLevel
	}{
		{
			name:     "custom path with allowed method",
			path:     "/custom",
			method:   "GET",
			expected: access.AdminAccess,
		},
		{
			name:     "custom path with another allowed method",
			path:     "/custom",
			method:   "POST",
			expected: access.AdminAccess,
		},
		{
			name:     "custom path with disallowed method",
			path:     "/custom",
			method:   "PUT",
			expected: access.AuthenticatedAccess, // Default access
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
				DefaultAccess: access.AuthenticatedAccess,
				PublicPaths:   []string{"/login"},
				AdminPaths:    []string{"/admin"},
			},
			expectError: false,
		},
		{
			name: "invalid default access level",
			config: &access.Config{
				DefaultAccess: 999, // Invalid access level
				PublicPaths:   []string{"/login"},
				AdminPaths:    []string{"/admin"},
			},
			expectError: true,
		},
		{
			name: "valid public access level",
			config: &access.Config{
				DefaultAccess: access.PublicAccess,
				PublicPaths:   []string{"/login"},
				AdminPaths:    []string{"/admin"},
			},
			expectError: false,
		},
		{
			name: "valid admin access level",
			config: &access.Config{
				DefaultAccess: access.AdminAccess,
				PublicPaths:   []string{"/login"},
				AdminPaths:    []string{"/admin"},
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
	essentialPaths := map[string]access.AccessLevel{
		"/":                access.PublicAccess,
		"/login":           access.PublicAccess,
		"/signup":          access.PublicAccess,
		"/dashboard":       access.AuthenticatedAccess,
		"/admin":           access.AdminAccess,
		"/api/v1/forms":    access.AuthenticatedAccess,
		"/api/v1/admin":    access.AdminAccess,
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