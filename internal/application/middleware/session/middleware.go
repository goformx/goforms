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
			// Skip session check for exempt paths
			if sm.isPathExempt(c.Request().URL.Path) {
				return next(c)
			}

			// Get session cookie
			cookie, err := c.Cookie(sm.cookieName)
			if err != nil {
				sm.logger.Debug("SessionMiddleware: No session cookie found", "path", c.Request().URL.Path)
				return sm.handleAuthError(c, "no session found")
			}
			sm.logger.Debug("SessionMiddleware: Found session cookie", "cookie", cookie.Value, "path", c.Request().URL.Path)

			// Get session from manager
			session, exists := sm.GetSession(cookie.Value)
			if !exists {
				sm.logger.Debug("SessionMiddleware: Session not found", "cookie", cookie.Value, "path", c.Request().URL.Path)
				return sm.handleAuthError(c, "invalid session")
			}
			sm.logger.Debug("SessionMiddleware: Session found", "user_id", session.UserID, "path", c.Request().URL.Path)

			// Check if session is expired
			if time.Now().After(session.ExpiresAt) {
				sm.logger.Debug("SessionMiddleware: Session expired", "user_id", session.UserID, "path", c.Request().URL.Path)
				sm.DeleteSession(cookie.Value)
				return sm.handleAuthError(c, "session expired")
			}

			// Store session in context
			sm.logger.Debug("SessionMiddleware: Setting session in context",
				"user_id", session.UserID,
				"path", c.Request().URL.Path,
			)
			c.Set(string(context.SessionKey), session)
			context.SetUserID(c, session.UserID)
			context.SetEmail(c, session.Email)
			context.SetRole(c, session.Role)

			return next(c)
		}
	}
}

// isPathExempt checks if a path is exempt from session authentication
func (sm *Manager) isPathExempt(path string) bool {
	// Use accessManager to check if the path is public
	if sm.accessManager != nil {
		if sm.accessManager.GetRequiredAccess(path, "GET") == access.PublicAccess {
			return true
		}
	}
	// Skip authentication for:
	// 1. Static assets (files, images, etc.)
	// 2. Public API endpoints
	// 3. Health checks and monitoring
	// 4. Development tools and debugging endpoints
	// 5. Public pages (login, signup, etc.)

	// Check if it's a static file or asset
	if sm.isStaticFile(path) {
		return true
	}

	// Check if it's a public API endpoint
	if strings.HasPrefix(path, "/api/v1/public/") {
		return true
	}

	// Check if it's a validation endpoint
	if strings.HasPrefix(path, "/api/v1/validation/") {
		return true
	}

	// Check if it's a health check or monitoring endpoint
	if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/metrics") {
		return true
	}

	// Check if it's a development tool endpoint
	if sm.config.Config.App.Env == "development" && (strings.HasPrefix(path, "/.well-known/") ||
		strings.HasPrefix(path, "/debug/") ||
		strings.HasPrefix(path, "/dev/")) {
		return true
	}

	// Check public paths (login, signup, etc.)
	for _, publicPath := range sm.config.PublicPaths {
		if path == publicPath {
			return true
		}
	}

	// Check exempt paths
	for _, exemptPath := range sm.config.ExemptPaths {
		if strings.HasPrefix(path, exemptPath) {
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

	// Check if this is a public path
	isPublicPath := false
	for _, publicPath := range sm.config.PublicPaths {
		if path == publicPath {
			isPublicPath = true
			break
		}
	}

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

// SessionMiddleware creates a new session middleware
func (sm *Manager) SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Skip session check for exempt paths
			if sm.isPathExempt(path) {
				return next(c)
			}

			// Get session cookie
			cookie, err := c.Cookie(sm.cookieName)
			if err != nil {
				return sm.handleAuthError(c, "no session found")
			}

			// Get session from manager
			session, exists := sm.GetSession(cookie.Value)
			if !exists {
				return sm.handleAuthError(c, "invalid session")
			}

			// Check session expiration
			if session.ExpiresAt.Before(time.Now()) {
				return sm.handleAuthError(c, "session expired")
			}

			// Set user ID in context
			c.Set("user_id", session.UserID)
			c.Set("user_email", session.Email)

			// Refresh session if needed
			if time.Until(session.ExpiresAt) < sm.expiryTime/2 {
				// Create a new session with the same user data
				newSessionID, createErr := sm.CreateSession(session.UserID, session.Email, session.Role)
				if createErr != nil {
					sm.logger.Error("failed to create new session", "error", createErr)
					return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
				}
				// Delete old session
				sm.DeleteSession(cookie.Value)
				// Set new session cookie
				sm.SetSessionCookie(c, newSessionID)
			}

			return next(c)
		}
	}
}
