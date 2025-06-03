package form

import (
	"time"

	"github.com/goformx/goforms/internal/domain/form/model"
)

// JSON represents a JSON object
type JSON map[string]any

// Form represents a form in the system
type Form struct {
	ID          string    `json:"id"`
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
	GetByID(id string) (*Form, error)
	GetByUserID(userID uint) ([]*Form, error)
	Delete(id string) error
	Update(f *Form) error
	GetFormSubmissions(formID string) ([]*model.FormSubmission, error)
}

// Service defines the interface for form operations
type Service interface {
	CreateForm(userID uint, title, description string, schema JSON) (*Form, error)
	GetForm(id string) (*Form, error)
	GetUserForms(userID uint) ([]*Form, error)
	UpdateForm(form *Form) error
	DeleteForm(id string) error
	GetFormSubmissions(formID string) ([]*model.FormSubmission, error)
}
