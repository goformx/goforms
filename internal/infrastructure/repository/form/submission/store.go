// Package repository provides the form submission repository implementation
package repository

import (
	"context"
	"errors"
	"fmt"

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
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form submission
func (s *Store) Create(ctx context.Context, submission *model.FormSubmission) error {
	if err := s.db.WithContext(ctx).Create(submission).Error; err != nil {
		return fmt.Errorf("failed to create form submission: %w", err)
	}
	return nil
}

// GetByID retrieves a form submission by ID
func (s *Store) GetByID(ctx context.Context, id string) (*model.FormSubmission, error) {
	var submission model.FormSubmission
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get form submission: %w", err)
	}
	return &submission, nil
}

// GetByFormID retrieves all submissions for a specific form
func (s *Store) GetByFormID(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).Where("form_id = ?", formID).Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}
	return submissions, nil
}

// Update updates an existing form submission
func (s *Store) Update(ctx context.Context, submission *model.FormSubmission) error {
	if err := s.db.WithContext(ctx).Save(submission).Error; err != nil {
		return fmt.Errorf("failed to update form submission: %w", err)
	}
	return nil
}

// Delete deletes a form submission by ID
func (s *Store) Delete(ctx context.Context, id string) error {
	if err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.FormSubmission{}).Error; err != nil {
		return fmt.Errorf("failed to delete form submission: %w", err)
	}
	return nil
}

// List retrieves a paginated list of form submissions
func (s *Store) List(ctx context.Context, offset, limit int) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to list form submissions: %w", err)
	}
	return submissions, nil
}

// GetByFormIDPaginated retrieves paginated submissions for a specific form
func (s *Store) GetByFormIDPaginated(ctx context.Context, formID string, params common.PaginationParams) (*common.PaginationResult, error) {
	var submissions []*model.FormSubmission
	var total int64

	// Get total count for this form
	if err := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Where("form_id = ?", formID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count form submissions: %w", err)
	}

	// Get paginated results for this form
	if err := s.db.WithContext(ctx).Where("form_id = ?", formID).Offset(params.GetOffset()).Limit(params.GetLimit()).Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get paginated form submissions: %w", err)
	}

	return &common.PaginationResult{
		Items:      submissions,
		TotalItems: int(total),
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: (int(total) + params.PageSize - 1) / params.PageSize,
	}, nil
}

// CountByFormID counts submissions for a specific form
func (s *Store) CountByFormID(ctx context.Context, formID string) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Where("form_id = ?", formID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count form submissions: %w", err)
	}
	return count, nil
}

// GetByFormIDAndUserID retrieves a submission by form ID and user ID
func (s *Store) GetByFormIDAndUserID(ctx context.Context, formID, userID string) (*model.FormSubmission, error) {
	var submission model.FormSubmission
	if err := s.db.WithContext(ctx).Where("form_id = ? AND user_id = ?", formID, userID).First(&submission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get form submission by form and user: %w", err)
	}
	return &submission, nil
}

// GetByStatus retrieves submissions by status
func (s *Store) GetByStatus(ctx context.Context, status model.SubmissionStatus) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).Where("status = ?", status).Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get form submissions by status: %w", err)
	}
	return submissions, nil
}

// GetActiveSubmissions retrieves active submissions (not deleted)
func (s *Store) GetActiveSubmissions(ctx context.Context, active bool) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	query := s.db.WithContext(ctx)

	if active {
		query = query.Where("deleted_at IS NULL")
	} else {
		query = query.Where("deleted_at IS NOT NULL")
	}

	if err := query.Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get active form submissions: %w", err)
	}
	return submissions, nil
}

// Search searches submissions by query
func (s *Store) Search(ctx context.Context, query string, offset, limit int) ([]*model.FormSubmission, error) {
	var submissions []*model.FormSubmission
	searchQuery := "%" + query + "%"

	if err := s.db.WithContext(ctx).
		Where("data::text ILIKE ? OR status::text ILIKE ?", searchQuery, searchQuery).
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to search form submissions: %w", err)
	}
	return submissions, nil
}

// UpdateStatus updates the status of a form submission
func (s *Store) UpdateStatus(ctx context.Context, id string, status model.SubmissionStatus) error {
	if err := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update form submission status: %w", err)
	}
	return nil
}

// Count returns the total number of form submissions
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.FormSubmission{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count form submissions: %w", err)
	}
	return int(count), nil
}

// CreateSubmission creates a new form submission
func (s *Store) CreateSubmission(ctx context.Context, submission *model.FormSubmission) error {
	return s.Create(ctx, submission)
}

// UpdateSubmission updates an existing form submission
func (s *Store) UpdateSubmission(ctx context.Context, submission *model.FormSubmission) error {
	return s.Update(ctx, submission)
}

// DeleteSubmission deletes a form submission
func (s *Store) DeleteSubmission(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// GetByFormAndUser retrieves a submission by form ID and user ID
func (s *Store) GetByFormAndUser(ctx context.Context, formID, userID string) (*model.FormSubmission, error) {
	return s.GetByFormIDAndUserID(ctx, formID, userID)
}

// GetSubmissionsByStatus retrieves submissions by status
func (s *Store) GetSubmissionsByStatus(ctx context.Context, status model.SubmissionStatus, params common.PaginationParams) (*common.PaginationResult, error) {
	submissions, err := s.GetByStatus(ctx, status)
	if err != nil {
		return nil, err
	}

	// Simple pagination implementation
	start := params.GetOffset()
	end := start + params.GetLimit()
	if start >= len(submissions) {
		return &common.PaginationResult{
			Items:      []*model.FormSubmission{},
			TotalItems: len(submissions),
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: (len(submissions) + params.PageSize - 1) / params.PageSize,
		}, nil
	}

	if end > len(submissions) {
		end = len(submissions)
	}

	return &common.PaginationResult{
		Items:      submissions[start:end],
		TotalItems: len(submissions),
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: (len(submissions) + params.PageSize - 1) / params.PageSize,
	}, nil
}
