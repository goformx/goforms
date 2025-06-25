// Package sensitive provides utilities for detecting and handling sensitive data
// in log messages and other text content. It includes pattern matching for
// common sensitive data types like emails, passwords, and tokens.
package sensitive

import "strings"

// IsKey returns true if the key matches any sensitive pattern
func IsKey(key string) bool {
	keyLower := strings.ToLower(key)
	for _, pattern := range Patterns {
		if strings.Contains(keyLower, pattern) {
			return true
		}
	}
	return false
}

// MaskValue returns the standard mask for sensitive values
func MaskValue() string {
	return "****"
}
