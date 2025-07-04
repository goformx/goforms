// Package session provides session management middleware and utilities for the application.
package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// NewManager creates a new session manager
func NewManager(
	logger logging.Logger,
	cfg *Config,
	lc fx.Lifecycle,
	accessManager *access.Manager,
) *Manager {
	// Create tmp directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(cfg.StoreFile), 0o750); err != nil {
		logger.Error("failed to create session directory", "error", err)
	}

	storage := NewFileStorage(cfg.StoreFile, logger)

	sm := &Manager{
		logger:        logger,
		storage:       storage,
		sessions:      make(map[string]*Session),
		expiryTime:    cfg.MaxAge,
		secureCookie:  cfg.Secure,
		cookieName:    cfg.CookieName,
		stopChan:      make(chan struct{}),
		config:        cfg,
		accessManager: accessManager,
	}

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			// Initialize session store
			if err := sm.initialize(); err != nil {
				return fmt.Errorf("failed to initialize session store: %w", err)
			}

			// Start cleanup routine
			go sm.cleanupRoutine()

			return nil
		},
		OnStop: func(_ context.Context) error {
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
func (sm *Manager) initialize() error {
	// Load existing sessions
	sessions, err := sm.storage.Load()
	if err != nil {
		sm.logger.Error("failed to load sessions", "error", err)

		return fmt.Errorf("failed to load sessions: %w", err)
	}

	sm.mutex.Lock()
	sm.sessions = sessions
	sm.mutex.Unlock()

	return nil
}

// cleanupRoutine periodically cleans up expired sessions
func (sm *Manager) cleanupRoutine() {
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
func (sm *Manager) cleanupExpiredSessions() {
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

// saveSessions saves sessions to the store
func (sm *Manager) saveSessions() error {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if err := sm.storage.Save(sm.sessions); err != nil {
		return fmt.Errorf("save sessions to storage: %w", err)
	}

	return nil
}

// CreateSession creates a new session for a user
func (sm *Manager) CreateSession(userID, email, role string) (string, error) {
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
func (sm *Manager) GetSession(sessionID string) (*Session, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.sessions[sessionID], sm.sessions[sessionID] != nil
}

// DeleteSession removes a session
func (sm *Manager) DeleteSession(sessionID string) {
	sm.mutex.Lock()
	delete(sm.sessions, sessionID)
	sm.mutex.Unlock()

	// Save sessions to file
	if err := sm.saveSessions(); err != nil {
		sm.logger.Error("failed to save sessions", "error", err)
	}
}

// GetCookieName returns the name of the session cookie
func (sm *Manager) GetCookieName() string {
	return sm.cookieName
}
