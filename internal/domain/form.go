package domain

import "context"

// Form represents a form in the system
type Form struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	UserID      string `json:"user_id"`
}

// FormService defines the interface for form operations
type FormService interface {
	// GetAllForms retrieves all forms
	GetAllForms(ctx context.Context) ([]Form, error)

	// GetFormByID retrieves a form by its ID
	GetFormByID(ctx context.Context, id string) (*Form, error)

	// GetUserForms retrieves all forms for a specific user
	GetUserForms(ctx context.Context, userID string) ([]Form, error)

	// CreateForm creates a new form
	CreateForm(ctx context.Context, form *Form) error

	// UpdateForm updates an existing form
	UpdateForm(ctx context.Context, form *Form) error

	// DeleteForm deletes a form by its ID
	DeleteForm(ctx context.Context, id string) error
}
