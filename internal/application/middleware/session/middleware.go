package session

import (
	"net/http"
	"strings"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/context"
	"github.com/labstack/echo/v4"
)

// Middleware creates a new session middleware
func (sm *Manager) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			method := c.Request().Method

			// Special handling for schema endpoints
			if sm.isSchemaEndpoint(path, method) {
				return sm.handleSchemaEndpoint(c, method, next)
			}

			// Check if this is a path that should skip session processing entirely
			if sm.shouldSkipSession(path) {
				return next(c)
			}

			// Process session
			return sm.processSession(c, path, next)
		}
	}
}

// isSchemaEndpoint checks if this is a schema endpoint
func (sm *Manager) isSchemaEndpoint(path, _ string) bool {
	return strings.HasSuffix(path, "/schema") && strings.HasPrefix(path, "/api/v1/forms/")
}

// handleSchemaEndpoint handles schema endpoint requests
func (sm *Manager) handleSchemaEndpoint(c echo.Context, method string, next echo.HandlerFunc) error {
	// For GET requests to schema endpoints, allow without authentication (for embedded forms)
	if method == "GET" {
		return next(c)
	}
	// For PUT requests to schema endpoints, require authentication
	// Continue with normal session processing
	return sm.processSession(c, c.Request().URL.Path, next)
}

// processSession handles session processing logic
func (sm *Manager) processSession(c echo.Context, path string, next echo.HandlerFunc) error {
	// Always try to get session cookie and validate it
	cookie, err := c.Cookie(sm.cookieName)
	if err != nil {
		// For public paths, continue without authentication
		if sm.isPublicPath(path) {
			return next(c)
		}
		return sm.handleAuthError(c, "no session found")
	}

	// Get session from manager
	session, exists := sm.GetSession(cookie.Value)
	if !exists {
		// For public paths, continue without authentication
		if sm.isPublicPath(path) {
			return next(c)
		}
		return sm.handleAuthError(c, "invalid session")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		sm.DeleteSession(cookie.Value)
		// For public paths, continue without authentication
		if sm.isPublicPath(path) {
			return next(c)
		}
		return sm.handleAuthError(c, "session expired")
	}

	// Store session in context (always do this if we have a valid session)
	c.Set(string(context.SessionKey), session)
	context.SetUserID(c, session.UserID)
	context.SetEmail(c, session.Email)
	context.SetRole(c, session.Role)

	return next(c)
}

// shouldSkipSession checks if a path should skip session processing entirely
func (sm *Manager) shouldSkipSession(path string) bool {
	// Skip session processing for static files and assets
	if sm.isStaticFile(path) {
		return true
	}

	// Skip session processing for API endpoints that don't need authentication
	if strings.HasPrefix(path, "/api/v1/public/") {
		return true
	}

	// Skip session processing for validation endpoints
	if strings.HasPrefix(path, "/api/v1/validation/") {
		return true
	}

	// Skip session processing for health check or monitoring endpoints
	if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/metrics") {
		return true
	}

	// Skip session processing for development tool endpoints
	if sm.config.Config.App.Env == "development" && (strings.HasPrefix(path, "/.well-known/") ||
		strings.HasPrefix(path, "/debug/") ||
		strings.HasPrefix(path, "/dev/")) {
		return true
	}

	// Skip session processing for exempt paths
	for _, exemptPath := range sm.config.ExemptPaths {
		if strings.HasPrefix(path, exemptPath) {
			return true
		}
	}

	return false
}

// isPublicPath checks if a path is public (but still needs session checking)
func (sm *Manager) isPublicPath(path string) bool {
	// Use accessManager to check if the path is public
	if sm.accessManager != nil {
		if sm.accessManager.GetRequiredAccess(path, "GET") == access.PublicAccess {
			return true
		}
	}

	// Check public paths from config
	for _, publicPath := range sm.config.PublicPaths {
		if path == publicPath {
			return true
		}
	}

	return false
}

// isStaticFile checks if a path corresponds to a static file
func (sm *Manager) isStaticFile(path string) bool {
	// List of static file extensions
	staticExtensions := []string{
		".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".ico",
		".svg", ".woff", ".woff2", ".ttf", ".eot", ".otf",
		".pdf", ".txt", ".xml", ".json", ".webp", ".webm",
		".mp4", ".mp3", ".wav", ".ogg", ".map",
	}

	// Check if the path ends with any static file extension
	for _, ext := range staticExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	// Check if the path starts with common static asset paths
	staticPaths := []string{
		"/assets/",
		"/static/",
		"/images/",
		"/css/",
		"/js/",
		"/fonts/",
	}

	for _, prefix := range staticPaths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// handleAuthError handles authentication errors
func (sm *Manager) handleAuthError(c echo.Context, message string) error {
	path := c.Request().URL.Path

	// Check if this is a public path using the access manager
	isPublicPath := sm.isPublicPath(path)

	// Check if user has a valid session
	cookie, err := c.Cookie(sm.cookieName)
	hasValidSession := false
	if err == nil {
		if session, exists := sm.GetSession(cookie.Value); exists && time.Now().Before(session.ExpiresAt) {
			hasValidSession = true
		}
	}

	// If user is authenticated and trying to access a public path, redirect to dashboard
	if hasValidSession && isPublicPath {
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	}

	// If not authenticated and trying to access a protected path, handle accordingly
	if !hasValidSession {
		// Check if this is an API request
		isAPIRequest := strings.HasPrefix(path, "/api/")
		acceptsJSON := strings.Contains(c.Request().Header.Get("Accept"), "application/json")

		if isAPIRequest || acceptsJSON {
			// Return JSON error response for API requests
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": message,
			})
		}

		// For web requests, redirect to login
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	// If we get here, it means the user is authenticated and accessing a protected path
	// or unauthenticated and accessing a public path - both are fine
	return nil
}

// SetSessionCookie sets the session cookie
func (sm *Manager) SetSessionCookie(c echo.Context, sessionID string) {
	cookie := new(http.Cookie)
	cookie.Name = sm.cookieName
	cookie.Value = sessionID
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = sm.secureCookie
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Expires = time.Now().Add(sm.expiryTime)
	c.SetCookie(cookie)
}

// ClearSessionCookie clears the session cookie
func (sm *Manager) ClearSessionCookie(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = sm.cookieName
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = sm.secureCookie
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	c.SetCookie(cookie)
}
