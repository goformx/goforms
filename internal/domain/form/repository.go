package form

import (
	"context"
	"errors"

	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
)

// ErrFormSchemaNotFound is returned when a form schema cannot be found
var ErrFormSchemaNotFound = errors.New("form schema not found")

// Repository defines the interface for form data access
type Repository interface {
	// Form methods
	CreateForm(ctx context.Context, form *model.Form) error
	GetFormByID(ctx context.Context, formID string) (*model.Form, error)
	ListForms(ctx context.Context, offset, limit int) ([]*model.Form, error)
	UpdateForm(ctx context.Context, form *model.Form) error
	DeleteForm(ctx context.Context, formID string) error
	GetFormsByStatus(ctx context.Context, active bool) ([]*model.Form, error)

	// Submission methods
	CreateSubmission(ctx context.Context, submission *model.FormSubmission) error
	GetSubmissionByID(ctx context.Context, submissionID string) (*model.FormSubmission, error)
	ListSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error)
	UpdateSubmission(ctx context.Context, submission *model.FormSubmission) error
	DeleteSubmission(ctx context.Context, submissionID string) error
	GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error)
	GetByFormIDPaginated(ctx context.Context, formID string, params common.PaginationParams) (*common.PaginationResult, error)
	GetByFormAndUser(ctx context.Context, formID, userID string) (*model.FormSubmission, error)
	GetSubmissionsByStatus(ctx context.Context, status model.SubmissionStatus) ([]*model.FormSubmission, error)
}
