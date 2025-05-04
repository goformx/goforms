package form

import (
	"context"
	"time"

	"github.com/jonesrussell/goforms/internal/domain/form/model"
)

// JSON represents a JSON object
type JSON map[string]any

// Form represents a form in the system
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

// Store defines the interface for form storage operations
type Store interface {
	Create(f *Form) error
	GetByID(id uint) (*Form, error)
	GetByUserID(userID uint) ([]*Form, error)
	Delete(id uint) error
	Update(f *Form) error
	GetFormSubmissions(formID uint) ([]*model.FormSubmission, error)
}

// Service defines the interface for form operations
type Service interface {
	CreateForm(userID uint, title, description string, schema JSON) (*Form, error)
	GetForm(id uint) (*Form, error)
	GetUserForms(userID uint) ([]*Form, error)
	UpdateForm(form *Form) error
	DeleteForm(id uint) error
	GetFormSubmissions(formID uint) ([]*model.FormSubmission, error)
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
