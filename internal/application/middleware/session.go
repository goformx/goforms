package middleware

import (
	"context"
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

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// SessionStorage defines the interface for session storage operations
type SessionStorage interface {
	Load() (map[string]*Session, error)
	Save(sessions map[string]*Session) error
	Delete(sessionID string) error
}

// FileSessionStorage implements SessionStorage using file system
type FileSessionStorage struct {
	storeFile string
	logger    logging.Logger
}

// SessionConfig extends the base config with additional session-specific settings
type SessionConfig struct {
	*config.SessionConfig
	PublicPaths  []string `json:"public_paths"`
	ExemptPaths  []string `json:"exempt_paths"`
	StaticPaths  []string `json:"static_paths"`
	ErrorHandler func(c echo.Context, message string) error
}

const (
	// SessionExpiryHours is the number of hours before a session expires
	SessionExpiryHours = 24
	// SessionIDLength is the length of the session ID in bytes
	SessionIDLength = 32
	// SessionKey is a key used in the context
	SessionKey     = "session"
	sessionTimeout = 5 * time.Second
	// cleanupInterval is how often to run session cleanup
	cleanupInterval = 1 * time.Hour
)

// Session represents a user session
type Session struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SessionManager manages user sessions
type SessionManager struct {
	logger       logging.Logger
	storage      SessionStorage
	sessions     map[string]*Session
	mutex        sync.RWMutex
	expiryTime   time.Duration
	secureCookie bool
	cookieName   string
	stopChan     chan struct{}
	config       *SessionConfig
}

// NewSessionManager creates a new session manager
func NewSessionManager(logger logging.Logger, cfg *SessionConfig, lc fx.Lifecycle) *SessionManager {
	// Create tmp directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(cfg.StoreFile), 0o755); err != nil {
		logger.Error("failed to create session directory", "error", err)
	}

	storage := NewFileSessionStorage(cfg.StoreFile, logger)

	sm := &SessionManager{
		logger:       logger,
		storage:      storage,
		sessions:     make(map[string]*Session),
		expiryTime:   cfg.TTL,
		secureCookie: cfg.Secure,
		cookieName:   cfg.CookieName,
		stopChan:     make(chan struct{}),
		config:       cfg,
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Initialize session store
			if err := sm.initialize(); err != nil {
				return fmt.Errorf("failed to initialize session store: %w", err)
			}

			// Start cleanup routine
			go sm.cleanupRoutine()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			// Stop cleanup routine
			close(sm.stopChan)

			// Save sessions before shutdown
			if err := sm.saveSessions(); err != nil {
				sm.logger.Error("failed to save sessions during shutdown", "error", err)
			}

			return nil
		},
	})

	return sm
}

// initialize sets up the session store
func (sm *SessionManager) initialize() error {
	// Load existing sessions
	sessions, err := sm.storage.Load()
	if err != nil {
		sm.logger.Error("failed to load sessions", "error", err)
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	sm.mutex.Lock()
	sm.sessions = sessions
	sm.mutex.Unlock()

	sm.logger.Info("session store initialized", "total_sessions", len(sm.sessions))
	return nil
}

// cleanupRoutine periodically cleans up expired sessions
func (sm *SessionManager) cleanupRoutine() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.cleanupExpiredSessions()
		case <-sm.stopChan:
			return
		}
	}
}

// cleanupExpiredSessions removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for id, session := range sm.sessions {
		if session.ExpiresAt.Before(now) {
			delete(sm.sessions, id)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		sm.logger.Info("cleaned up expired sessions", "count", expiredCount, "remaining", len(sm.sessions))

		// Save sessions after cleanup
		if err := sm.saveSessions(); err != nil {
			sm.logger.Error("failed to save sessions after cleanup", "error", err)
		}
	}
}

// saveSessions saves sessions to the store file
func (sm *SessionManager) saveSessions() error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.storage.Save(sm.sessions)
}

// SessionMiddleware creates a new session middleware
func (sm *SessionManager) SessionMiddleware() echo.MiddlewareFunc {
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

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID, email, role string) (string, error) {
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
		sm.logger.Error("failed to save sessions", "error", err)
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	return sessionIDStr, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*Session, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.sessions[sessionID], sm.sessions[sessionID] != nil
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	delete(sm.sessions, sessionID)
	sm.mutex.Unlock()

	// Save sessions to file
	if err := sm.saveSessions(); err != nil {
		sm.logger.Error("failed to save sessions", "error", err)
	}
}

// isPathExempt checks if a path is exempt from session authentication
func (sm *SessionManager) isPathExempt(path string) bool {
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
func (sm *SessionManager) isStaticFile(path string) bool {
	for _, staticPath := range sm.config.StaticPaths {
		if len(path) > len(staticPath) && path[:len(staticPath)] == staticPath {
			return true
		}
	}
	return false
}

// handleAuthError handles authentication errors
func (sm *SessionManager) handleAuthError(c echo.Context, message string) error {
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
func (sm *SessionManager) SetSessionCookie(c echo.Context, sessionID string) {
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
func (sm *SessionManager) ClearSessionCookie(c echo.Context) {
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

// GetCookieName returns the name of the session cookie
func (sm *SessionManager) GetCookieName() string {
	return sm.cookieName
}

// NewFileSessionStorage creates a new file-based session storage
func NewFileSessionStorage(storeFile string, logger logging.Logger) *FileSessionStorage {
	return &FileSessionStorage{
		storeFile: storeFile,
		logger:    logger,
	}
}

// Load implements SessionStorage.Load
func (fs *FileSessionStorage) Load() (map[string]*Session, error) {
	data, err := os.ReadFile(fs.storeFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*Session), nil
		}
		return nil, fmt.Errorf("failed to read sessions file: %w", err)
	}

	tempSessions := make(map[string]map[string]any)
	if err := json.Unmarshal(data, &tempSessions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sessions: %w", err)
	}

	sessions := make(map[string]*Session)
	now := time.Now()
	for id, data := range tempSessions {
		session, err := fs.parseSessionData(data)
		if err != nil {
			fs.logger.Warn("failed to parse session data", "session_id", id, "error", err)
			continue
		}

		if session.ExpiresAt.Before(now) {
			continue
		}

		sessions[id] = session
	}

	return sessions, nil
}

// Save implements SessionStorage.Save
func (fs *FileSessionStorage) Save(sessions map[string]*Session) error {
	sessionsMap := make(map[string]map[string]any)
	for id, session := range sessions {
		sessionsMap[id] = map[string]any{
			"user_id":    session.UserID,
			"email":      session.Email,
			"role":       session.Role,
			"created_at": session.CreatedAt.Format(time.RFC3339),
			"expires_at": session.ExpiresAt.Format(time.RFC3339),
		}
	}

	data, err := json.MarshalIndent(sessionsMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	if err := os.WriteFile(fs.storeFile, data, 0o600); err != nil {
		return fmt.Errorf("failed to write session store: %w", err)
	}

	return nil
}

// Delete implements SessionStorage.Delete
func (fs *FileSessionStorage) Delete(sessionID string) error {
	sessions, err := fs.Load()
	if err != nil {
		return err
	}

	delete(sessions, sessionID)
	return fs.Save(sessions)
}

// parseSessionData parses session data into a Session object
func (fs *FileSessionStorage) parseSessionData(data map[string]any) (*Session, error) {
	createdAt, ok := data["created_at"].(string)
	if !ok {
		return nil, errors.New("invalid created_at type")
	}

	expiresAt, ok := data["expires_at"].(string)
	if !ok {
		return nil, errors.New("invalid expires_at type")
	}

	userID, ok := data["user_id"].(string)
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

	createdAtTime, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at format: %w", err)
	}

	expiresAtTime, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at format: %w", err)
	}

	return &Session{
		UserID:    userID,
		Email:     email,
		Role:      role,
		CreatedAt: createdAtTime,
		ExpiresAt: expiresAtTime,
	}, nil
}
