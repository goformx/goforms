package formsubmission

import (
	"context"
	"errors"
	"fmt"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"gorm.io/gorm"
)

var (
	// ErrFormSubmissionNotFound is returned when a form submission cannot be found
	ErrFormSubmissionNotFound = errors.New("form submission not found")
)

// Store implements form.SubmissionStore interface
type Store struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewStore creates a new form submission store
func NewStore(db *database.GormDB, logger logging.Logger) form.SubmissionStore {
	logger.Debug("creating form submission store",
		logging.BoolField("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form submission
func (s *Store) Create(ctx context.Context, submission *model.FormSubmission) error {
	result := s.db.WithContext(ctx).Create(submission)
	if result.Error != nil {
		return fmt.Errorf("failed to insert form submission: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a form submission by its ID
func (s *Store) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	var submission model.FormSubmission
	result := s.db.WithContext(ctx).Where("uuid = ?", id).First(&submission)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrFormSubmissionNotFound
		}
		return nil, fmt.Errorf("failed to get form submission: %w", result.Error)
	}
	return &submission, nil
}

// GetByFormID retrieves all submissions for a form
func (s *Store) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	result := s.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("created_at DESC").
		Find(&submissions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list form submissions: %w", result.Error)
	}
	return submissions, nil
}

// Update updates a form submission
func (s *Store) Update(ctx context.Context, submission *model.FormSubmission) error {
	result := s.db.WithContext(ctx).Save(submission)
	if result.Error != nil {
		return fmt.Errorf("failed to update form submission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("form submission not found: %s", submission.ID)
	}
	return nil
}

// Delete deletes a form submission
func (s *Store) Delete(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Where("uuid = ?", id).Delete(&model.FormSubmission{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete form submission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("form submission not found: %s", id)
	}
	return nil
}
