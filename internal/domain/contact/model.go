package contact

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNameRequired indicates that the name field is required
	ErrNameRequired = errors.New("name is required")
	// ErrEmailRequired indicates that the email field is required
	ErrEmailRequired = errors.New("email is required")
	// ErrMessageRequired indicates that the message field is required
	ErrMessageRequired = errors.New("message is required")
	// ErrInvalidStatus indicates that the status is invalid
	ErrInvalidStatus = errors.New("invalid status")
)

// Status represents the status of a contact form submission
type Status string

const (
	// StatusPending indicates a pending submission
	StatusPending Status = "pending"
	// StatusApproved indicates an approved submission
	StatusApproved Status = "approved"
	// StatusRejected indicates a rejected submission
	StatusRejected Status = "rejected"
)

// Submission represents a contact form submission
type Submission struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Message   string    `json:"message" db:"message"`
	Status    Status    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Store defines the interface for contact form persistence
type Store interface {
	Create(ctx context.Context, sub *Submission) error
	List(ctx context.Context) ([]Submission, error)
	Get(ctx context.Context, id int64) (*Submission, error)
	UpdateStatus(ctx context.Context, id int64, status Status) error
}
