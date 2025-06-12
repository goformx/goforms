package submission

import (
	"context"
	"errors"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
	"gorm.io/gorm"
)

// Store implements form.SubmissionStore interface
type Store struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewStore creates a new form submission store
func NewStore(db *database.GormDB, logger logging.Logger) form.SubmissionStore {
	logger.Debug("submission store initialized", "service", "submission")
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form submission
func (s *Store) Create(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("creating submission", "form_id", submission.FormID, "id", submission.ID)

	if err := s.db.WithContext(ctx).Create(submission).Error; err != nil {
		s.logger.Error("failed to create submission", "error", err, "form_id", submission.FormID)
		return common.NewDatabaseError("create", "form_submission", submission.ID, err)
	}
	s.logger.Debug("submission created", "id", submission.ID, "form_id", submission.FormID)
	return nil
}

// GetByID retrieves a form submission by ID
func (s *Store) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	s.logger.Debug("getting submission", "id", id)

	var submission model.FormSubmission
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("form submission not found",
				"submission_id", id,
			)
			return nil, common.NewNotFoundError("get", "form_submission", id)
		}
		s.logger.Error("failed to get submission", "error", err, "id", id)
		return nil, common.NewDatabaseError("get", "form_submission", id, err)
	}
	s.logger.Debug("submission found", "id", id)
	return &submission, nil
}

// GetByFormID retrieves all submissions for a form
func (s *Store) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("getting form submissions", "form_id", formID)

	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("form_uuid = ?", formID).
		Order("created_at DESC").
		Find(&submissions).Error; err != nil {
		s.logger.Error("failed to get form submissions", "error", err, "form_id", formID)
		return nil, common.NewDatabaseError("get_by_form", "form_submission", formID, err)
	}
	return submissions, nil
}

// Update updates a form submission
func (s *Store) Update(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("updating submission", "id", submission.ID)

	result := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Where("id = ?", submission.ID).Updates(submission)
	if result.Error != nil {
		s.logger.Error("failed to update submission", "error", result.Error, "id", submission.ID)
		return common.NewDatabaseError("update", "form_submission", submission.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Debug("form submission not found for update",
			"submission_id", submission.ID,
		)
		return common.NewNotFoundError("update", "form_submission", submission.ID)
	}
	s.logger.Debug("submission updated", "id", submission.ID)
	return nil
}

// Delete deletes a form submission
func (s *Store) Delete(ctx context.Context, id string) error {
	s.logger.Debug("deleting submission", "id", id)

	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.FormSubmission{})
	if result.Error != nil {
		s.logger.Error("failed to delete submission", "error", result.Error, "id", id)
		return common.NewDatabaseError("delete", "form_submission", id, result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Debug("form submission not found for deletion",
			"submission_id", id,
		)
		return common.NewNotFoundError("delete", "form_submission", id)
	}
	s.logger.Debug("submission deleted", "id", id)
	return nil
}

// List returns a paginated list of form submissions
func (s *Store) List(ctx context.Context, params common.PaginationParams) (*common.PaginationResult, error) {
	s.logger.Debug("listing submissions", "page", params.Page, "page_size", params.PageSize)

	var submissions []*model.FormSubmission
	var total int64

	// Get total count
	if err := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Count(&total).Error; err != nil {
		s.logger.Error("failed to count submissions", "error", err)
		return nil, common.NewDatabaseError("count", "form_submission", "", err)
	}

	// Get paginated results
	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&submissions).Error; err != nil {
		s.logger.Error("failed to list submissions", "error", err)
		return nil, common.NewDatabaseError("list", "form_submission", "", err)
	}

	return &common.PaginationResult{
		Items:      submissions,
		TotalItems: int(total),
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: (int(total) + params.PageSize - 1) / params.PageSize,
	}, nil
}

// GetByFormIDPaginated returns a paginated list of submissions for a form
func (s *Store) GetByFormIDPaginated(
	ctx context.Context,
	formID string,
	params common.PaginationParams,
) (*common.PaginationResult, error) {
	s.logger.Debug("getting paginated form submissions",
		"form_id", formID,
		"page", params.Page,
		"page_size", params.PageSize,
	)

	var submissions []*model.FormSubmission
	var total int64

	// Get total count for this form
	if err := s.db.WithContext(ctx).
		Model(&model.FormSubmission{}).
		Where("form_uuid = ?", formID).
		Count(&total).Error; err != nil {
		s.logger.Error("database error while counting form submissions",
			"form_id", formID,
			"error", err,
		)
		return nil, common.NewDatabaseError("count", "form_submission", formID, err)
	}

	// Get paginated results for this form
	if err := s.db.WithContext(ctx).
		Where("form_uuid = ?", formID).
		Order("created_at DESC").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&submissions).Error; err != nil {
		s.logger.Error("database error while getting paginated form submissions",
			"form_id", formID,
			"error", err,
		)
		return nil, common.NewDatabaseError("get_by_form_paginated", "form_submission", formID, err)
	}

	return &common.PaginationResult{
		Items:      submissions,
		TotalItems: int(total),
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: (int(total) + params.PageSize - 1) / params.PageSize,
	}, nil
}

// GetByFormAndUser retrieves a form submission by form ID and user ID
func (s *Store) GetByFormAndUser(ctx context.Context, formID, userID string) (*model.FormSubmission, error) {
	s.logger.Debug("getting submission by form and user", "form_id", formID, "user_id", userID)

	var submission model.FormSubmission
	query := s.db.WithContext(ctx).Where("form_uuid = ? AND user_id = ?", formID, userID)
	if err := query.First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("form submission not found by form and user",
				"form_id", formID,
				"user_id", userID,
			)
			return nil, common.NewNotFoundError("get", "form_submission", formID)
		}
		s.logger.Error("failed to get submission by form and user", "error", err, "form_id", formID, "user_id", userID)
		return nil, common.NewDatabaseError("get", "form_submission", formID, err)
	}
	s.logger.Debug("submission found by form and user", "form_id", formID, "user_id", userID)
	return &submission, nil
}
