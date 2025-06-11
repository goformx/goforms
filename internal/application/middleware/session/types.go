package session

import (
	"sync"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

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

// SessionStorage defines the interface for session storage operations
type SessionStorage interface {
	Load() (map[string]*Session, error)
	Save(sessions map[string]*Session) error
	Delete(sessionID string) error
}

// SessionConfig extends the base config with additional session-specific settings
type SessionConfig struct {
	*config.SessionConfig
	*config.Config
	PublicPaths  []string `json:"public_paths"`
	ExemptPaths  []string `json:"exempt_paths"`
	StaticPaths  []string `json:"static_paths"`
	ErrorHandler func(c echo.Context, message string) error
}

// Manager manages user sessions
type Manager struct {
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
