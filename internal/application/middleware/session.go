package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

const (
	// SessionExpiryHours is the number of hours before a session expires
	SessionExpiryHours = 24
	// SessionIDLength is the length of the session ID in bytes
	SessionIDLength = 32
)

// Session represents a user session
type Session struct {
	UserID    uint
	Email     string
	Role      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionManager manages user sessions
type SessionManager struct {
	logger     logging.Logger
	sessions   map[string]*Session
	mutex      sync.RWMutex
	expiryTime time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(logger logging.Logger) *SessionManager {
	return &SessionManager{
		logger:     logger,
		sessions:   make(map[string]*Session),
		expiryTime: SessionExpiryHours * time.Hour, // Sessions expire after 24 hours
	}
}

// SessionMiddleware creates a new session middleware
func (sm *SessionManager) SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip session check for certain paths
			if sm.isSessionExempt(c.Request().URL.Path) {
				return next(c)
			}

			// Get session ID from cookie
			cookie, err := c.Cookie("session_id")
			if err != nil {
				return sm.handleAuthError(c, "no session found")
			}

			// Get session from manager
			session, exists := sm.GetSession(cookie.Value)
			if !exists {
				return sm.handleAuthError(c, "invalid session")
			}

			// Check if session is expired
			if time.Now().After(session.ExpiresAt) {
				sm.DeleteSession(cookie.Value)
				return sm.handleAuthError(c, "session expired")
			}

			// Store session in context
			c.Set("session", session)
			return next(c)
		}
	}
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID uint, email, role string) (string, error) {
	// Generate random session ID
	sessionID := make([]byte, SessionIDLength)
	if _, err := rand.Read(sessionID); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	sessionIDStr := base64.URLEncoding.EncodeToString(sessionID)

	// Create session
	session := &Session{
		UserID:    userID,
		Email:     email,
		Role:      role,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sm.expiryTime),
	}

	// Store session
	sm.mutex.Lock()
	sm.sessions[sessionIDStr] = session
	sm.mutex.Unlock()

	return sessionIDStr, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	session, exists := sm.sessions[sessionID]
	return session, exists
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	delete(sm.sessions, sessionID)
	sm.mutex.Unlock()
}

// isSessionExempt checks if a path is exempt from session authentication
func (sm *SessionManager) isSessionExempt(path string) bool {
	// Check if it's a static file
	if isStaticFile(path) {
		return true
	}

	// Check other exempt paths
	exemptPaths := []string{
		"/api/validation/",
		"/login",
		"/signup",
		"/forgot-password",
		"/contact",
		"/demo",
	}

	for _, exemptPath := range exemptPaths {
		if path == exemptPath || len(path) > len(exemptPath) && path[:len(exemptPath)] == exemptPath {
			return true
		}
	}
	return false
}

// handleAuthError handles authentication errors
func (sm *SessionManager) handleAuthError(c echo.Context, message string) error {
	sm.logger.Debug("authentication error",
		logging.StringField("message", message),
		logging.StringField("path", c.Request().URL.Path),
	)

	// For API requests, return 401
	if c.Request().Header.Get("Accept") == "application/json" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": message,
		})
	}

	// For web requests, redirect to login
	return c.Redirect(http.StatusSeeOther, "/login")
}

// SetSessionCookie sets the session cookie
func (sm *SessionManager) SetSessionCookie(c echo.Context, sessionID string) {
	cookie := new(http.Cookie)
	cookie.Name = "session_id"
	cookie.Value = sessionID
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Expires = time.Now().Add(sm.expiryTime)
	c.SetCookie(cookie)
}

// ClearSessionCookie clears the session cookie
func (sm *SessionManager) ClearSessionCookie(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = "session_id"
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	c.SetCookie(cookie)
}

// Note: isPublicRoute is defined in middleware.go
