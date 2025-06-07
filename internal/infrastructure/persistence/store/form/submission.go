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

// formSubmissionStore implements form.SubmissionStore interface
type formSubmissionStore struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewSubmissionStore creates a new form submission store
func NewSubmissionStore(db *database.GormDB, logger logging.Logger) form.SubmissionStore {
	logger.Debug("creating form submission store",
		logging.BoolField("db_available", db != nil),
	)
	return &formSubmissionStore{
		db:     db,
		logger: logger,
	}
}

func (s *formSubmissionStore) Create(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("Create called", logging.StringField("form_id", submission.FormID))
	if err := s.db.WithContext(ctx).Create(submission).Error; err != nil {
		return fmt.Errorf("failed to create form submission: %w", err)
	}
	return nil
}

func (s *formSubmissionStore) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	s.logger.Debug("GetByID called", logging.StringField("submission_id", id))
	var submission model.FormSubmission
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrFormNotFound
		}
		return nil, fmt.Errorf("failed to get form submission: %w", err)
	}
	return &submission, nil
}

func (s *formSubmissionStore) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("GetByFormID called", logging.StringField("form_id", formID))
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("submitted_at DESC").
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}
	return submissions, nil
}

func (s *formSubmissionStore) Update(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("Update called", logging.StringField("submission_id", submission.ID))
	result := s.db.WithContext(ctx).Save(submission)
	if result.Error != nil {
		return fmt.Errorf("failed to update form submission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrFormNotFound
	}
	return nil
}

func (s *formSubmissionStore) Delete(ctx context.Context, id string) error {
	s.logger.Debug("Delete called", logging.StringField("submission_id", id))
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.FormSubmission{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete form submission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrFormNotFound
	}
	return nil
}
