package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// FileStorage implements SessionStorage using file system
type FileStorage struct {
	storeFile string
	logger    logging.Logger
}

// NewFileStorage creates a new file-based session storage
func NewFileStorage(storeFile string, logger logging.Logger) *FileStorage {
	return &FileStorage{
		storeFile: storeFile,
		logger:    logger,
	}
}

// Load implements SessionStorage.Load
func (fs *FileStorage) Load() (map[string]*Session, error) {
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
func (fs *FileStorage) Save(sessions map[string]*Session) error {
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
func (fs *FileStorage) Delete(sessionID string) error {
	sessions, err := fs.Load()
	if err != nil {
		return err
	}

	delete(sessions, sessionID)
	return fs.Save(sessions)
}

// parseSessionData parses session data into a Session object
func (fs *FileStorage) parseSessionData(data map[string]any) (*Session, error) {
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
