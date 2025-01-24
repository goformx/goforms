package subscription

import (
	"errors"
	"time"
)

// Status represents the status of a subscription
type Status string

const (
	// StatusPending indicates a pending subscription
	StatusPending Status = "pending"
	// StatusActive indicates an active subscription
	StatusActive Status = "active"
	// StatusCancelled indicates a cancelled subscription
	StatusCancelled Status = "cancelled"
)

// Subscription represents a newsletter subscription
type Subscription struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Status    Status    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the subscription data is valid
func (s *Subscription) Validate() error {
	if s.Email == "" {
		return errors.New("email is required")
	}
	if s.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
