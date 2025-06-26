package web

import (
	"context"
	"fmt"

	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/user"
)

// AuthService provides authentication services.
type AuthService struct {
	UserService    user.Service
	SessionManager *session.Manager
}

// NewAuthService creates a new AuthService.
func NewAuthService(userService user.Service, sessionManager *session.Manager) *AuthService {
	return &AuthService{
		UserService:    userService,
		SessionManager: sessionManager,
	}
}

// Login authenticates a user and returns the user and session ID
func (s *AuthService) Login(ctx context.Context, email, password, userAgent string) (*entities.User, string, error) {
	loginResp, err := s.UserService.Login(ctx, &user.Login{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, "", fmt.Errorf("user login failed: %w", err)
	}

	sessionID, err := s.SessionManager.CreateSession(loginResp.User.ID, loginResp.User.Email, userAgent)
	if err != nil {
		return nil, "", fmt.Errorf("create session failed: %w", err)
	}

	return loginResp.User, sessionID, nil
}

// Signup creates a new user and session ID
func (s *AuthService) Signup(
	ctx context.Context,
	signup user.Signup,
	userAgent string,
) (*entities.User, string, error) {
	newUser, err := s.UserService.SignUp(ctx, &signup)
	if err != nil {
		return nil, "", fmt.Errorf("user signup failed: %w", err)
	}

	sessionID, err := s.SessionManager.CreateSession(newUser.ID, newUser.Email, userAgent)
	if err != nil {
		return newUser, "", fmt.Errorf("create session failed: %w", err)
	}

	return newUser, sessionID, nil
}
