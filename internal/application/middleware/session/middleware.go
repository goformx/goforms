package session

import (
	"net/http"
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
				path := c.Request().URL.Path
				if path == "/" || path == "/login" || path == "/signup" {
					// Let the handler render the page
					return next(c)
				}
				return sm.handleAuthError(c, "no session found")
			}
			sm.logger.Debug("SessionMiddleware: Found session cookie", "cookie", cookie.Value, "path", c.Request().URL.Path)

			// Get session from manager
			session, exists := sm.GetSession(cookie.Value)
			if !exists {
				sm.logger.Debug("SessionMiddleware: Session not found", "cookie", cookie.Value, "path", c.Request().URL.Path)
				path := c.Request().URL.Path
				if path == "/" || path == "/login" || path == "/signup" {
					// Let the handler render the page
					return next(c)
				}
				return sm.handleAuthError(c, "invalid session")
			}
			sm.logger.Debug("SessionMiddleware: Session found", "user_id", session.UserID, "path", c.Request().URL.Path)

			// Check if session is expired
			if time.Now().After(session.ExpiresAt) {
				sm.logger.Debug("SessionMiddleware: Session expired", "user_id", session.UserID, "path", c.Request().URL.Path)
				sm.DeleteSession(cookie.Value)
				path := c.Request().URL.Path
				if path == "/" || path == "/login" || path == "/signup" {
					// Let the handler render the page
					return next(c)
				}
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
	// Check if it's a static file
	if sm.isStaticFile(path) {
		return true
	}

	// Check public paths
	for _, publicPath := range sm.config.PublicPaths {
		if path == publicPath {
			return true
		}
	}

	// Check exempt paths
	for _, exemptPath := range sm.config.ExemptPaths {
		if path == exemptPath || len(path) > len(exemptPath) && path[:len(exemptPath)] == exemptPath {
			sm.logger.Debug("SessionMiddleware: Path is exempt from session check",
				"path", path,
				"exempt_path", exemptPath,
			)
			return true
		}
	}
	return false
}

// isStaticFile checks if a path is a static file
func (sm *Manager) isStaticFile(path string) bool {
	for _, staticPath := range sm.config.StaticPaths {
		if len(path) > len(staticPath) && path[:len(staticPath)] == staticPath {
			return true
		}
	}
	return false
}

// handleAuthError handles authentication errors
func (sm *Manager) handleAuthError(c echo.Context, message string) error {
	if sm.config.ErrorHandler != nil {
		return sm.config.ErrorHandler(c, message)
	}

	path := c.Request().URL.Path

	// For API requests, return 401
	if c.Request().Header.Get("Accept") == "application/json" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": message,
		})
	}

	// If already on /login or /signup, just render the page, don't redirect
	if sm.isPathExempt(path) {
		return nil // Let the handler render the page
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
