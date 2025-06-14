// Package repository provides the form submission repository implementation
package repository

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
	"gorm.io/gorm"
)

// Store implements repository.Repository for form submissions
type Store struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewStore creates a new form submission store
func NewStore(db *database.GormDB, logger logging.Logger) form.SubmissionRepository {
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
func (s *Store) List(ctx context.Context, offset, limit int) ([]*model.FormSubmission, error) {
	s.logger.Debug("listing submissions", "offset", offset, "limit", limit)

	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error; err != nil {
		s.logger.Error("failed to list submissions", "error", err)
		return nil, common.NewDatabaseError("list", "form_submission", "", err)
	}
	return submissions, nil
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

// GetSubmissionsByStatus retrieves submissions by status
func (s *Store) GetSubmissionsByStatus(
	ctx context.Context,
	status model.SubmissionStatus,
	params common.PaginationParams,
) (*common.PaginationResult, error) {
	var submissions []*model.FormSubmission
	var total int64

	// Get total count
	if err := s.db.WithContext(ctx).Model(&model.FormSubmission{}).
		Where("status = ?", status).
		Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count submissions: %w", err)
	}

	// Get paginated results
	if err := s.db.WithContext(ctx).
		Where("status = ?", status).
		Offset((params.Page - 1) * params.PageSize).
		Limit(params.PageSize).
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	return &common.PaginationResult{
		Items:      submissions,
		TotalItems: int(total),
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(params.PageSize))),
	}, nil
}

// GetFormsByStatus returns forms by their active status
func (s *Store) GetFormsByStatus(ctx context.Context, active bool) ([]*model.Form, error) {
	s.logger.Debug("getting forms by status", "active", active)

	var forms []*model.Form
	if err := s.db.WithContext(ctx).
		Where("is_active = ?", active).
		Order("created_at DESC").
		Find(&forms).Error; err != nil {
		s.logger.Error("failed to get forms by status", "error", err, "active", active)
		return nil, common.NewDatabaseError("get_by_status", "form", "", err)
	}
	return forms, nil
}

// Count returns the total number of form submissions
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Count(&count).Error; err != nil {
		s.logger.Error("failed to count submissions", "error", err)
		return 0, common.NewDatabaseError("count", "form_submission", "", err)
	}
	return int(count), nil
}

// Search searches form submissions
func (s *Store) Search(ctx context.Context, query string, offset, limit int) ([]*model.FormSubmission, error) {
	s.logger.Debug("searching submissions", "query", query, "offset", offset, "limit", limit)

	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("data::text ILIKE ?", "%"+query+"%").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error; err != nil {
		s.logger.Error("failed to search submissions", "error", err)
		return nil, common.NewDatabaseError("search", "form_submission", "", err)
	}
	return submissions, nil
}

// CreateSubmission creates a new form submission
func (s *Store) CreateSubmission(ctx context.Context, submission *model.FormSubmission) error {
	result := s.db.WithContext(ctx).Create(submission)
	if result.Error != nil {
		return fmt.Errorf("failed to create form submission: %w", result.Error)
	}
	return nil
}

// DeleteSubmission deletes a form submission
func (s *Store) DeleteSubmission(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&model.FormSubmission{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete form submission: %w", result.Error)
	}
	return nil
}

// UpdateSubmission updates a form submission
func (s *Store) UpdateSubmission(ctx context.Context, submission *model.FormSubmission) error {
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
