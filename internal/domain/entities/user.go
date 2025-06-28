// Package entities provides domain entities that represent the core business
// objects in the application, including users, forms, and other domain models.
package entities

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// ErrInvalidEmail represents an invalid email format error
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrInvalidPassword represents an invalid password error
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
)

const (
	// MinPasswordLength is the minimum length required for passwords
	MinPasswordLength = 8
)

// User represents a user entity
type User struct {
	ID             string         `json:"id" gorm:"column:uuid;primaryKey;type:uuid;default:gen_random_uuid()"`
	Email          string         `json:"email" gorm:"uniqueIndex;not null;size:255"`
	HashedPassword string         `json:"-" gorm:"column:hashed_password;not null;size:255"`
	FirstName      string         `json:"first_name" gorm:"not null;size:100"`
	LastName       string         `json:"last_name" gorm:"not null;size:100"`
	Role           string         `json:"role" gorm:"not null;size:50;default:user"`
	Active         bool           `json:"active" gorm:"not null;default:true"`
	CreatedAt      time.Time      `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"not null;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for the User model
func (u *User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}

	if u.Role == "" {
		u.Role = "user"
	}

	if !u.Active {
		u.Active = true
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(_ *gorm.DB) error {
	u.UpdatedAt = time.Now()

	return nil
}

// AfterFind is a GORM hook that runs after finding a user
func (u *User) AfterFind(_ *gorm.DB) error {
	// Ensure UUID is properly formatted
	if u.ID != "" {
		// Try to parse as UUID to validate format
		if _, err := uuid.Parse(u.ID); err != nil {
			return fmt.Errorf("invalid UUID format: %w", err)
		}
	}

	return nil
}

// NewUser creates a new user instance with validation
func NewUser(email, password, firstName, lastName string) (*User, error) {
	if email == "" {
		return nil, ErrInvalidEmail
	}

	if len(password) < MinPasswordLength {
		return nil, ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return &User{
		ID:             uuid.New().String(),
		Email:          email,
		HashedPassword: string(hashedPassword),
		FirstName:      firstName,
		LastName:       lastName,
		Role:           "user",
		Active:         true,
	}, nil
}

// Validate performs validation on the user entity
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}

	if u.HashedPassword == "" {
		return errors.New("password is required")
	}

	if !isValidEmail(u.Email) {
		return ErrInvalidEmail
	}

	return nil
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	if len(password) < MinPasswordLength {
		return ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("generate password hash: %w", err)
	}

	u.HashedPassword = string(hashedPassword)

	return nil
}

// CheckPassword verifies if the provided password matches the user's hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))

	return err == nil
}

// Deactivate marks the user as inactive
func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now()
}

// Activate marks the user as active
func (u *User) Activate() {
	u.Active = true
	u.UpdatedAt = time.Now()
}

// UpdateProfile updates the user's profile information
func (u *User) UpdateProfile(firstName, lastName string) {
	u.FirstName = firstName
	u.LastName = lastName
	u.UpdatedAt = time.Now()
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// isValidEmail validates an email address
func isValidEmail(email string) bool {
	// TODO: Implement proper email validation
	return len(email) > 3 && len(email) < 255
}

// GetID returns the user's ID
func (u *User) GetID() string {
	return u.ID
}

// SetID sets the user's ID
func (u *User) SetID(id string) {
	u.ID = id
}

// ValidatePassword validates a password
func (u *User) ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return ErrInvalidPassword
	}

	return nil
}
