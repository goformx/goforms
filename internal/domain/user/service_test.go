package user_test

import (
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
	mockStore := NewMockStore()
	mockLogger := mocklogging.NewMockLogger()
	service := user.NewService(mockStore, mockLogger)

	t.Run("signup and login flow", func(t *testing.T) {
		// Create signup request
		signup := &user.Signup{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		// Mock store expectations
		mockStore.Create(&user.User{
			Email:     signup.Email,
			FirstName: signup.FirstName,
			LastName:  signup.LastName,
		})

		newUser, signupErr := service.SignUp(ctx, signup)
		require.NoError(t, signupErr)
		require.NotNil(t, newUser)

		// Create login request
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		// Mock store expectations
		mockStore.GetByEmail(login.Email)

		// Test successful login
		authUser, loginErr := service.Login(ctx, login)
		require.NoError(t, loginErr)
		require.NotNil(t, authUser)

		// Test invalid password
		invalidLogin := &user.Login{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		// Mock store expectations
		mockStore.GetByEmail(invalidLogin.Email)

		_, invalidLoginErr := service.Login(ctx, invalidLogin)
		require.Error(t, invalidLoginErr)

		// Test non-existent user
		nonExistentLogin := &user.Login{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		// Mock store expectations
		mockStore.GetByEmail(nonExistentLogin.Email)

		_, nonExistentErr := service.Login(ctx, nonExistentLogin)
		require.Error(t, nonExistentErr)
	})

	t.Run("Logout", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		mockLogger.ExpectInfo("user logged in").WithFields(map[string]any{
			"email": login.Email,
		})

		tokens, err := service.Login(ctx, login)
		require.NoError(t, err)

		mockLogger.ExpectInfo("user logged out").WithFields(map[string]any{
			"email": login.Email,
		})

		err = service.Logout(ctx, tokens.AccessToken)
		require.NoError(t, err)

		assert.True(t, service.IsTokenBlacklisted(tokens.AccessToken))
		require.NoError(t, mockLogger.Verify())
	})

	t.Run("RefreshToken", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		mockLogger.ExpectInfo("user logged in").WithFields(map[string]any{
			"email": login.Email,
		})

		tokens, err := service.Login(ctx, login)
		require.NoError(t, err)

		mockLogger.ExpectInfo("token refreshed").WithFields(map[string]any{
			"email": login.Email,
		})

		newTokens, err := service.RefreshToken(ctx, tokens.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newTokens.AccessToken)
		assert.NotEmpty(t, newTokens.RefreshToken)
		assert.NotEqual(t, tokens.AccessToken, newTokens.AccessToken)
		require.NoError(t, mockLogger.Verify())
	})
}
