package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
	mocklogging "github.com/goformx/goforms/test/mocks/logging"
	mockuser "github.com/goformx/goforms/test/mocks/user"
)

func TestService_SignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockuser.NewMockRepository(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	svc := user.NewService(repo, logger)

	t.Run("successful signup", func(t *testing.T) {
		signup := &user.Signup{
			Email:           "test@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
		}

		// Mock repository calls - use common.ErrNotFound to indicate user doesn't exist
		repo.EXPECT().GetByEmail(gomock.Any(), signup.Email).Return(nil, common.ErrNotFound)
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *entities.User) error {
			// Set timestamps to simulate DB behavior
			u.CreatedAt = time.Now()
			u.UpdatedAt = time.Now()
			assert.Equal(t, signup.Email, u.Email)
			assert.True(t, u.CheckPassword(signup.Password))
			assert.NotEmpty(t, u.ID)
			assert.False(t, u.CreatedAt.IsZero())
			assert.False(t, u.UpdatedAt.IsZero())

			return nil
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.SignUp(ctx, signup)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, signup.Email, result.Email)
		assert.True(t, result.CheckPassword(signup.Password))
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())
	})

	t.Run("email already exists", func(t *testing.T) {
		signup := &user.Signup{
			Email:           "existing@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
		}

		existingUser := &entities.User{
			ID:    "existing-user-id",
			Email: signup.Email,
		}

		// Mock repository to return existing user
		repo.EXPECT().GetByEmail(gomock.Any(), signup.Email).Return(existingUser, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.SignUp(ctx, signup)
		require.Error(t, err)
		require.Nil(t, result)
		assert.ErrorIs(t, err, user.ErrUserExists)
	})

	t.Run("database error during email check", func(t *testing.T) {
		signup := &user.Signup{
			Email:           "test@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
		}

		// Mock repository to return a real error (not "not found")
		repo.EXPECT().GetByEmail(gomock.Any(), signup.Email).Return(nil, errors.New("database connection failed"))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.SignUp(ctx, signup)
		require.Error(t, err)
		require.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check existing user")
	})

	t.Run("database error during user creation", func(t *testing.T) {
		signup := &user.Signup{
			Email:           "test@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
		}

		// Mock repository calls - user not found (good), then creation fails
		repo.EXPECT().GetByEmail(gomock.Any(), signup.Email).Return(nil, common.ErrNotFound)
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("database constraint violation"))
		logger.EXPECT().Error(gomock.Any(), gomock.Any())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.SignUp(ctx, signup)
		require.Error(t, err)
		require.Nil(t, result)
		assert.Contains(t, err.Error(), "create:")
	})

	t.Run("invalid email format", func(t *testing.T) {
		signup := &user.Signup{
			Email:           "invalid-email",
			Password:        "password123",
			ConfirmPassword: "password123",
		}

		// Mock repository calls - user not found, but email has no @ symbol
		repo.EXPECT().GetByEmail(gomock.Any(), signup.Email).Return(nil, common.ErrNotFound)
		// Create should NOT be called because email validation fails first

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.SignUp(ctx, signup)
		require.Error(t, err)
		require.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("password too short", func(t *testing.T) {
		signup := &user.Signup{
			Email:           "test@example.com",
			Password:        "short",
			ConfirmPassword: "short",
		}

		// Mock repository calls - user not found
		repo.EXPECT().GetByEmail(gomock.Any(), signup.Email).Return(nil, common.ErrNotFound)
		// Create will be called but NewUser will fail due to password validation
		logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.SignUp(ctx, signup)
		require.Error(t, err)
		require.Nil(t, result)
		assert.Contains(t, err.Error(), "password")
	})
}

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockuser.NewMockRepository(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	svc := user.NewService(repo, logger)

	t.Run("successful login", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		// Create a test user with known password
		testUser, err := entities.NewUser(login.Email, login.Password, "Test", "User")
		require.NoError(t, err)

		// Mock repository call
		repo.EXPECT().GetByEmail(gomock.Any(), login.Email).Return(testUser, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.Login(ctx, login)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testUser, result.User)
	})

	t.Run("user not found", func(t *testing.T) {
		login := &user.Login{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		// Mock repository call - use common.ErrNotFound
		repo.EXPECT().GetByEmail(gomock.Any(), login.Email).Return(nil, common.ErrNotFound)
		logger.EXPECT().Error(gomock.Any(), gomock.Any())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.Login(ctx, login)
		require.Error(t, err)
		require.Nil(t, result)
		assert.ErrorIs(t, err, user.ErrInvalidCredentials)
	})

	t.Run("invalid password", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		// Create a test user with different password
		testUser, err := entities.NewUser(login.Email, "correctpassword", "Test", "User")
		require.NoError(t, err)

		// Mock repository call
		repo.EXPECT().GetByEmail(gomock.Any(), login.Email).Return(testUser, nil)
		logger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.Login(ctx, login)
		require.Error(t, err)
		require.Nil(t, result)
		assert.ErrorIs(t, err, user.ErrInvalidCredentials)
	})

	t.Run("database error", func(t *testing.T) {
		login := &user.Login{
			Email:    "test@example.com",
			Password: "password123",
		}

		// Mock repository call
		repo.EXPECT().GetByEmail(gomock.Any(), login.Email).Return(nil, errors.New("database connection failed"))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.Login(ctx, login)
		require.Error(t, err)
		require.Nil(t, result)
		assert.ErrorIs(t, err, user.ErrInvalidCredentials)
	})
}

func TestService_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockuser.NewMockRepository(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	svc := user.NewService(repo, logger)

	t.Run("user found", func(t *testing.T) {
		userID := "test-user-id"
		expectedUser := &entities.User{
			ID:    userID,
			Email: "test@example.com",
		}

		repo.EXPECT().GetByID(gomock.Any(), userID).Return(expectedUser, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.GetUserByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, result)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent-user-id"

		repo.EXPECT().GetByID(gomock.Any(), userID).Return(nil, user.ErrUserNotFound)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.GetUserByID(ctx, userID)
		require.Error(t, err)
		require.Nil(t, result)
		assert.ErrorIs(t, err, user.ErrUserNotFound)
	})
}

func TestService_GetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := mockuser.NewMockRepository(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)

	svc := user.NewService(repo, logger)

	t.Run("user found", func(t *testing.T) {
		email := "test@example.com"
		expectedUser := &entities.User{
			ID:    "test-user-id",
			Email: email,
		}

		repo.EXPECT().GetByEmail(gomock.Any(), email).Return(expectedUser, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, expectedUser, result)
	})

	t.Run("user not found", func(t *testing.T) {
		email := "nonexistent@example.com"

		repo.EXPECT().GetByEmail(gomock.Any(), email).Return(nil, user.ErrUserNotFound)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := svc.GetUserByEmail(ctx, email)
		require.Error(t, err)
		require.Nil(t, result)
		assert.ErrorIs(t, err, user.ErrUserNotFound)
	})
}
