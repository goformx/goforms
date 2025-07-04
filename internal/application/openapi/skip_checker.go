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
	for _, skipPath := range s.skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// Check skip methods
	for _, skipMethod := range s.skipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}
