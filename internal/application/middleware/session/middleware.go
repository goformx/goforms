package session

import (
	"net/http"
	"strings"
	"time"

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
			sm.logger.Debug("SessionMiddleware: Setting session in context", "user_id", session.UserID, "path", c.Request().URL.Path)
			c.Set(SessionKey, session)
			c.Set("user_id", session.UserID)
			c.Set("email", session.Email)
			c.Set("role", session.Role)

			return next(c)
		}
	}
}

// isPathExempt checks if a path is exempt from session authentication
func (sm *Manager) isPathExempt(path string) bool {
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

// isStaticFile checks if a path is a static file
func (sm *Manager) isStaticFile(path string) bool {
	// Check for common static file extensions
	staticExtensions := []string{
		".ico", ".png", ".jpg", ".jpeg", ".gif", ".svg",
		".css", ".js", ".woff", ".woff2", ".ttf", ".eot",
		".map", ".json", ".txt", ".xml", ".pdf",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	// Check static paths
	for _, staticPath := range sm.config.StaticPaths {
		if strings.HasPrefix(path, staticPath) {
			return true
		}
	}

	return false
}

// handleAuthError handles authentication errors
func (sm *Manager) handleAuthError(c echo.Context, message string) error {
	// Special case for homepage - if authenticated, redirect to dashboard
	if c.Request().URL.Path == "/" {
		// Check if user has a valid session
		if cookie, err := c.Cookie(sm.cookieName); err == nil {
			if session, exists := sm.GetSession(cookie.Value); exists && time.Now().Before(session.ExpiresAt) {
				return c.Redirect(http.StatusSeeOther, "/dashboard")
			}
		}
		// If not authenticated, allow access to homepage
		return nil
	}

	// Check if this is an API request
	isAPIRequest := strings.HasPrefix(c.Request().URL.Path, "/api/")
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
				// Create new session with same user data
				newSessionID, err := sm.CreateSession(session.UserID, session.Email, session.Role)
				if err != nil {
					sm.logger.Error("failed to refresh session", "error", err)
				} else {
					// Delete old session
					sm.DeleteSession(cookie.Value)
					// Set new session cookie
					sm.SetSessionCookie(c, newSessionID)
				}
			}

			return next(c)
		}
	}
}
