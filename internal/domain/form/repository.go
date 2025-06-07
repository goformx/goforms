package form

import (
	"context"
	"errors"

	"github.com/goformx/goforms/internal/domain/form/model"
)

var (
	// ErrFormSchemaNotFound is returned when a form schema cannot be found
	ErrFormSchemaNotFound = errors.New("form schema not found")
)

// Repository defines the interface for form data access
type Repository interface {
	// Create creates a new form
	Create(ctx context.Context, form *model.Form) error
	// GetByID gets a form by ID
	GetByID(ctx context.Context, id string) (*model.Form, error)
	// GetByUserID gets all forms for a user
	GetByUserID(ctx context.Context, userID uint) ([]*model.Form, error)
	// Update updates a form
	Update(ctx context.Context, form *model.Form) error
	// Delete deletes a form
	Delete(ctx context.Context, id string) error
	// GetFormSubmissions gets all submissions for a form
	GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error)
}

// SubmissionStore defines the interface for form submission persistence
type SubmissionStore interface {
	// Create creates a new form submission
	Create(ctx context.Context, submission *model.FormSubmission) error

	// GetByID retrieves a form submission by its ID
	GetByID(ctx context.Context, id string) (*model.FormSubmission, error)

	// GetByFormID retrieves all submissions for a specific form
	GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error)

	// Update updates an existing form submission
	Update(ctx context.Context, submission *model.FormSubmission) error

	// Delete deletes a form submission by its ID
	Delete(ctx context.Context, id string) error
}
