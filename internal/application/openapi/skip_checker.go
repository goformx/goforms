package openapi

import "strings"

// skipConditionChecker implements SkipConditionChecker
type skipConditionChecker struct {
	skipPaths   []string
	skipMethods []string
}

// NewSkipConditionChecker creates a new skip condition checker
func NewSkipConditionChecker(config *Config) *skipConditionChecker {
	return &skipConditionChecker{
		skipPaths:   config.SkipPaths,
		skipMethods: config.SkipMethods,
	}
}

// ShouldSkip checks if validation should be skipped for this request
func (s *skipConditionChecker) ShouldSkip(path, method string) bool {
	// Check skip paths
	if s.shouldSkipByPath(path) {
		return true
	}

	// Check skip methods
	return s.shouldSkipByMethod(method)
}

// shouldSkipByPath checks if the path should be skipped
func (s *skipConditionChecker) shouldSkipByPath(path string) bool {
	for _, skipPath := range s.skipPaths {
		if s.matchesPath(path, skipPath) {
			return true
		}
	}

	return false
}

// shouldSkipByMethod checks if the method should be skipped
func (s *skipConditionChecker) shouldSkipByMethod(method string) bool {
	for _, skipMethod := range s.skipMethods {
		// Handle empty method - should not match anything
		if skipMethod == "" {
			continue
		}

		if method == skipMethod {
			return true
		}
	}

	return false
}

// matchesPath checks if a path matches a skip path pattern
func (s *skipConditionChecker) matchesPath(path, skipPath string) bool {
	// Handle empty path - should match everything
	if skipPath == "" {
		return true
	}

	// Handle root path - should match exactly or be a prefix
	if skipPath == "/" {
		return path == "/" || strings.HasPrefix(path, "/")
	}

	// For other paths, use proper prefix matching
	// This ensures "/health" matches "/health" and "/health/status" but not "/healthcheck"
	if path == skipPath {
		return true
	}

	if strings.HasPrefix(path, skipPath) {
		// Check if the next character is '/' or the path ends there
		pathLen := len(skipPath)

		return len(path) == pathLen || path[pathLen] == '/'
	}

	return false
}
