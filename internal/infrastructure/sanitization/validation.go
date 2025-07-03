// Package sanitization provides utilities for cleaning and validating user input
// to prevent XSS attacks, injection attacks, and other security vulnerabilities.
package sanitization

import "strings"

// IsValidEmail checks if an email is valid after sanitization
func IsValidEmail(s ServiceInterface, input string) bool {
	sanitized := s.Email(input)

	return sanitized != "" && strings.Contains(sanitized, "@")
}

// IsValidURL checks if a URL is valid after sanitization
func IsValidURL(s ServiceInterface, input string) bool {
	sanitized := s.URL(input)

	return sanitized != "" && (strings.HasPrefix(sanitized, "http://") || strings.HasPrefix(sanitized, "https://"))
}
