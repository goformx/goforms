package user

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidUserID indicates that the provided user ID is invalid
	ErrInvalidUserID = errors.New("invalid user ID")
)

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginResponse represents the response from a successful login
type LoginResponse struct {
	User  *User      `json:"user"`
	Token *TokenPair `json:"token"`
}

// User represents a user in the system
type User struct {
	ID             uint      `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	HashedPassword string    `json:"-" db:"hashed_password"`
	FirstName      string    `json:"first_name" db:"first_name"`
	LastName       string    `json:"last_name" db:"last_name"`
	Role           string    `json:"role" db:"role"`
	Active         bool      `json:"active" db:"active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Signup represents the user signup request
type Signup struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// Login represents the user login request
type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashedPassword)
	return nil
}

// CheckPassword verifies if the provided password matches the user's hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}
