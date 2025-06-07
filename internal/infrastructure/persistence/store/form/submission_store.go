package form

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
	// ErrSubmissionNotFound is returned when a form submission cannot be found
	ErrSubmissionNotFound = errors.New("form submission not found")
)

// FormSubmissionStore implements form.SubmissionStore
type FormSubmissionStore struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewFormSubmissionStore creates a new form submission store
func NewFormSubmissionStore(db *database.GormDB, logger logging.Logger) form.SubmissionStore {
	return &FormSubmissionStore{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form submission
func (s *FormSubmissionStore) Create(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("creating form submission", logging.StringField("form_id", submission.FormID))
	if err := s.db.WithContext(ctx).Create(submission).Error; err != nil {
		return fmt.Errorf("failed to create form submission: %w", err)
	}
	return nil
}

// GetByID retrieves a form submission by its ID
func (s *FormSubmissionStore) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	s.logger.Debug("getting form submission by id", logging.StringField("submission_id", id))
	var submission model.FormSubmission
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("failed to get form submission: %w", err)
	}
	return &submission, nil
}

// GetByFormID retrieves all submissions for a specific form
func (s *FormSubmissionStore) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("getting form submissions by form id", logging.StringField("form_id", formID))
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("created_at DESC").
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}
	return submissions, nil
}

// GetByUserID retrieves all submissions made by a specific user
func (s *FormSubmissionStore) GetByUserID(ctx context.Context, userID uint) ([]*model.FormSubmission, error) {
	s.logger.Debug("getting form submissions by user id", logging.UintField("user_id", userID))
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user submissions: %w", err)
	}
	return submissions, nil
}

// Update updates an existing form submission
func (s *FormSubmissionStore) Update(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("updating form submission", logging.StringField("submission_id", submission.ID))
	result := s.db.WithContext(ctx).Save(submission)
	if result.Error != nil {
		return fmt.Errorf("failed to update form submission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSubmissionNotFound
	}
	return nil
}

// Delete deletes a form submission by its ID
func (s *FormSubmissionStore) Delete(ctx context.Context, id string) error {
	s.logger.Debug("deleting form submission", logging.StringField("submission_id", id))
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.FormSubmission{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete form submission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSubmissionNotFound
	}
	return nil
}

// List retrieves a paginated list of form submissions
func (s *FormSubmissionStore) List(ctx context.Context, offset, limit int) ([]*model.FormSubmission, error) {
	s.logger.Debug("listing form submissions",
		logging.IntField("offset", offset),
		logging.IntField("limit", limit),
	)
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to list form submissions: %w", err)
	}
	return submissions, nil
}

// Count returns the total number of form submissions
func (s *FormSubmissionStore) Count(ctx context.Context) (int, error) {
	s.logger.Debug("counting form submissions")
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&model.FormSubmission{}).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count form submissions: %w", err)
	}
	return int(count), nil
}

// Search searches form submissions by form ID and user ID
func (s *FormSubmissionStore) Search(ctx context.Context, formID string, userID uint, offset, limit int) ([]*model.FormSubmission, error) {
	s.logger.Debug("searching form submissions",
		logging.StringField("form_id", formID),
		logging.UintField("user_id", userID),
		logging.IntField("offset", offset),
		logging.IntField("limit", limit),
	)
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("form_id = ? AND user_id = ?", formID, userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to search form submissions: %w", err)
	}
	return submissions, nil
}
