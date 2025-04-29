package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

var (
	// ErrUserNotFound indicates that a user was not found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists indicates that a user with the given email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrInvalidCredentials indicates that the provided credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidToken indicates that the provided token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenBlacklisted indicates that the token has been blacklisted
	ErrTokenBlacklisted = errors.New("token is blacklisted")
	// ErrInvalidUserIDClaim indicates that the user_id claim type is invalid
	ErrInvalidUserIDClaim = errors.New("invalid user_id claim type")
	// ErrInvalidUserID indicates that the user_id claim type is invalid
	ErrInvalidUserID = errors.New("invalid user_id claim type")
	// ErrUserExists indicates that a user with the given email already exists
	ErrUserExists = errors.New("user already exists")
)

const (
	accessTokenExpiry  = 15 * time.Minute
	refreshTokenExpiry = 7 * 24 * time.Hour
)

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Service defines the user service interface
type Service interface {
	SignUp(ctx context.Context, signup *Signup) (*User, error)
	Login(ctx context.Context, login *Login) (*TokenPair, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	GetUserByID(ctx context.Context, id uint) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context) ([]User, error)
	ValidateToken(token string) (*jwt.Token, error)
	IsTokenBlacklisted(token string) bool
	GetUserIDFromToken(token string) (string, error)
	GetByID(ctx context.Context, id string) (*User, error)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	logger         logging.Logger
	store          Store
	jwtSecret      []byte
	tokenBlacklist sync.Map // Using new sync.Map implementation from Go 1.24
}

// NewService creates a new user service
func NewService(store Store, logger logging.Logger, jwtSecret string) Service {
	return &ServiceImpl{
		store:     store,
		logger:    logger,
		jwtSecret: []byte(jwtSecret),
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
	existingUser, err := s.store.GetByEmail(ctx, signup.Email)
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
	if createErr := s.store.Create(ctx, user); createErr != nil {
		s.logger.Error("failed to create user in store", logging.Error(createErr))
		return nil, fmt.Errorf("failed to create user: %w", createErr)
	}

	s.logger.Debug("user created successfully", 
		logging.Uint("id", user.ID),
		logging.String("email", user.Email),
	)

	return user, nil
}

// Login authenticates a user and returns a token pair
func (s *ServiceImpl) Login(ctx context.Context, login *Login) (*TokenPair, error) {
	s.logger.Debug("attempting login", 
		logging.String("email", login.Email),
		logging.Bool("has_password", login.Password != ""),
	)

	user, err := s.store.GetByEmail(ctx, login.Email)
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

	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		s.logger.Error("failed to generate token pair", 
			logging.Error(err),
			logging.String("email", login.Email),
		)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.logger.Debug("login successful", logging.String("email", login.Email))
	return tokenPair, nil
}

// Logout adds a token to the blacklist
func (s *ServiceImpl) Logout(ctx context.Context, token string) error {
	s.tokenBlacklist.Store(token, time.Now())
	return nil
}

// RefreshToken generates a new token pair using a refresh token
func (s *ServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	token, validateErr := s.ValidateToken(refreshToken)
	if validateErr != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", ErrInvalidToken)
	}

	// Check if token is blacklisted
	if s.IsTokenBlacklisted(refreshToken) {
		return nil, fmt.Errorf("failed to refresh token: %w", ErrTokenBlacklisted)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to refresh token: %w", ErrInvalidToken)
	}

	// Get user from claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, ErrInvalidUserIDClaim
	}
	user, lookupErr := s.GetUserByID(ctx, uint(userID))
	if lookupErr != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", lookupErr)
	}

	// Generate new token pair
	tokenPair, genErr := s.generateTokenPair(user)
	if genErr != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", genErr)
	}

	// Blacklist the old refresh token
	s.tokenBlacklist.Store(refreshToken, true)

	return tokenPair, nil
}

// ValidateToken validates a JWT token
func (s *ServiceImpl) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	validateErr := s.validateTokenClaims(token)
	if validateErr != nil {
		return nil, validateErr
	}

	return token, nil
}

// parseToken parses and validates the JWT token
func (s *ServiceImpl) parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return token, nil
}

// validateTokenClaims validates the token claims using maps package
func (s *ServiceImpl) validateTokenClaims(token *jwt.Token) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	// Check required claims
	requiredClaims := []string{"user_id", "exp"}
	for _, claim := range requiredClaims {
		if _, exists := claims[claim]; !exists {
			return errors.New("missing required claims")
		}
	}

	if err := s.validateUserIDClaim(claims); err != nil {
		return err
	}

	if err := s.validateExpirationClaim(claims); err != nil {
		return err
	}

	return nil
}

// validateUserIDClaim validates the user_id claim
func (s *ServiceImpl) validateUserIDClaim(claims jwt.MapClaims) error {
	if _, ok := claims["user_id"].(float64); !ok {
		return ErrInvalidUserIDClaim
	}
	return nil
}

// validateExpirationClaim validates the exp claim
func (s *ServiceImpl) validateExpirationClaim(claims jwt.MapClaims) error {
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid exp claim")
	}

	if time.Now().Unix() > int64(exp) {
		return errors.New("token expired")
	}

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (s *ServiceImpl) IsTokenBlacklisted(token string) bool {
	_, exists := s.tokenBlacklist.Load(token)
	return exists
}

// generateTokenPair creates a new access and refresh token pair
func (s *ServiceImpl) generateTokenPair(user *User) (*TokenPair, error) {
	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"type":    "access",
		"exp":     time.Now().Add(accessTokenExpiry).Unix(),
	})

	// Generate refresh token with longer expiry
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"type":    "refresh",
		"exp":     time.Now().Add(refreshTokenExpiry).Unix(),
	})

	// Sign tokens
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *ServiceImpl) GetUserByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.store.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		s.logger.Error("failed to get user by id", logging.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.store.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		s.logger.Error("failed to get user by email", logging.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing user
func (s *ServiceImpl) UpdateUser(ctx context.Context, user *User) error {
	if err := s.store.Update(ctx, user); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		s.logger.Error("failed to update user", logging.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// DeleteUser removes a user by ID
func (s *ServiceImpl) DeleteUser(ctx context.Context, id uint) error {
	if err := s.store.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		s.logger.Error("failed to delete user", logging.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers returns all users
func (s *ServiceImpl) ListUsers(ctx context.Context) ([]User, error) {
	users, err := s.store.List(ctx)
	if err != nil {
		s.logger.Error("failed to list users", logging.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

// GetUserIDFromToken retrieves the user ID from a token
func (s *ServiceImpl) GetUserIDFromToken(token string) (string, error) {
	parsedToken, err := s.parseToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidUserIDClaim
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return "", ErrInvalidUserIDClaim
	}

	return strconv.FormatInt(int64(userID), 10), nil
}

// GetByID retrieves a user by ID string
func (s *ServiceImpl) GetByID(ctx context.Context, id string) (*User, error) {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	return s.GetUserByID(ctx, uint(userID))
}
