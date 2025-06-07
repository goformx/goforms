package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	// SessionFile is the path to the session store file
	SessionFile = "tmp/sessions.json"
	// SessionKey is a key used in the context
	SessionKey = "session"
	// SessionCookieName is the name of the session cookie
	SessionCookieName = "session"
	sessionTimeout    = 5 * time.Second
)

// Session represents a user session
type Session struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SessionManager manages user sessions
type SessionManager struct {
	logger       logging.Logger
	sessions     map[string]*Session
	mutex        sync.RWMutex
	expiryTime   time.Duration
	storeFile    string
	secureCookie bool
}

// NewSessionManager creates a new session manager
func NewSessionManager(logger logging.Logger, secureCookie bool) *SessionManager {
	sm := &SessionManager{
		logger:       logger,
		sessions:     make(map[string]*Session),
		expiryTime:   SessionExpiryHours * time.Hour,
		storeFile:    SessionFile,
		secureCookie: secureCookie,
	}

	// Initialize session store with timeout
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Create tmp directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(SessionFile), 0755); err != nil {
			logger.Error("failed to create session directory", logging.ErrorField("error", err))
			return
		}

		// Load existing sessions
		if err := sm.loadSessions(); err != nil {
			logger.Error("failed to load sessions", logging.ErrorField("error", err))
			return
		}

		logger.Info("session store initialized successfully",
			logging.IntField("total_sessions", len(sm.sessions)),
			logging.StringField("sessions", fmt.Sprintf("%v", sm.sessions)),
		)
	}()

	// Wait for initialization with timeout
	select {
	case <-done:
		logger.Info("session store initialization completed",
			logging.IntField("total_sessions", len(sm.sessions)),
		)
	case <-time.After(sessionTimeout):
		logger.Warn("session store initialization timed out, continuing without loaded sessions")
	}

	return sm
}

// parseSessionData parses session data into a Session object
func (sm *SessionManager) parseSessionData(data map[string]any) (*Session, error) {
	createdStr, ok := data["created_at"].(string)
	if !ok {
		return nil, errors.New("invalid created_at type")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, createdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	expiresStr, ok := data["expires_at"].(string)
	if !ok {
		return nil, errors.New("invalid expires_at type")
	}

	expiresAt, err := time.Parse(time.RFC3339Nano, expiresStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expires_at: %w", err)
	}

	userID, ok := data["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user_id type")
	}

	email, ok := data["email"].(string)
	if !ok {
		return nil, errors.New("invalid email type")
	}

	role, ok := data["role"].(string)
	if !ok {
		return nil, errors.New("invalid role type")
	}

	return &Session{
		UserID:    uint(userID),
		Email:     email,
		Role:      role,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}, nil
}

// loadSessions loads sessions from the file
func (sm *SessionManager) loadSessions() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Read the file
	data, readErr := os.ReadFile(sm.storeFile)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			// File doesn't exist yet, that's okay
			return nil
		}
		return fmt.Errorf("failed to read sessions file: %w", readErr)
	}

	// Create a temporary map for unmarshaling
	tempSessions := make(map[string]map[string]any)

	// Unmarshal the data
	if unmarshalErr := json.Unmarshal(data, &tempSessions); unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal sessions: %w", unmarshalErr)
	}

	// Process each session
	validSessions := 0
	now := time.Now()
	for id, data := range tempSessions {
		session, err := sm.parseSessionData(data)
		if err != nil {
			sm.logger.Warn("failed to parse session data",
				logging.StringField("session_id", id),
				logging.ErrorField("error", err))
			continue
		}

		// Skip expired sessions
		if session.ExpiresAt.Before(now) {
			sm.logger.Debug("skipping expired session",
				logging.StringField("session_id", id),
				logging.StringField("expires_at", session.ExpiresAt.Format(time.RFC3339)))
			continue
		}

		sm.sessions[id] = session
		validSessions++
	}

	sm.logger.Info("loaded sessions from file",
		logging.IntField("total_sessions", len(tempSessions)),
		logging.IntField("valid_sessions", validSessions))
	return nil
}

// saveSessions saves sessions to the store file
func (sm *SessionManager) saveSessions() error {
	sm.logger.Debug("saveSessions: start")
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Create a map for JSON marshaling
	sessionsMap := make(map[string]map[string]any)
	for id, session := range sm.sessions {
		sessionsMap[id] = map[string]any{
			"user_id":    session.UserID,
			"email":      session.Email,
			"role":       session.Role,
			"created_at": session.CreatedAt.Format(time.RFC3339),
			"expires_at": session.ExpiresAt.Format(time.RFC3339),
		}
	}

	// Marshal sessions
	data, marshalErr := json.MarshalIndent(sessionsMap, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal sessions: %w", marshalErr)
	}

	// Write file
	if writeErr := os.WriteFile(sm.storeFile, data, 0600); writeErr != nil {
		return fmt.Errorf("failed to write session store: %w", writeErr)
	}

	sm.logger.Debug("saveSessions: end",
		logging.IntField("total_sessions", len(sm.sessions)),
		logging.StringField("sessions", fmt.Sprintf("%+v", sm.sessions)),
	)
	return nil
}

// SessionMiddleware creates a new session middleware
func (sm *SessionManager) SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sm.logger.Debug("SessionMiddleware: Processing request",
				logging.StringField("method", c.Request().Method),
				logging.StringField("path", c.Request().URL.Path),
			)

			// Skip session check for exempt paths
			if sm.isSessionExempt(c.Request().URL.Path) {
				sm.logger.Debug("SessionMiddleware: Path is exempt from session check",
					logging.StringField("path", c.Request().URL.Path),
				)
				return next(c)
			}

			// Get session cookie
			cookie, err := c.Cookie(SessionCookieName)
			if err != nil {
				sm.logger.Debug("SessionMiddleware: No session cookie found",
					logging.StringField("path", c.Request().URL.Path),
					logging.ErrorField("error", err),
				)
				return sm.handleAuthError(c, "no session found")
			}

			// Log all cookies for debugging
			sm.logger.Debug("SessionMiddleware: Cookies in request",
				logging.StringField("path", c.Request().URL.Path),
				logging.StringField("cookies", fmt.Sprintf("%+v", c.Request().Cookies())),
			)

			// Get session from manager
			session, exists := sm.GetSession(cookie.Value)
			if !exists {
				sm.logger.Debug("SessionMiddleware: No session found for cookie",
					logging.StringField("cookie_value", cookie.Value),
					logging.StringField("path", c.Request().URL.Path),
				)
				return sm.handleAuthError(c, "invalid session")
			}

			// Check if session is expired
			if time.Now().After(session.ExpiresAt) {
				sm.logger.Debug("SessionMiddleware: Session expired",
					logging.UintField("user_id", session.UserID),
					logging.StringField("email", session.Email),
					logging.StringField("expires_at", session.ExpiresAt.Format(time.RFC3339)),
				)
				sm.DeleteSession(cookie.Value)
				return sm.handleAuthError(c, "session expired")
			}

			// Log session details
			sm.logger.Debug("SessionMiddleware: Valid session found",
				logging.StringField("session_id", cookie.Value),
				logging.UintField("user_id", session.UserID),
				logging.StringField("email", session.Email),
				logging.StringField("role", session.Role),
				logging.StringField("expires_at", session.ExpiresAt.Format(time.RFC3339)),
			)

			// Store session in context
			c.Set(SessionKey, session)
			c.Set("user_id", session.UserID)
			c.Set("email", session.Email)
			c.Set("role", session.Role)

			return next(c)
		}
	}
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID uint, email, role string) (string, error) {
	sm.logger.Debug("CreateSession: start", logging.UintField("user_id", userID))
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

	// Save sessions to file
	if err := sm.saveSessions(); err != nil {
		sm.logger.Error("failed to save sessions", logging.ErrorField("error", err))
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	sm.logger.Debug("CreateSession: end",
		logging.StringField("session_id", sessionIDStr),
		logging.IntField("total_sessions", len(sm.sessions)),
	)
	return sessionIDStr, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]

	sm.logger.Debug("session lookup attempt",
		logging.StringField("session_id", sessionID),
		logging.BoolField("exists", exists),
		logging.StringField("manager_instance", fmt.Sprintf("%p", sm)),
		logging.IntField("total_sessions", len(sm.sessions)),
	)

	return session, exists
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	delete(sm.sessions, sessionID)
	sm.mutex.Unlock()

	// Save sessions to file
	if err := sm.saveSessions(); err != nil {
		sm.logger.Error("failed to save sessions", logging.ErrorField("error", err))
	}
}

// isSessionExempt checks if a path is exempt from session authentication
func (sm *SessionManager) isSessionExempt(path string) bool {
	// Check if it's a static file
	if isStaticFile(path) {
		return true
	}

	// Check for exact homepage match
	if path == "/" {
		sm.logger.Debug("SessionMiddleware: Homepage is exempt from session check",
			logging.StringField("path", path),
		)
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
			sm.logger.Debug("SessionMiddleware: Path is exempt from session check",
				logging.StringField("path", path),
				logging.StringField("exempt_path", exemptPath),
			)
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
	sm.logger.Debug(
		"Redirecting to /login from session middleware",
		logging.StringField("path", c.Request().URL.Path),
		logging.StringField("reason", message),
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
	sm.logger.Debug("SetSessionCookie: start", logging.StringField("session_id", sessionID))
	cookie := new(http.Cookie)
	cookie.Name = SessionCookieName
	cookie.Value = sessionID
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = sm.secureCookie
	cookie.SameSite = http.SameSiteLaxMode // Use Lax mode in development
	cookie.Expires = time.Now().Add(sm.expiryTime)

	sm.logger.Debug("setting session cookie",
		logging.StringField("session_id", sessionID),
		logging.StringField("cookie_path", cookie.Path),
		logging.BoolField("cookie_secure", cookie.Secure),
		logging.IntField("cookie_samesite", int(cookie.SameSite)),
		logging.StringField("cookie_expires", cookie.Expires.Format(time.RFC3339)),
	)

	c.SetCookie(cookie)
	sm.logger.Debug("SetSessionCookie: end", logging.StringField("session_id", sessionID))
}

// ClearSessionCookie clears the session cookie
func (sm *SessionManager) ClearSessionCookie(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = SessionCookieName
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = sm.secureCookie
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	c.SetCookie(cookie)
}

// Note: isPublicRoute is defined in middleware.go
