package form

import (
	"context"

	"github.com/goformx/goforms/internal/domain/common/repository"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
)

// SubmissionRepository defines the interface for form submission storage
type SubmissionRepository interface {
	repository.Repository[*model.FormSubmission]
	// GetByFormID retrieves all submissions for a form
	GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error)
	// GetByFormIDPaginated retrieves paginated submissions for a form
	GetByFormIDPaginated(ctx context.Context, formID string, params common.PaginationParams) (*common.PaginationResult, error)
	// GetByFormAndUser retrieves a form submission by form ID and user ID
	GetByFormAndUser(ctx context.Context, formID, userID string) (*model.FormSubmission, error)
	// GetSubmissionsByStatus retrieves submissions by status
	GetSubmissionsByStatus(ctx context.Context, status model.SubmissionStatus) ([]*model.FormSubmission, error)
}
