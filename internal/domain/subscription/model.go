package subscription

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Status represents the status of a subscription
type Status string

const (
	// StatusPending indicates a pending subscription
	StatusPending Status = "pending"
	// StatusActive indicates an active subscription
	StatusActive Status = "active"
	// StatusInactive indicates an inactive subscription
	StatusInactive Status = "inactive"
	// StatusCancelled indicates a cancelled subscription
	StatusCancelled Status = "cancelled"
)

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusActive, StatusInactive, StatusCancelled:
		return true
	default:
		return false
	}
}

// Subscription represents a demo form submission
type Subscription struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Name      string    `json:"name" db:"name" validate:"required"`
	Status    Status    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the subscription data is valid
func (s *Subscription) Validate() error {
	if s == nil {
		return ErrInvalidSubscription
	}
	if s.Email == "" {
		return ErrEmailRequired
	}
	if !strings.Contains(s.Email, "@") || !strings.Contains(s.Email, ".") {
		return ErrInvalidEmail
	}
	if s.Name == "" {
		return ErrNameRequired
	}
	return nil
}

// ParseID parses a string into a subscription ID
func ParseID(id string) (int64, error) {
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse subscription ID %q: %w", id, err)
	}
	return parsed, nil
}
