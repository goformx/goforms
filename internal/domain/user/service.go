package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var (
	// ErrUserNotFound indicates that a user was not found
	ErrUserNotFound = domainerrors.New(domainerrors.ErrCodeNotFound, "user not found", nil)
	// ErrEmailAlreadyExists indicates that a user with the given email already exists
	ErrEmailAlreadyExists = domainerrors.New(domainerrors.ErrCodeAlreadyExists, "email already exists", nil)
	// ErrInvalidCredentials indicates that the provided credentials are invalid
	ErrInvalidCredentials = domainerrors.New(domainerrors.ErrCodeAuthentication, "invalid credentials", nil)
	// ErrUserExists indicates that a user with the given email already exists
	ErrUserExists = domainerrors.New(domainerrors.ErrCodeAlreadyExists, "user already exists", nil)
	// ErrInvalidToken indicates that the provided token is invalid
	ErrInvalidToken = domainerrors.New(domainerrors.ErrCodeInvalidToken, "invalid token", nil)
	// ErrTokenExpired indicates that the provided token has expired
	ErrTokenExpired = domainerrors.New(domainerrors.ErrCodeInvalidToken, "token has expired", nil)
)

// Service defines the user service interface
type Service interface {
	SignUp(ctx context.Context, signup *Signup) (*User, error)
	Login(ctx context.Context, login *Login) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserByID(ctx context.Context, id uint) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	ValidateToken(ctx context.Context, token string) error
	GetUserIDFromToken(ctx context.Context, token string) (uint, error)
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	Authenticate(ctx context.Context, email, password string) (*User, error)
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
func (s *ServiceImpl) SignUp(ctx context.Context, signup *Signup) (*User, error) {
	s.logger.Debug("starting signup process",
		logging.String("email", signup.Email),
		logging.String("first_name", signup.FirstName),
		logging.String("last_name", signup.LastName),
	)

	// Check if email already exists
	existingUser, err := s.repo.GetByEmail(ctx, signup.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			s.logger.Debug("user not found, proceeding with signup",
				logging.String("email", signup.Email),
			)
		}
	}
	if existingUser != nil {
		s.logger.Debug("user already exists", logging.String("email", existingUser.Email))
		return nil, ErrUserExists
	}

	s.logger.Debug("proceeding with signup",
		logging.String("email", signup.Email),
		logging.String("first_name", signup.FirstName),
		logging.String("last_name", signup.LastName),
	)

	s.logger.Debug("creating new user")

	// Create user
	user := &User{
		Email:     signup.Email,
		FirstName: signup.FirstName,
		LastName:  signup.LastName,
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set password
	if pwErr := user.SetPassword(signup.Password); pwErr != nil {
		s.logger.Error("failed to set password", logging.Error(pwErr))
		return nil, fmt.Errorf("failed to set password: %w", pwErr)
	}

	// Save user
	if createErr := s.repo.Create(ctx, user); createErr != nil {
		s.logger.Error("failed to create user in store", logging.Error(createErr))
		return nil, fmt.Errorf("failed to create user: %w", createErr)
	}

	s.logger.Debug("user created successfully",
		logging.Uint("id", user.ID),
		logging.String("email", user.Email),
	)

	return user, nil
}

// Login authenticates a user
func (s *ServiceImpl) Login(ctx context.Context, login *Login) (*LoginResponse, error) {
	s.logger.Debug("attempting login",
		logging.String("email", login.Email),
		logging.Bool("has_password", login.Password != ""),
	)

	user, err := s.repo.GetByEmail(ctx, login.Email)
	if err != nil {
		s.logger.Error("failed to get user by email",
			logging.Error(err),
			logging.String("email", login.Email),
		)
		return nil, ErrInvalidCredentials
	}
	if user == nil {
		s.logger.Error("user not found", logging.String("email", login.Email))
		return nil, ErrInvalidCredentials
	}

	s.logger.Debug("user found",
		logging.String("email", user.Email),
		logging.Bool("active", user.Active),
	)

	if !user.CheckPassword(login.Password) {
		s.logger.Error("password mismatch", logging.String("email", login.Email))
		return nil, ErrInvalidCredentials
	}

	// TODO: Implement proper token generation
	// For now, return dummy tokens
	tokenPair := &TokenPair{
		AccessToken:  "dummy_access_token",
		RefreshToken: "dummy_refresh_token",
	}

	s.logger.Debug("login successful", logging.String("email", login.Email))
	return &LoginResponse{
		User:  user,
		Token: tokenPair,
	}, nil
}

// Logout blacklists a refresh token
func (s *ServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	s.logger.Debug("logging out user",
		logging.String("refresh_token", refreshToken),
	)

	// TODO: Implement token blacklisting
	// For now, we'll just log the logout attempt
	s.logger.Debug("logout successful",
		logging.String("refresh_token", refreshToken),
	)

	return nil
}

// GetUserByID retrieves a user by ID
func (s *ServiceImpl) GetUserByID(ctx context.Context, id uint) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetUserByEmail retrieves a user by email
func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}

// UpdateUser updates a user
func (s *ServiceImpl) UpdateUser(ctx context.Context, user *User) error {
	return s.repo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *ServiceImpl) DeleteUser(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// ListUsers lists all users
func (s *ServiceImpl) ListUsers(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

// GetByID retrieves a user by ID string
func (s *ServiceImpl) GetByID(ctx context.Context, id string) (*User, error) {
	return s.repo.GetByIDString(ctx, id)
}

// ValidateToken validates a token
func (s *ServiceImpl) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidToken
	}
	// TODO: Implement proper JWT validation
	return nil
}

// GetUserIDFromToken extracts the user ID from a token
func (s *ServiceImpl) GetUserIDFromToken(ctx context.Context, token string) (uint, error) {
	if token == "" {
		return 0, ErrInvalidToken
	}
	// TODO: Implement proper JWT parsing
	return 0, nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *ServiceImpl) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, ErrInvalidToken
	}
	// TODO: Implement proper token blacklist check
	return false, nil
}

// Authenticate matches the domain.UserService interface
func (s *ServiceImpl) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if domainerrors.GetErrorCode(err) == domainerrors.ErrCodeNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, domainerrors.WrapError(err, domainerrors.ErrCodeAuthentication, "failed to get user by email")
	}

	if !user.CheckPassword(password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
