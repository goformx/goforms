package services

import (
	"context"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/application/dto"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/common/interfaces"
	"github.com/goformx/goforms/internal/domain/user"
)

// AuthUseCaseService handles authentication use cases
type AuthUseCaseService struct {
	userService    user.Service
	sessionManager *session.Manager
	logger         interfaces.Logger
}

// NewAuthUseCaseService creates a new authentication use case service
func NewAuthUseCaseService(
	userService user.Service,
	sessionManager *session.Manager,
	logger interfaces.Logger,
) *AuthUseCaseService {
	return &AuthUseCaseService{
		userService:    userService,
		sessionManager: sessionManager,
		logger:         logger,
	}
}

// Login handles user login use case
func (s *AuthUseCaseService) Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error) {
	s.logger.Info("processing login request", "email", request.Email)

	// Convert DTO to domain request
	loginRequest := &user.Login{
		Email:    request.Email,
		Password: request.Password,
	}

	// Call domain service
	loginResponse, err := s.userService.Login(ctx, loginRequest)
	if err != nil {
		s.logger.Error("login failed", "email", request.Email, "error", err)

		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Create session
	sessionID, err := s.sessionManager.CreateSession(
		loginResponse.User.ID,
		loginResponse.User.Email,
		loginResponse.User.Role,
	)
	if err != nil {
		s.logger.Error("failed to create session", "user_id", loginResponse.User.ID, "error", err)

		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Build response
	response := &dto.LoginResponse{
		User:      loginResponse.User,
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session expires in 24 hours
	}

	s.logger.Info("login successful", "user_id", loginResponse.User.ID, "session_id", sessionID)

	return response, nil
}

// Signup handles user signup use case
func (s *AuthUseCaseService) Signup(ctx context.Context, request *dto.SignupRequest) (*dto.SignupResponse, error) {
	s.logger.Info("processing signup request", "email", request.Email)

	// Convert DTO to domain request
	signupRequest := &user.Signup{
		Email:           request.Email,
		Password:        request.Password,
		ConfirmPassword: request.ConfirmPassword,
	}

	// Call domain service
	userEntity, err := s.userService.SignUp(ctx, signupRequest)
	if err != nil {
		s.logger.Error("signup failed", "email", request.Email, "error", err)

		return nil, fmt.Errorf("signup failed: %w", err)
	}

	// Create session
	sessionID, err := s.sessionManager.CreateSession(
		userEntity.ID,
		userEntity.Email,
		userEntity.Role,
	)
	if err != nil {
		s.logger.Error("failed to create session", "user_id", userEntity.ID, "error", err)

		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Build response
	response := &dto.SignupResponse{
		User:      userEntity,
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session expires in 24 hours
	}

	s.logger.Info("signup successful", "user_id", userEntity.ID, "session_id", sessionID)

	return response, nil
}

// Logout handles user logout use case
func (s *AuthUseCaseService) Logout(ctx context.Context, request *dto.LogoutRequest) (*dto.LogoutResponse, error) {
	s.logger.Info("processing logout request", "user_id", request.UserID, "session_id", request.SessionID)

	// Destroy session
	s.sessionManager.DeleteSession(request.SessionID)

	response := &dto.LogoutResponse{
		Message: "Successfully logged out",
	}

	s.logger.Info("logout successful", "user_id", request.UserID, "session_id", request.SessionID)

	return response, nil
}

// ValidateSession validates a user session
func (s *AuthUseCaseService) ValidateSession(ctx context.Context, sessionID string) (*dto.LoginResponse, error) {
	s.logger.Debug("validating session", "session_id", sessionID)

	// Get session data
	sessionData, exists := s.sessionManager.GetSession(sessionID)
	if !exists {
		s.logger.Debug("session not found or expired", "session_id", sessionID)

		return nil, fmt.Errorf("invalid session: session not found")
	}

	// Get user from domain service
	userEntity, err := s.userService.GetUserByID(ctx, sessionData.UserID)
	if err != nil {
		s.logger.Error("failed to get user for session", "user_id", sessionData.UserID, "error", err)

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	response := &dto.LoginResponse{
		User:      userEntity,
		SessionID: sessionID,
		ExpiresAt: sessionData.ExpiresAt,
	}

	return response, nil
}
