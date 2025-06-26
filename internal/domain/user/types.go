package user

import "github.com/goformx/goforms/internal/domain/entities"

// Signup represents a user signup request
type Signup struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=password"`
}

// Login represents a user login request
type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a user login response
type LoginResponse struct {
	User *entities.User
}
