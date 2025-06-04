package middleware

import "strings"

// isStaticFile checks if the given path is a static file
func isStaticFile(path string) bool {
	// System files that should always be considered static
	if strings.HasPrefix(path, "/.well-known/") ||
		path == "/favicon.ico" ||
		path == "/robots.txt" {
		return true
	}

	// Application static files
	return strings.HasPrefix(path, "/public/") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".jpeg") ||
		strings.HasSuffix(path, ".gif") ||
		strings.HasSuffix(path, ".svg") ||
		strings.HasSuffix(path, ".woff") ||
		strings.HasSuffix(path, ".woff2") ||
		strings.HasSuffix(path, ".ttf") ||
		strings.HasSuffix(path, ".eot")
}
