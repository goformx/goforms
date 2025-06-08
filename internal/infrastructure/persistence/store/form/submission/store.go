package submission

import (
	"context"
	"errors"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/persistence/store/common"
	"gorm.io/gorm"
)

// Store implements form.SubmissionStore interface
type Store struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewStore creates a new form submission store
func NewStore(db *database.GormDB, logger logging.Logger) form.SubmissionStore {
	logger.Debug("creating form submission store",
		logging.Bool("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form submission
func (s *Store) Create(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("creating form submission",
		logging.String("form_id", submission.FormID),
		logging.String("submission_id", submission.ID),
	)

	if err := s.db.WithContext(ctx).Create(submission).Error; err != nil {
		s.logger.Error("failed to create form submission",
			logging.String("form_id", submission.FormID),
			logging.String("submission_id", submission.ID),
			logging.Error(err),
		)
		return common.NewDatabaseError("create", "form_submission", submission.ID, err)
	}
	return nil
}

// GetByID retrieves a form submission by ID
func (s *Store) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	s.logger.Debug("getting form submission by id",
		logging.String("submission_id", id),
	)

	var submission model.FormSubmission
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("form submission not found",
				logging.String("submission_id", id),
			)
			return nil, common.NewNotFoundError("get", "form_submission", id)
		}
		s.logger.Error("database error while getting form submission",
			logging.String("submission_id", id),
			logging.Error(err),
		)
		return nil, common.NewDatabaseError("get", "form_submission", id, err)
	}
	return &submission, nil
}

// GetByFormID retrieves all submissions for a form
func (s *Store) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("getting form submissions by form id",
		logging.String("form_id", formID),
	)

	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("created_at DESC").
		Find(&submissions).Error; err != nil {
		s.logger.Error("database error while getting form submissions",
			logging.String("form_id", formID),
			logging.Error(err),
		)
		return nil, common.NewDatabaseError("get_by_form", "form_submission", formID, err)
	}
	return submissions, nil
}

// Update updates a form submission
func (s *Store) Update(ctx context.Context, submission *model.FormSubmission) error {
	s.logger.Debug("updating form submission",
		logging.String("submission_id", submission.ID),
	)

	result := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Where("id = ?", submission.ID).Updates(submission)
	if result.Error != nil {
		s.logger.Error("database error while updating form submission",
			logging.String("submission_id", submission.ID),
			logging.Error(result.Error),
		)
		return common.NewDatabaseError("update", "form_submission", submission.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Debug("form submission not found for update",
			logging.String("submission_id", submission.ID),
		)
		return common.NewNotFoundError("update", "form_submission", submission.ID)
	}
	return nil
}

// Delete deletes a form submission
func (s *Store) Delete(ctx context.Context, id string) error {
	s.logger.Debug("deleting form submission",
		logging.String("submission_id", id),
	)

	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.FormSubmission{})
	if result.Error != nil {
		s.logger.Error("database error while deleting form submission",
			logging.String("submission_id", id),
			logging.Error(result.Error),
		)
		return common.NewDatabaseError("delete", "form_submission", id, result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Debug("form submission not found for deletion",
			logging.String("submission_id", id),
		)
		return common.NewNotFoundError("delete", "form_submission", id)
	}
	return nil
}

// List returns a paginated list of form submissions
func (s *Store) List(ctx context.Context, params common.PaginationParams) (*common.PaginationResult, error) {
	s.logger.Debug("listing form submissions",
		logging.Int("page", params.Page),
		logging.Int("page_size", params.PageSize),
	)

	var submissions []*model.FormSubmission
	var total int64

	// Get total count
	if err := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Count(&total).Error; err != nil {
		s.logger.Error("database error while counting form submissions",
			logging.Error(err),
		)
		return nil, common.NewDatabaseError("count", "form_submission", "", err)
	}

	// Get paginated results
	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&submissions).Error; err != nil {
		s.logger.Error("database error while listing form submissions",
			logging.Error(err),
		)
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
		logging.String("form_id", formID),
		logging.Int("page", params.Page),
		logging.Int("page_size", params.PageSize),
	)

	var submissions []*model.FormSubmission
	var total int64

	// Get total count for this form
	if err := s.db.WithContext(ctx).
		Model(&model.FormSubmission{}).
		Where("form_id = ?", formID).
		Count(&total).Error; err != nil {
		s.logger.Error("database error while counting form submissions",
			logging.String("form_id", formID),
			logging.Error(err),
		)
		return nil, common.NewDatabaseError("count", "form_submission", formID, err)
	}

	// Get paginated results for this form
	if err := s.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("created_at DESC").
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&submissions).Error; err != nil {
		s.logger.Error("database error while getting paginated form submissions",
			logging.String("form_id", formID),
			logging.Error(err),
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
