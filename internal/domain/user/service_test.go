package user_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonesrussell/goforms/internal/domain/user"
	mocklogging "github.com/jonesrussell/goforms/test/mocks/logging"
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
func (s *MockStore) Create(u *user.User) error {
	s.users[u.ID] = u
	s.email[u.Email] = u
	return nil
}

// GetByID implements Store.GetByID
func (s *MockStore) GetByID(id uint) (*user.User, error) {
	if u, ok := s.users[id]; ok {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

// GetByEmail implements Store.GetByEmail
func (s *MockStore) GetByEmail(email string) (*user.User, error) {
	if u, ok := s.email[email]; ok {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

// Update implements Store.Update
func (s *MockStore) Update(u *user.User) error {
	if _, ok := s.users[u.ID]; !ok {
		return user.ErrUserNotFound
	}
	s.users[u.ID] = u
	s.email[u.Email] = u
	return nil
}

// Delete implements Store.Delete
func (s *MockStore) Delete(id uint) error {
	if u, ok := s.users[id]; ok {
		delete(s.users, id)
		delete(s.email, u.Email)
		return nil
	}
	return user.ErrUserNotFound
}

// List implements Store.List
func (s *MockStore) List() ([]user.User, error) {
	users := make([]user.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, *u)
	}
	return users, nil
}

func TestUserService(t *testing.T) {
	ctx := t.Context()

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()

	testDir := filepath.Join(os.TempDir(), "goforms-test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	t.Chdir(testDir)

	store := NewMockStore()
	logger := mocklogging.NewMockLogger()
	service := user.NewService(store, logger)

	t.Run("SignUp", func(t *testing.T) {
		signup := &user.Signup{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		logger.ExpectInfo("user signed up").WithFields(map[string]interface{}{
			"email": signup.Email,
		})

		newUser, err := service.SignUp(ctx, signup)
		require.NoError(t, err)
		assert.NotNil(t, newUser)
		assert.Equal(t, signup.Email, newUser.Email)
		assert.Equal(t, signup.FirstName, newUser.FirstName)
		assert.Equal(t, signup.LastName, newUser.LastName)
		require.NoError(t, logger.Verify())
	})

	t.Run("Login", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		logger.ExpectInfo("user logged in").WithFields(map[string]interface{}{
			"email": login.Email,
		})

		tokens, err := service.Login(ctx, login)
		require.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
		require.NoError(t, logger.Verify())
	})

	t.Run("Logout", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		logger.ExpectInfo("user logged in").WithFields(map[string]interface{}{
			"email": login.Email,
		})

		tokens, err := service.Login(ctx, login)
		require.NoError(t, err)

		logger.ExpectInfo("user logged out").WithFields(map[string]interface{}{
			"email": login.Email,
		})

		err = service.Logout(ctx, tokens.AccessToken)
		require.NoError(t, err)

		assert.True(t, service.IsTokenBlacklisted(tokens.AccessToken))
		require.NoError(t, logger.Verify())
	})

	t.Run("RefreshToken", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		logger.ExpectInfo("user logged in").WithFields(map[string]interface{}{
			"email": login.Email,
		})

		tokens, err := service.Login(ctx, login)
		require.NoError(t, err)

		logger.ExpectInfo("token refreshed").WithFields(map[string]interface{}{
			"email": login.Email,
		})

		newTokens, err := service.RefreshToken(ctx, tokens.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)
		assert.NotEqual(t, tokens.AccessToken, newTokens.AccessToken)
		require.NoError(t, logger.Verify())
	})
}
