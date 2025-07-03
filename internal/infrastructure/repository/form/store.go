// Package repository provides the form repository implementation
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/repository/common"
)

// Store implements form.Repository interface
type Store struct {
	db     database.DB
	logger logging.Logger
}

// NewStore creates a new form store
func NewStore(db database.DB, logger logging.Logger) form.Repository {
	return &Store{
		db:     db,
		logger: logger,
	}
}

// CreateForm creates a new form
func (s *Store) CreateForm(ctx context.Context, formModel *model.Form) error {
	if err := s.db.GetDB().WithContext(ctx).Create(formModel).Error; err != nil {
		s.logger.Error("failed to create form",
			"form_id", formModel.ID,
			"error", err,
		)

		return fmt.Errorf("create form: %w", common.NewDatabaseError("create", "form", formModel.ID, err))
	}

	return nil
}

// GetFormByID retrieves a form by ID
func (s *Store) GetFormByID(ctx context.Context, id string) (*model.Form, error) {
	// Normalize the UUID by trimming spaces and converting to lowercase
	normalizedID := strings.TrimSpace(strings.ToLower(id))

	// Validate UUID format
	if _, err := uuid.Parse(normalizedID); err != nil {
		s.logger.Warn("invalid form ID format received",
			"id_length", len(id),
			"error_type", "invalid_uuid_format")

		invalidErr := common.NewInvalidInputError("get", "form", id, err)

		return nil, fmt.Errorf("get form by ID: %w", invalidErr)
	}

	var formModel model.Form
	if err := s.db.GetDB().WithContext(ctx).Where("uuid = ?", normalizedID).First(&formModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("form not found in database",
				"id_length", len(normalizedID),
				"error_type", "not_found")

			return nil, fmt.Errorf("get form by ID: %w", common.NewNotFoundError("get", "form", normalizedID))
		}

		s.logger.Error("database error while getting form",
			"id_length", len(normalizedID),
			"error", err,
			"error_type", "database_error")

		dbErr := common.NewDatabaseError("get", "form", normalizedID, err)

		return nil, fmt.Errorf("get form by ID: %w", dbErr)
	}

	s.logger.Debug("form retrieved successfully",
		"id_length", len(normalizedID),
		"form_title", formModel.Title)

	return &formModel, nil
}

// ListForms retrieves all forms for a user
func (s *Store) ListForms(ctx context.Context, userID string) ([]*model.Form, error) {
	var forms []*model.Form
	if err := s.db.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&forms).Error; err != nil {
		s.logger.Error("failed to list forms",
			"user_id", userID,
			"error", err,
		)

		return nil, fmt.Errorf("list forms: %w", common.NewDatabaseError("list", "form", "", err))
	}

	return forms, nil
}

// UpdateForm updates a form
func (s *Store) UpdateForm(ctx context.Context, formModel *model.Form) error {
	result := s.db.GetDB().WithContext(ctx).Model(&model.Form{}).Where("uuid = ?", formModel.ID).Updates(formModel)
	if result.Error != nil {
		return fmt.Errorf("update form: %w", common.NewDatabaseError("update", "form", formModel.ID, result.Error))
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("update form: %w", common.NewNotFoundError("update", "form", formModel.ID))
	}

	return nil
}

// DeleteForm deletes a form
func (s *Store) DeleteForm(ctx context.Context, id string) error {
	// Normalize the UUID by trimming spaces and converting to lowercase
	normalizedID := strings.TrimSpace(strings.ToLower(id))

	// Validate UUID format
	if _, err := uuid.Parse(normalizedID); err != nil {
		s.logger.Warn("invalid form ID format received for deletion",
			"id_length", len(id),
			"error_type", "invalid_uuid_format")

		invalidErr := common.NewInvalidInputError("delete", "form", id, err)

		return fmt.Errorf("delete form: %w", invalidErr)
	}

	result := s.db.GetDB().WithContext(ctx).Where("uuid = ?", normalizedID).Delete(&model.Form{})
	if result.Error != nil {
		s.logger.Error("failed to delete form",
			"id_length", len(normalizedID),
			"error", result.Error,
			"error_type", "database_error")

		return fmt.Errorf("delete form: %w", common.NewDatabaseError("delete", "form", normalizedID, result.Error))
	}

	if result.RowsAffected == 0 {
		s.logger.Debug("form not found for deletion",
			"id_length", len(normalizedID),
			"error_type", "not_found")

		return fmt.Errorf("delete form: %w", common.NewNotFoundError("delete", "form", normalizedID))
	}

	s.logger.Debug("form deleted successfully",
		"id_length", len(normalizedID))

	return nil
}

// GetFormsByStatus returns forms by their active status
func (s *Store) GetFormsByStatus(ctx context.Context, status string) ([]*model.Form, error) {
	var forms []*model.Form
	if err := s.db.GetDB().WithContext(ctx).Where("status = ?", status).Find(&forms).Error; err != nil {
		return nil, fmt.Errorf("failed to get forms by status: %w", err)
	}

	return forms, nil
}

// CreateSubmission creates a new form submission
func (s *Store) CreateSubmission(ctx context.Context, submission *model.FormSubmission) error {
	if err := s.db.GetDB().WithContext(ctx).Create(submission).Error; err != nil {
		s.logger.Error("failed to create form submission",
			"submission_id", submission.ID,
			"form_id", submission.FormID,
			"error", err,
		)

		return fmt.Errorf("create submission: %w", common.NewDatabaseError("create", "form_submission", submission.ID, err))
	}

	return nil
}

// GetSubmissionByID retrieves a form submission by ID
func (s *Store) GetSubmissionByID(ctx context.Context, submissionID string) (*model.FormSubmission, error) {
	var submission model.FormSubmission
	if err := s.db.GetDB().WithContext(ctx).Where("uuid = ?", submissionID).First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get submission by ID: %w",
				common.NewNotFoundError("get", "form_submission", submissionID))
		}

		return nil, fmt.Errorf("get submission by ID: %w",
			common.NewDatabaseError("get", "form_submission", submissionID, err))
	}

	return &submission, nil
}

// ListSubmissions retrieves all submissions for a form
func (s *Store) ListSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	if err := s.db.GetDB().WithContext(ctx).Where("form_id = ?", formID).Find(&submissions).Error; err != nil {
		s.logger.Error("failed to list form submissions",
			"form_id", formID,
			"error", err,
		)

		return nil, fmt.Errorf("list form submissions: %w", common.NewDatabaseError("list", "form_submission", formID, err))
	}

	return submissions, nil
}

// UpdateSubmission updates a form submission
func (s *Store) UpdateSubmission(ctx context.Context, submission *model.FormSubmission) error {
	result := s.db.GetDB().WithContext(ctx).
		Model(&model.FormSubmission{}).
		Where("uuid = ?", submission.ID).
		Updates(submission)
	if result.Error != nil {
		s.logger.Error("failed to update form submission",
			"submission_id", submission.ID,
			"error", result.Error,
		)

		return fmt.Errorf("update submission: %w",
			common.NewDatabaseError("update", "form_submission", submission.ID, result.Error))
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("update submission: %w", common.NewNotFoundError("update", "form_submission", submission.ID))
	}

	return nil
}

// DeleteSubmission deletes a form submission
func (s *Store) DeleteSubmission(ctx context.Context, submissionID string) error {
	result := s.db.GetDB().WithContext(ctx).Where("uuid = ?", submissionID).Delete(&model.FormSubmission{})
	if result.Error != nil {
		s.logger.Error("failed to delete form submission",
			"submission_id", submissionID,
			"error", result.Error,
		)

		return fmt.Errorf("delete submission: %w",
			common.NewDatabaseError("delete", "form_submission", submissionID, result.Error))
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete submission: %w", common.NewNotFoundError("delete", "form_submission", submissionID))
	}

	return nil
}

// GetByFormID retrieves all submissions for a form
func (s *Store) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	return s.ListSubmissions(ctx, formID)
}

// GetByFormIDPaginated retrieves paginated submissions for a form
func (s *Store) GetByFormIDPaginated(
	ctx context.Context,
	formID string,
	params common.PaginationParams,
) (*common.PaginationResult, error) {
	var total int64

	query := s.db.GetDB().WithContext(ctx).Model(&model.FormSubmission{}).Where("form_id = ?", formID)
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count submissions: %w", err)
	}

	var submissions []*model.FormSubmission
	if err := query.
		Offset(params.GetOffset()).
		Limit(params.GetLimit()).
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	return &common.PaginationResult{
		Items:      submissions,
		TotalItems: int(total),
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: (int(total) + params.PageSize - 1) / params.PageSize,
	}, nil
}

// GetByFormAndUser retrieves a submission by form ID and user ID
func (s *Store) GetByFormAndUser(
	ctx context.Context,
	formID string,
	userID string,
) (*model.FormSubmission, error) {
	var submission model.FormSubmission

	query := s.db.GetDB().WithContext(ctx).
		Where("form_id = ? AND user_id = ?", formID, userID).
		First(&submission)
	if err := query.Error; err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	return &submission, nil
}

// GetSubmissionsByStatus retrieves submissions by status
func (s *Store) GetSubmissionsByStatus(
	ctx context.Context,
	status model.SubmissionStatus,
) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	if err := s.db.GetDB().WithContext(ctx).
		Where("status = ?", status).
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	return submissions, nil
}
