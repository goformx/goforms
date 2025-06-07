package form

import (
	"context"
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

// Service defines the interface for form operations
type Service interface {
	// CreateForm creates a new form
	CreateForm(ctx context.Context, userID uint, title, description string, schema model.JSON) (*model.Form, error)

	// GetForm retrieves a form by its ID
	GetForm(ctx context.Context, id string) (*model.Form, error)

	// GetUserForms retrieves all forms for a specific user
	GetUserForms(ctx context.Context, userID uint) ([]*model.Form, error)

	// UpdateForm updates an existing form
	UpdateForm(ctx context.Context, form *model.Form) error

	// DeleteForm deletes a form by its ID
	DeleteForm(ctx context.Context, id string) error

	// GetFormSubmissions returns all submissions for a form
	GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error)
}
