package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Subscription represents a newsletter subscription
type Subscription struct {
	ID        uint      `json:"id" db:"id"`
	Email     string    `json:"email" db:"email" validate:"required,email"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate validates the subscription
func (s *Subscription) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

// SubscriptionStore defines the interface for subscription storage
type SubscriptionStore interface {
	Create(subscription *Subscription) error
}
