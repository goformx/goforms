package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/jonesrussell/goforms/internal/logger"
	"github.com/jonesrussell/goforms/internal/models"
)

var (
	// ErrUserNotFound indicates that a user was not found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists indicates that a user with the given email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials indicates that the provided credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Service defines the interface for user operations
type Service interface {
	SignUp(ctx context.Context, signup *models.UserSignup) (*models.User, error)
	Login(ctx context.Context, login *models.UserLogin) (string, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context) ([]models.User, error)
}

// ServiceImpl implements the user Service interface
type ServiceImpl struct {
	log       logger.Logger
	store     models.UserStore
	jwtSecret []byte
}

// NewService creates a new user service
func NewService(log logger.Logger, store models.UserStore, jwtSecret string) Service {
	return &ServiceImpl{
		log:       log,
		store:     store,
		jwtSecret: []byte(jwtSecret),
	}
}

// SignUp handles user registration
func (s *ServiceImpl) SignUp(ctx context.Context, signup *models.UserSignup) (*models.User, error) {
	// Check if email already exists
	existing, err := s.store.GetByEmail(signup.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		s.log.Error("failed to check existing user", logger.Error(err))
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Create new user
	user := &models.User{
		Email:     signup.Email,
		FirstName: signup.FirstName,
		LastName:  signup.LastName,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set password
	if err := user.SetPassword(signup.Password); err != nil {
		s.log.Error("failed to hash password", logger.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user
	if err := s.store.Create(user); err != nil {
		s.log.Error("failed to create user", logger.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login handles user authentication and returns a JWT token
func (s *ServiceImpl) Login(ctx context.Context, login *models.UserLogin) (string, error) {
	// Get user by email
	user, err := s.store.GetByEmail(login.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		s.log.Error("failed to get user", logger.Error(err))
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if !user.CheckPassword(login.Password) {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.log.Error("failed to generate token", logger.Error(err))
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// GetUserByID retrieves a user by ID
func (s *ServiceImpl) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.store.GetByID(id)
	if err != nil {
		s.log.Error("failed to get user", logger.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.store.GetByEmail(email)
	if err != nil {
		s.log.Error("failed to get user", logger.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateUser updates user information
func (s *ServiceImpl) UpdateUser(ctx context.Context, user *models.User) error {
	if err := s.store.Update(user); err != nil {
		s.log.Error("failed to update user", logger.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// DeleteUser removes a user
func (s *ServiceImpl) DeleteUser(ctx context.Context, id uint) error {
	if err := s.store.Delete(id); err != nil {
		s.log.Error("failed to delete user", logger.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers returns all users
func (s *ServiceImpl) ListUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.store.List()
	if err != nil {
		s.log.Error("failed to list users", logger.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}
