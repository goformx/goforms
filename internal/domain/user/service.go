package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var (
	// ErrUserNotFound indicates that a user was not found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists indicates that a user with the given email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials indicates that the provided credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserExists indicates that a user with the given email already exists
	ErrUserExists = errors.New("user already exists")
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
	store  Store
}

// NewService creates a new user service
func NewService(store Store, logger logging.Logger) Service {
	return &ServiceImpl{
		store:  store,
		logger: logger,
	}
}

// SignUp registers a new user
func (s *ServiceImpl) SignUp(ctx context.Context, signup *Signup) (*User, error) {
	s.logger.Debug("starting signup process",
		logging.StringField("email", signup.Email),
		logging.StringField("first_name", signup.FirstName),
		logging.StringField("last_name", signup.LastName),
	)

	// Check if email already exists
	existingUser, err := s.store.GetByEmail(ctx, signup.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			s.logger.Debug("user not found, proceeding with signup",
				logging.StringField("email", signup.Email),
			)
		}
	}
	if existingUser != nil {
		s.logger.Debug("user already exists", logging.StringField("email", existingUser.Email))
		return nil, ErrUserExists
	}

	s.logger.Debug("proceeding with signup",
		logging.StringField("email", signup.Email),
		logging.StringField("first_name", signup.FirstName),
		logging.StringField("last_name", signup.LastName),
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
		s.logger.Error("failed to set password", logging.ErrorField("error", pwErr))
		return nil, fmt.Errorf("failed to set password: %w", pwErr)
	}

	// Save user
	if createErr := s.store.Create(ctx, user); createErr != nil {
		s.logger.Error("failed to create user in store", logging.ErrorField("error", createErr))
		return nil, fmt.Errorf("failed to create user: %w", createErr)
	}

	s.logger.Debug("user created successfully",
		logging.UintField("id", user.ID),
		logging.StringField("email", user.Email),
	)

	return user, nil
}

// Login authenticates a user
func (s *ServiceImpl) Login(ctx context.Context, login *Login) (*LoginResponse, error) {
	s.logger.Debug("attempting login",
		logging.StringField("email", login.Email),
		logging.BoolField("has_password", login.Password != ""),
	)

	user, err := s.store.GetByEmail(ctx, login.Email)
	if err != nil {
		s.logger.Error("failed to get user by email",
			logging.ErrorField("error", err),
			logging.StringField("email", login.Email),
		)
		return nil, ErrInvalidCredentials
	}
	if user == nil {
		s.logger.Error("user not found", logging.StringField("email", login.Email))
		return nil, ErrInvalidCredentials
	}

	s.logger.Debug("user found",
		logging.StringField("email", user.Email),
		logging.BoolField("active", user.Active),
	)

	if !user.CheckPassword(login.Password) {
		s.logger.Error("password mismatch", logging.StringField("email", login.Email))
		return nil, ErrInvalidCredentials
	}

	// TODO: Implement proper token generation
	// For now, return dummy tokens
	tokenPair := &TokenPair{
		AccessToken:  "dummy_access_token",
		RefreshToken: "dummy_refresh_token",
	}

	s.logger.Debug("login successful", logging.StringField("email", login.Email))
	return &LoginResponse{
		User:  user,
		Token: tokenPair,
	}, nil
}

// Logout blacklists a refresh token
func (s *ServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	s.logger.Debug("logging out user",
		logging.StringField("refresh_token", refreshToken),
	)

	// TODO: Implement token blacklisting
	// For now, we'll just log the logout attempt
	s.logger.Debug("logout successful",
		logging.StringField("refresh_token", refreshToken),
	)

	return nil
}

// GetUserByID retrieves a user by ID
func (s *ServiceImpl) GetUserByID(ctx context.Context, id uint) (*User, error) {
	return s.store.GetByID(ctx, id)
}

// GetUserByEmail retrieves a user by email
func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.store.GetByEmail(ctx, email)
}

// UpdateUser updates a user
func (s *ServiceImpl) UpdateUser(ctx context.Context, user *User) error {
	return s.store.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *ServiceImpl) DeleteUser(ctx context.Context, id uint) error {
	return s.store.Delete(ctx, id)
}

// ListUsers lists all users
func (s *ServiceImpl) ListUsers(ctx context.Context) ([]User, error) {
	return s.store.List(ctx)
}

// GetByID retrieves a user by ID string
func (s *ServiceImpl) GetByID(ctx context.Context, id string) (*User, error) {
	return s.store.GetByIDString(ctx, id)
}

// ValidateToken validates a token
func (s *ServiceImpl) ValidateToken(ctx context.Context, token string) error {
	// TODO: Implement token validation
	return nil
}

// GetUserIDFromToken extracts the user ID from a token
func (s *ServiceImpl) GetUserIDFromToken(ctx context.Context, token string) (uint, error) {
	// TODO: Implement token parsing
	return 0, nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *ServiceImpl) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	// TODO: Implement token blacklist check
	return false, nil
}

// Authenticate matches the domain.UserService interface
func (s *ServiceImpl) Authenticate(ctx context.Context, email, password string) (*User, error) {
	s.logger.Debug("attempting authenticate",
		logging.StringField("email", email),
		logging.BoolField("has_password", password != ""),
	)

	user, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("failed to get user by email",
			logging.ErrorField("error", err),
			logging.StringField("email", email),
		)
		return nil, ErrInvalidCredentials
	}
	if user == nil {
		s.logger.Error("user not found", logging.StringField("email", email))
		return nil, ErrInvalidCredentials
	}

	s.logger.Debug("user found",
		logging.StringField("email", user.Email),
		logging.BoolField("active", user.Active),
	)

	if !user.CheckPassword(password) {
		s.logger.Error("password mismatch", logging.StringField("email", email))
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
