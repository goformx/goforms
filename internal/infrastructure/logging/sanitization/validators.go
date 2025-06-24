package sanitization

import (
	"strings"
)

// validatePath checks if a string is a valid URL path
func validatePath(path string) bool {
	if len(path) > MaxPathLength {
		return false
	}
	// Basic path validation - should start with / and contain only valid characters
	if path == "" || path[0] != '/' {
		return false
	}

	// Check for potentially dangerous characters
	dangerousChars := []string{"\\", "<", ">", "\"", "'", "\x00", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") || strings.Contains(path, "//") {
		return false
	}

	return true
}

// validateUserAgent checks if a string is a valid user agent
func validateUserAgent(userAgent string) bool {
	if len(userAgent) > MaxStringLength {
		return false
	}

	// Check for potentially dangerous characters in user agent
	dangerousChars := []string{"\x00", "\n", "\r", "<", ">", "\"", "'"}
	for _, char := range dangerousChars {
		if strings.Contains(userAgent, char) {
			return false
		}
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{"<script", "javascript:", "vbscript:", "onload=", "onerror="}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(userAgent), pattern) {
			return false
		}
	}

	return true
}
