package form

import (
	"context"
	"time"
)

// Form represents a form created by a user
type Form struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Schema      JSON      `json:"schema"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JSON is a type alias for the form schema
type JSON map[string]any

// Store defines the interface for form persistence
type Store interface {
	Create(form *Form) error
	GetByID(id uint) (*Form, error)
	GetByUserID(userID uint) ([]*Form, error)
	Update(form *Form) error
	Delete(id uint) error
}

// Service defines the interface for form business logic
type Service interface {
	CreateForm(userID uint, title, description string, schema JSON) (*Form, error)
	GetForm(id uint) (*Form, error)
	GetUserForms(userID uint) ([]*Form, error)
	UpdateForm(id uint, title, description string, schema JSON) (*Form, error)
	DeleteForm(id uint) error
}

// Field represents a form field
type Field struct {
	Name string
	Type string
}

// FormOptions represents form configuration options
type FormOptions struct {
	// Add form options as needed
}

// Response represents a form submission response
type Response struct {
	ID          string
	FormID      string
	Values      map[string]any
	SubmittedAt time.Time
}

// Client represents a form client interface
type Client interface {
	SubmitForm(ctx context.Context, form Form) error
	GetForm(ctx context.Context, formID string) (*Form, error)
	ListForms(ctx context.Context) ([]Form, error)
	DeleteForm(ctx context.Context, formID string) error
	UpdateForm(ctx context.Context, formID string, form Form) error
	SubmitResponse(ctx context.Context, formID string, response Response) error
	GetResponse(ctx context.Context, responseID string) (*Response, error)
	ListResponses(ctx context.Context, formID string) ([]Response, error)
	DeleteResponse(ctx context.Context, responseID string) error
}
