package user_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStore implements the Store interface for testing
type MockStore struct {
	users map[uint]*user.User
	email map[string]*user.User
}

// NewMockStore creates a new mock store
func NewMockStore() *MockStore {
	return &MockStore{
		users: make(map[uint]*user.User),
		email: make(map[string]*user.User),
	}
}

// Create implements Store.Create
func (s *MockStore) Create(user *user.User) error {
	s.users[user.ID] = user
	s.email[user.Email] = user
	return nil
}

// GetByID implements Store.GetByID
func (s *MockStore) GetByID(id uint) (*user.User, error) {
	if user, ok := s.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

// GetByEmail implements Store.GetByEmail
func (s *MockStore) GetByEmail(email string) (*user.User, error) {
	if user, ok := s.email[email]; ok {
		return user, nil
	}
	return nil, nil
}

// Update implements Store.Update
func (s *MockStore) Update(user *user.User) error {
	if _, ok := s.users[user.ID]; !ok {
		return nil
	}
	s.users[user.ID] = user
	s.email[user.Email] = user
	return nil
}

// Delete implements Store.Delete
func (s *MockStore) Delete(id uint) error {
	if user, ok := s.users[id]; ok {
		delete(s.users, id)
		delete(s.email, user.Email)
	}
	return nil
}

// List implements Store.List
func (s *MockStore) List() ([]user.User, error) {
	users := make([]user.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, *user)
	}
	return users, nil
}

// MockLogger implements the logging.Logger interface for testing
type MockLogger struct{}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (l *MockLogger) Debug(msg string, fields ...logging.Field)  {}
func (l *MockLogger) Info(msg string, fields ...logging.Field)   {}
func (l *MockLogger) Warn(msg string, fields ...logging.Field)   {}
func (l *MockLogger) Error(msg string, fields ...logging.Field)  {}
func (l *MockLogger) Int(key string, val int) logging.Field     { return logging.Field{} }
func (l *MockLogger) Int32(key string, val int32) logging.Field { return logging.Field{} }
func (l *MockLogger) Int64(key string, val int64) logging.Field { return logging.Field{} }
func (l *MockLogger) Uint(key string, val uint) logging.Field   { return logging.Field{} }
func (l *MockLogger) Uint32(key string, val uint32) logging.Field { return logging.Field{} }
func (l *MockLogger) Uint64(key string, val uint64) logging.Field { return logging.Field{} }

func TestUserService(t *testing.T) {
	// Use new T.Context() for test context
	ctx := t.Context()
	
	// Use new T.Chdir() for test directory management
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()
	
	// Create a temporary test directory
	testDir := filepath.Join(os.TempDir(), "goforms-test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	t.Chdir(testDir)
	
	// Test setup
	store := NewMockStore()
	logger := NewMockLogger()
	service := user.NewService(store, logger)

	t.Run("SignUp", func(t *testing.T) {
		signup := &user.Signup{
			Email:    "test@example.com",
			Password: "password123",
		}

		user, err := service.SignUp(ctx, signup)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, signup.Email, user.Email)
	})

	t.Run("Login", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		tokens, err := service.Login(ctx, login)
		require.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("Logout", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		tokens, err := service.Login(ctx, login)
		require.NoError(t, err)

		err = service.Logout(ctx, tokens.AccessToken)
		require.NoError(t, err)

		// Verify token is blacklisted
		assert.True(t, service.IsTokenBlacklisted(tokens.AccessToken))
	})

	t.Run("RefreshToken", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		tokens, err := service.Login(ctx, login)
		require.NoError(t, err)

		newTokens, err := service.RefreshToken(ctx, tokens.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)
		assert.NotEqual(t, tokens.AccessToken, newTokens.AccessToken)
	})
}
