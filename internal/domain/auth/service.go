package auth

import (
	"context"
	"errors"

	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Authentication errors
var (
	ErrNotAuthenticated = errors.New("user is not authenticated")
)

// Service defines the interface for authentication operations
type Service interface {
	// GetAuthenticatedUser returns the currently authenticated user
	GetAuthenticatedUser(ctx context.Context) (*user.User, error)

	// RequireAuth ensures that a user is authenticated
	RequireAuth(ctx context.Context) error

	// RequireAdmin ensures that a user is authenticated and has admin privileges
	RequireAdmin(ctx context.Context) error

	// ValidateUser validates user credentials
	ValidateUser(ctx context.Context, email, password string) (*user.User, error)
}

// service implements the Service interface
type service struct {
	userService user.Service
	logger      logging.Logger
}

// NewService creates a new auth service
func NewService(userService user.Service, logger logging.Logger) Service {
	return &service{
		userService: userService,
		logger:      logger,
	}
}

// GetAuthenticatedUser returns the currently authenticated user
func (s *service) GetAuthenticatedUser(ctx context.Context) (*user.User, error) {
	// TODO: Implement user authentication logic
	return nil, ErrNotAuthenticated
}

// RequireAuth ensures that a user is authenticated
func (s *service) RequireAuth(ctx context.Context) error {
	// TODO: Implement authentication check
	return nil
}

// RequireAdmin ensures that a user is authenticated and has admin privileges
func (s *service) RequireAdmin(ctx context.Context) error {
	// TODO: Implement admin check
	return nil
}

// ValidateUser validates user credentials
func (s *service) ValidateUser(ctx context.Context, email, password string) (*user.User, error) {
	// Retrieve user by email
	u, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Verify password
	_, authErr := s.userService.Authenticate(ctx, email, password)
	if authErr != nil {
		return nil, authErr
	}

	return u, nil
}
