// Package dto provides data transfer objects for application layer
package dto

import (
	"time"

	"github.com/goformx/goforms/internal/domain/entities"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User      *entities.User `json:"user"`
	SessionID string         `json:"session_id"`
	Token     string         `json:"token,omitempty"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// SignupRequest represents a signup request
type SignupRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// SignupResponse represents a successful signup response
type SignupResponse struct {
	User      *entities.User `json:"user"`
	SessionID string         `json:"session_id"`
	Token     string         `json:"token,omitempty"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	SessionID string `json:"session_id" validate:"required"`
}

// LogoutResponse represents a successful logout response
type LogoutResponse struct {
	Message string `json:"message"`
}

// AuthError represents authentication-related errors
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}
