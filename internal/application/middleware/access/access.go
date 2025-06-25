// Package access provides access control middleware and utilities for the application.
package access

import (
	"fmt"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/common/errors"
)

// AccessLevel represents the level of access required for a route
type AccessLevel int

const (
	// PublicAccess means no authentication required
	PublicAccess AccessLevel = iota
	// AuthenticatedAccess means user must be authenticated
	AuthenticatedAccess
	// AdminAccess means user must be an admin
	AdminAccess
)

// AccessRule defines a rule for route access
type AccessRule struct {
	Path        string
	AccessLevel AccessLevel
	Methods     []string // If empty, applies to all methods
}

// Config holds the configuration for the access middleware
type Config struct {
	// DefaultAccess is the default access level for routes
	DefaultAccess AccessLevel
	// PublicPaths are paths that are always accessible
	PublicPaths []string
	// AdminPaths are paths that require admin access
	AdminPaths []string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultAccess: AuthenticatedAccess,
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
		},
		AdminPaths: []string{
			constants.PathAdmin,
		},
	}
}

// AccessManager manages access control rules
type AccessManager struct {
	config *Config
	rules  []AccessRule
}

// NewAccessManager creates a new access manager
func NewAccessManager(config *Config, rules []AccessRule) *AccessManager {
	return &AccessManager{
		config: config,
		rules:  rules,
	}
}

// AddRule adds a new access rule
func (am *AccessManager) AddRule(rule AccessRule) {
	am.rules = append(am.rules, rule)
}

// IsPublicPath checks if a path is public
func (am *AccessManager) IsPublicPath(path string) bool {
	// Check exact matches first
	for _, p := range am.config.PublicPaths {
		if path == p {
			return true
		}
	}

	// Check if path starts with any public path
	for _, p := range am.config.PublicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

// IsAdminPath checks if a path requires admin access
func (am *AccessManager) IsAdminPath(path string) bool {
	// Check exact matches first
	for _, p := range am.config.AdminPaths {
		if path == p {
			return true
		}
	}

	// Check if path starts with any admin path
	for _, p := range am.config.AdminPaths {
		if strings.HasPrefix(path, p) {
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
func (am *AccessManager) GetRequiredAccess(path, method string) AccessLevel {
	// Check if path is public
	if am.IsPublicPath(path) {
		return PublicAccess
	}

	// Check if path requires admin access
	if am.IsAdminPath(path) {
		return AdminAccess
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
	if c.DefaultAccess < PublicAccess || c.DefaultAccess > AdminAccess {
		return errors.New(errors.ErrCodeValidation, "invalid default access level", nil)
	}
	return nil
}

// DefaultRules returns the default access rules for the application
func DefaultRules() []AccessRule {
	return []AccessRule{
		// Public paths
		{Path: constants.PathHome, AccessLevel: PublicAccess},
		{Path: constants.PathLogin, AccessLevel: PublicAccess},
		{Path: constants.PathSignup, AccessLevel: PublicAccess},
		{Path: constants.PathForgotPassword, AccessLevel: PublicAccess},
		{Path: constants.PathResetPassword, AccessLevel: PublicAccess},
		{Path: constants.PathVerifyEmail, AccessLevel: PublicAccess},
		{Path: constants.PathDemo, AccessLevel: PublicAccess},
		{Path: constants.PathHealth, AccessLevel: PublicAccess},
		{Path: constants.PathMetrics, AccessLevel: PublicAccess},

		// Static asset paths
		{Path: constants.PathAssets, AccessLevel: PublicAccess},
		{Path: constants.PathFonts, AccessLevel: PublicAccess},
		{Path: constants.PathCSS, AccessLevel: PublicAccess},
		{Path: constants.PathJS, AccessLevel: PublicAccess},
		{Path: constants.PathImages, AccessLevel: PublicAccess},
		{Path: constants.PathStatic, AccessLevel: PublicAccess},
		{Path: constants.PathFavicon, AccessLevel: PublicAccess},
		{Path: constants.PathRobotsTxt, AccessLevel: PublicAccess},

		// Authenticated paths
		{Path: constants.PathDashboard, AccessLevel: AuthenticatedAccess},
		{Path: constants.PathForms, AccessLevel: AuthenticatedAccess},
		{Path: constants.PathProfile, AccessLevel: AuthenticatedAccess},
		{Path: constants.PathSettings, AccessLevel: AuthenticatedAccess},

		// Admin paths
		{Path: constants.PathAdmin, AccessLevel: AdminAccess},
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
