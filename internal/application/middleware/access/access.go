// Package access provides access control middleware and utilities for the application.
package access

import (
	"fmt"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/common/errors"
)

// Level represents the level of access required for a route
type Level int

const (
	// Public means no authentication required
	Public Level = iota
	// Authenticated means user must be authenticated
	Authenticated
	// Admin means user must be an admin
	Admin
)

// Rule defines a rule for route access
type Rule struct {
	Path        string
	AccessLevel Level
	Methods     []string // If empty, applies to all methods
}

// Config holds the configuration for the access middleware
type Config struct {
	// DefaultAccess is the default access level for routes
	DefaultAccess Level
	// PublicPaths are paths that are always accessible
	PublicPaths []string
	// AdminPaths are paths that require admin access
	AdminPaths []string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultAccess: Authenticated,
		PublicPaths: []string{
			constants.PathLogin,
			constants.PathSignup,
			constants.PathForgotPassword,
			constants.PathResetPassword,
			constants.PathVerifyEmail,
			constants.PathAssets,
			constants.PathFonts,
			constants.PathCSS,
			constants.PathJS,
			constants.PathFavicon,
			constants.PathRobotsTxt,
			constants.PathStatic,
			constants.PathImages,
			constants.PathAPIFormsLaravel, // Laravel assertion API: auth via X-User-Id/X-Signature on route group
		},
		AdminPaths: []string{
			constants.PathAdmin,
		},
	}
}

// Manager manages access control rules
type Manager struct {
	config *Config
	rules  []Rule
}

// NewManager creates a new access manager
func NewManager(config *Config, rules []Rule) *Manager {
	return &Manager{
		config: config,
		rules:  rules,
	}
}

// AddRule adds a new access rule
func (am *Manager) AddRule(rule Rule) {
	am.rules = append(am.rules, rule)
}

// IsPublicPath checks if a path is public
func (am *Manager) IsPublicPath(path string) bool {
	for _, p := range am.config.PublicPaths {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}

	return false
}

// IsAdminPath checks if a path requires admin access
func (am *Manager) IsAdminPath(path string) bool {
	for _, p := range am.config.AdminPaths {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}

	return false
}

// matchPathPattern checks if a path matches a pattern with parameters
func matchPathPattern(pattern, path string) bool {
	// Split both pattern and path into segments
	patternSegments := strings.Split(pattern, "/")
	pathSegments := strings.Split(path, "/")

	// Check if they have the same number of segments
	if len(patternSegments) != len(pathSegments) {
		return false
	}

	// Compare each segment
	for i, patternSeg := range patternSegments {
		pathSeg := pathSegments[i]

		// If pattern segment starts with ":", it's a parameter - always match
		if strings.HasPrefix(patternSeg, ":") {
			continue
		}

		// Otherwise, segments must match exactly
		if patternSeg != pathSeg {
			return false
		}
	}

	return true
}

// GetRequiredAccess returns the required access level for a path and method
func (am *Manager) GetRequiredAccess(path, method string) Level {
	// Check if path is public
	if am.IsPublicPath(path) {
		return Public
	}

	// Check if path requires admin access
	if am.IsAdminPath(path) {
		return Admin
	}

	// Check specific rules with pattern matching
	for _, rule := range am.rules {
		if matchPathPattern(rule.Path, path) {
			// If no methods specified, rule applies to all methods
			if len(rule.Methods) == 0 {
				return rule.AccessLevel
			}
			// Check if method is in the allowed methods
			for _, m := range rule.Methods {
				if m == method {
					return rule.AccessLevel
				}
			}
		}
	}

	// Default to requiring authentication if no rule matches
	return am.config.DefaultAccess
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DefaultAccess < Public || c.DefaultAccess > Admin {
		return errors.New(errors.ErrCodeValidation, "invalid default access level", nil)
	}

	return nil
}

// DefaultRules returns the default access rules for the application
func DefaultRules() []Rule {
	return []Rule{
		// Public paths
		{Path: constants.PathLogin, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathSignup, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathForgotPassword, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathResetPassword, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathVerifyEmail, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathHealth, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathMetrics, AccessLevel: Public, Methods: []string{}},

		// Static asset paths
		{Path: constants.PathAssets, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathFonts, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathCSS, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathJS, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathImages, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathStatic, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathFavicon, AccessLevel: Public, Methods: []string{}},
		{Path: constants.PathRobotsTxt, AccessLevel: Public, Methods: []string{}},

		// Authenticated paths
		{Path: constants.PathDashboard, AccessLevel: Authenticated, Methods: []string{}},
		{Path: constants.PathForms, AccessLevel: Authenticated, Methods: []string{}},
		{Path: constants.PathProfile, AccessLevel: Authenticated, Methods: []string{}},
		{Path: constants.PathSettings, AccessLevel: Authenticated, Methods: []string{}},

		// Admin paths
		{Path: constants.PathAdmin, AccessLevel: Admin, Methods: []string{}},
	}
}

// TestMatchPathPattern is a simple test function to verify pattern matching
func TestMatchPathPattern() {
	testCases := []struct {
		pattern string
		path    string
		expect  bool
	}{
		{"/api/v1/forms/:id/schema", "/api/v1/forms/61af2a0f-5b54-476f-9bf6-c2ee6ce5b822/schema", true},
		{"/api/v1/forms/:id/schema", "/api/v1/forms/123/schema", true},
		{"/api/v1/forms/:id/schema", "/api/v1/forms/123/submit", false},
		{"/api/v1/forms/:id/schema", "/api/v1/forms/schema", false},
		{"/api/v1/forms/:id/schema", "/api/v1/forms/123/schema/extra", false},
	}

	for _, tc := range testCases {
		result := matchPathPattern(tc.pattern, tc.path)
		if result != tc.expect {
			panic(fmt.Sprintf("Pattern matching failed: %s vs %s, expected %v, got %v",
				tc.pattern, tc.path, tc.expect, result))
		}
	}
}
