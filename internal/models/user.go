package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jonesrussell/goforms/internal/infrastructure/validation"
)

// User represents a user in the system
type User struct {
	ID             uint      `json:"id" db:"id"`
	Email          string    `json:"email" db:"email" validate:"required,email"`
	HashedPassword string    `json:"-" db:"hashed_password"`
	FirstName      string    `json:"first_name" db:"first_name" validate:"required"`
	LastName       string    `json:"last_name" db:"last_name" validate:"required"`
	Role           string    `json:"role" db:"role"`
	Active         bool      `json:"active" db:"active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// UserSignup represents the data needed for user registration
type UserSignup struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
	FirstName       string `json:"first_name" validate:"required"`
	LastName        string `json:"last_name" validate:"required"`
}

// UserLogin represents the data needed for user login
type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Validate validates the user data
func (u *User) Validate() error {
	return validation.New().Struct(u)
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashedBytes)
	return nil
}

// CheckPassword verifies if the provided password matches the hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}

// UserStore defines the interface for user storage operations
type UserStore interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id uint) error
	List() ([]User, error)
}
