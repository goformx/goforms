// Package constants provides application-wide constants for HTTP status codes,
// paths, timeouts, headers, content types, and other configuration values.
// These constants are used throughout the application to ensure consistency
// and maintainability.
package constants

import "strings"

// PathManager provides centralized path management and validation
type PathManager struct {
	PublicPaths        []string
	StaticPaths        []string
	AdminPaths         []string
	APIValidationPaths []string
}

// NewPathManager creates a new path manager with all path constants
func NewPathManager() *PathManager {
	return &PathManager{
		PublicPaths: []string{
			PathLogin,
			PathSignup,
			PathHealth,
			PathMetrics,
			PathForgotPassword,
			PathResetPassword,
			PathVerifyEmail,
			PathAPIHealth,
			PathAPIValidation,
			PathAPIFormsLaravel, // Laravel assertion API: auth via X-User-Id/X-Signature on route group
		},
		StaticPaths: []string{
			PathStatic,
			PathAssets,
			PathImages,
			PathCSS,
			PathJS,
			PathFonts,
			PathFavicon,
			PathRobotsTxt,
		},
		AdminPaths: []string{
			PathAdmin,
			PathAdminUsers,
			PathAdminForms,
		},
		APIValidationPaths: []string{
			PathAPIValidation,
			PathAPIValidationLogin,
			PathAPIValidationSignup,
		},
	}
}

// IsPublicPath checks if a path is public (no authentication required)
func (pm *PathManager) IsPublicPath(path string) bool {
	return pm.containsPath(pm.PublicPaths, path)
}

// IsStaticPath checks if a path is for static assets
func (pm *PathManager) IsStaticPath(path string) bool {
	return pm.containsPath(pm.StaticPaths, path)
}

// IsAdminPath checks if a path requires admin access
func (pm *PathManager) IsAdminPath(path string) bool {
	return pm.containsPath(pm.AdminPaths, path)
}

// IsAPIValidationPath checks if a path is for API validation
func (pm *PathManager) IsAPIValidationPath(path string) bool {
	return pm.containsPath(pm.APIValidationPaths, path)
}

// containsPath checks if a path matches any of the provided path patterns
func (pm *PathManager) containsPath(patterns []string, path string) bool {
	for _, pattern := range patterns {
		if path == pattern || strings.HasPrefix(path, pattern+"/") {
			return true
		}
	}

	return false
}

// GetRequiredAccess determines the required access level for a path
func (pm *PathManager) GetRequiredAccess(path string) string {
	if pm.IsPublicPath(path) {
		return "public"
	}

	if pm.IsAdminPath(path) {
		return "admin"
	}

	return "authenticated" // default
}
