package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ErrInvalidUserID indicates that the provided user ID is invalid
var ErrInvalidUserID = errors.New("invalid user ID")

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
	ID             uint           `json:"id" gorm:"primaryKey"`
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
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Generate a unique ID using UUID
	u.ID = uint(uuid.New().ID())
	if u.Role == "" {
		u.Role = "user"
	}
	if !u.Active {
		u.Active = true
	}
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
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
