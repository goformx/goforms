//go:generate mockgen -typed -source=service.go -destination=../../../test/mocks/user/mock_service.go -package=user

package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
)

// Signup represents a user signup request
type Signup struct {
	Email           string `json:"email"            validate:"required,email"`
	Password        string `json:"password"         validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=password"`
}

// Login represents a user login request
type Login struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a user login response
type LoginResponse struct {
	User *entities.User
}

var (
	// ErrUserNotFound indicates that a user was not found
	ErrUserNotFound = domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
	// ErrEmailAlreadyExists indicates that a user with the given email already exists
	ErrEmailAlreadyExists = domainerrors.New(domainerrors.ErrCodeAlreadyExists, "email already exists", nil)
	// ErrInvalidCredentials indicates that the provided credentials are invalid
	ErrInvalidCredentials = domainerrors.New(domainerrors.ErrCodeAuthentication, "invalid credentials", nil)
	// ErrUserExists indicates that a user with the given email already exists
	ErrUserExists = domainerrors.New(domainerrors.ErrCodeAlreadyExists, "user already exists", nil)
)

// Service defines the user service interface
type Service interface {
	SignUp(ctx context.Context, signup *Signup) (*entities.User, error)
	Login(ctx context.Context, login *Login) (*LoginResponse, error)
	Logout(ctx context.Context) error
	GetUserByID(ctx context.Context, id string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, offset, limit int) ([]*entities.User, error)
	Authenticate(ctx context.Context, email, password string) (*entities.User, error)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	logger logging.Logger
	repo   Repository
}

// NewService creates a new user service
func NewService(repo Repository, logger logging.Logger) Service {
	return &ServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// SignUp registers a new user
func (s *ServiceImpl) SignUp(ctx context.Context, signup *Signup) (*entities.User, error) {
	// Check if email already exists
	existingUser, err := s.repo.GetByEmail(ctx, signup.Email)
	if err != nil {
		// Check if it's a "not found" error using errors.Is (expected for new users)
		if !errors.Is(err, common.ErrNotFound) && !domainerrors.IsNotFound(err) {
			// It's a real error, not just "not found"
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		}
		// User doesn't exist, which is what we want for signup
		existingUser = nil
	}

	if existingUser != nil {
		return nil, ErrUserExists
	}

	// Create user with default first/last name (extract username before @)
	atIndex := strings.Index(signup.Email, "@")
	if atIndex == -1 {
		return nil, errors.New("invalid email format: missing @ symbol")
	}
	username := signup.Email[:atIndex]

	user, err := entities.NewUser(signup.Email, signup.Password, username, "")
	if err != nil {
		s.logger.Error("failed to create user", "error", err)

		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save user
	if createErr := s.repo.Create(ctx, user); createErr != nil {
		s.logger.Error("failed to create user", "error", createErr)

		return nil, fmt.Errorf("create: %w", createErr)
	}

	return user, nil
}

// Login authenticates a user
func (s *ServiceImpl) Login(ctx context.Context, login *Login) (*LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, login.Email)
	if err != nil {
		s.logger.Error("failed to get user by email", "error", err)

		return nil, ErrInvalidCredentials
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.CheckPassword(login.Password) {
		s.logger.Error("password mismatch", "email", login.Email)

		return nil, ErrInvalidCredentials
	}

	return &LoginResponse{
		User: user,
	}, nil
}

// Logout handles user logout
func (s *ServiceImpl) Logout(_ context.Context) error {
	// Session-based logout is handled by session middleware
	return nil
}

// GetUserByID retrieves a user by ID
func (s *ServiceImpl) GetUserByID(ctx context.Context, id string) (*entities.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

// UpdateUser updates a user
func (s *ServiceImpl) UpdateUser(ctx context.Context, user *entities.User) error {
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user
func (s *ServiceImpl) DeleteUser(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}

// ListUsers lists all users
func (s *ServiceImpl) ListUsers(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	users, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

// Authenticate matches the domain.UserService interface
func (s *ServiceImpl) Authenticate(ctx context.Context, email, password string) (*entities.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if domainerrors.GetErrorCode(err) == domainerrors.ErrCodeNotFound {
			return nil, ErrInvalidCredentials
		}

		wrappedErr := domainerrors.WrapError(err, domainerrors.ErrCodeAuthentication, "failed to get user by email")

		return nil, fmt.Errorf("authenticate user: %w", wrappedErr)
	}

	if !user.CheckPassword(password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
