// Package repository provides the form repository implementation
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"encoding/json"
	"time"

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

// FormModel is the infrastructure representation of a form for GORM
// This struct contains all GORM-specific fields and tags
// and is mapped to/from the pure domain Form entity

type FormModel struct {
	ID          string         `gorm:"column:uuid;primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string         `gorm:"not null;index;type:uuid"`
	Title       string         `gorm:"not null;size:100"`
	Description string         `gorm:"size:500"`
	Schema      []byte         `gorm:"type:jsonb;not null"`
	Active      bool           `gorm:"not null;default:true"`
	CreatedAt   time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Status      string         `gorm:"size:20;not null;default:'draft'"`
	CorsOrigins []byte         `gorm:"type:json"`
	CorsMethods []byte         `gorm:"type:json"`
	CorsHeaders []byte         `gorm:"type:json"`
}

func (FormModel) TableName() string { return "forms" }

func (m *FormModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}

	if !m.Active {
		m.Active = true
	}

	if m.Status == "" {
		m.Status = "draft"
	}

	return nil
}

func (m *FormModel) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()

	return nil
}

// Mapper: domain <-> infra
func formModelFromDomain(f *model.Form) (*FormModel, error) {
	if f == nil {
		return nil, fmt.Errorf("form cannot be nil")
	}
	// Marshal JSON fields
	schema, err := json.Marshal(f.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal form schema: %w", err)
	}

	corsOrigins, err := json.Marshal(f.CorsOrigins)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CORS origins: %w", err)
	}

	corsMethods, err := json.Marshal(f.CorsMethods)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CORS methods: %w", err)
	}

	corsHeaders, err := json.Marshal(f.CorsHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CORS headers: %w", err)
	}

	return &FormModel{
		ID:          f.ID,
		UserID:      f.UserID,
		Title:       f.Title,
		Description: f.Description,
		Schema:      schema,
		Active:      f.Active,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
		Status:      f.Status,
		CorsOrigins: corsOrigins,
		CorsMethods: corsMethods,
		CorsHeaders: corsHeaders,
	}, nil
}

func (m *FormModel) ToDomain() (*model.Form, error) {
	if m == nil {
		return nil, fmt.Errorf("form model cannot be nil")
	}

	var schema model.JSON
	if err := json.Unmarshal(m.Schema, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal form schema: %w", err)
	}

	var corsOrigins model.JSON
	if err := json.Unmarshal(m.CorsOrigins, &corsOrigins); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CORS origins: %w", err)
	}

	var corsMethods model.JSON
	if err := json.Unmarshal(m.CorsMethods, &corsMethods); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CORS methods: %w", err)
	}

	var corsHeaders model.JSON
	if err := json.Unmarshal(m.CorsHeaders, &corsHeaders); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CORS headers: %w", err)
	}

	return &model.Form{
		ID:          m.ID,
		UserID:      m.UserID,
		Title:       m.Title,
		Description: m.Description,
		Schema:      schema,
		Active:      m.Active,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Status:      m.Status,
		CorsOrigins: corsOrigins,
		CorsMethods: corsMethods,
		CorsHeaders: corsHeaders,
	}, nil
}

// CreateForm creates a new form
func (s *Store) CreateForm(ctx context.Context, formEntity *model.Form) error {
	formModel, err := formModelFromDomain(formEntity)
	if err != nil {
		return fmt.Errorf("create form: %w", err)
	}

	if createErr := s.db.GetDB().WithContext(ctx).Create(formModel).Error; createErr != nil {
		s.logger.Error("failed to create form",
			"form_id", formEntity.ID,
			"user_id", formEntity.UserID,
			"error", createErr,
		)

		return fmt.Errorf("create form: %w", common.NewDatabaseError("create", "form", formEntity.ID, createErr))
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

		return nil, fmt.Errorf("get form: %w", invalidErr)
	}

	var formModel FormModel
	if err := s.db.GetDB().WithContext(ctx).Where("uuid = ?", normalizedID).First(&formModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("form not found",
				"id_length", len(normalizedID),
				"error_type", "not_found")

			return nil, fmt.Errorf("get form: %w", common.NewNotFoundError("get", "form", normalizedID))
		}

		s.logger.Error("failed to get form",
			"id_length", len(normalizedID),
			"error", err,
			"error_type", "database_error")

		return nil, fmt.Errorf("get form: %w", common.NewDatabaseError("get", "form", normalizedID, err))
	}

	formEntity, err := formModel.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("get form: %w", err)
	}

	s.logger.Debug("form retrieved successfully",
		"id_length", len(normalizedID),
		"form_title", formModel.Title)

	return formEntity, nil
}

// ListForms retrieves all forms for a user
func (s *Store) ListForms(ctx context.Context, userID string) ([]*model.Form, error) {
	var formModels []*FormModel
	if err := s.db.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&formModels).Error; err != nil {
		s.logger.Error("failed to list forms",
			"user_id", userID,
			"error", err,
		)

		return nil, fmt.Errorf("list forms: %w", common.NewDatabaseError("list", "form", "", err))
	}

	forms := make([]*model.Form, len(formModels))
	for i, formModel := range formModels {
		formEntity, err := formModel.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("list forms: %w", err)
		}

		forms[i] = formEntity
	}

	return forms, nil
}

// UpdateForm updates a form
func (s *Store) UpdateForm(ctx context.Context, formEntity *model.Form) error {
	formModel, err := formModelFromDomain(formEntity)
	if err != nil {
		return fmt.Errorf("update form: %w", err)
	}

	result := s.db.GetDB().WithContext(ctx).Model(&FormModel{}).Where("uuid = ?", formEntity.ID).Updates(formModel)
	if result.Error != nil {
		return fmt.Errorf("update form: %w", common.NewDatabaseError("update", "form", formEntity.ID, result.Error))
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("update form: %w", common.NewNotFoundError("update", "form", formEntity.ID))
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

	result := s.db.GetDB().WithContext(ctx).Where("uuid = ?", normalizedID).Delete(&FormModel{})
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
	var formModels []*FormModel
	if err := s.db.GetDB().WithContext(ctx).Where("status = ?", status).Find(&formModels).Error; err != nil {
		return nil, fmt.Errorf("failed to get forms by status: %w", err)
	}

	forms := make([]*model.Form, len(formModels))
	for i, formModel := range formModels {
		formEntity, err := formModel.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("get forms by status: %w", err)
		}

		forms[i] = formEntity
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
