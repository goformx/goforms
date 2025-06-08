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

// ErrFormNotFound is returned when a form cannot be found
var ErrFormNotFound = errors.New("form not found")

// Store implements form.Repository interface
type Store struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewStore creates a new form store
func NewStore(db *database.GormDB, logger logging.Logger) form.Repository {
	logger.Debug("creating form store",
		logging.BoolField("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form
func (s *Store) Create(ctx context.Context, formModel *model.Form) error {
	s.logger.Debug("creating form", logging.StringField("form_id", formModel.ID))
	if err := s.db.WithContext(ctx).Create(formModel).Error; err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}
	return nil
}

// GetByID retrieves a form by ID
func (s *Store) GetByID(ctx context.Context, id string) (*model.Form, error) {
	s.logger.Debug("getting form by id", logging.StringField("form_id", id))
	var formModel model.Form
	if err := s.db.WithContext(ctx).Where("uuid = ?", id).First(&formModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFormNotFound
		}
		return nil, fmt.Errorf("failed to get form: %w", err)
	}
	return &formModel, nil
}

// GetByUserID retrieves all forms created by a specific user
func (s *Store) GetByUserID(ctx context.Context, userID uint) ([]*model.Form, error) {
	s.logger.Debug("getting forms by user id", logging.UintField("user_id", userID))
	var forms []*model.Form
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&forms).Error; err != nil {
		return nil, fmt.Errorf("failed to get forms by user ID: %w", err)
	}
	return forms, nil
}

// Update updates a form
func (s *Store) Update(ctx context.Context, formModel *model.Form) error {
	s.logger.Debug("updating form", logging.StringField("form_id", formModel.ID))
	result := s.db.WithContext(ctx).Model(&model.Form{}).Where("uuid = ?", formModel.ID).Updates(formModel)
	if result.Error != nil {
		return fmt.Errorf("failed to update form: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFormNotFound
	}
	return nil
}

// Delete deletes a form
func (s *Store) Delete(ctx context.Context, id string) error {
	s.logger.Debug("deleting form", logging.StringField("form_id", id))
	result := s.db.WithContext(ctx).Where("uuid = ?", id).Delete(&model.Form{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete form: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFormNotFound
	}
	return nil
}

// List returns a paginated list of forms
func (s *Store) List(ctx context.Context, offset, limit int) ([]*model.Form, error) {
	s.logger.Debug("listing forms", logging.IntField("offset", offset), logging.IntField("limit", limit))
	var forms []*model.Form
	if err := s.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&forms).Error; err != nil {
		return nil, fmt.Errorf("failed to list forms: %w", err)
	}
	return forms, nil
}

// Count returns the total number of forms
func (s *Store) Count(ctx context.Context) (int, error) {
	s.logger.Debug("counting forms")
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Form{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count forms: %w", err)
	}
	return int(count), nil
}

// Search searches forms by title or description
func (s *Store) Search(ctx context.Context, query string, offset, limit int) ([]*model.Form, error) {
	s.logger.Debug("searching forms",
		logging.StringField("query", query),
		logging.IntField("offset", offset),
		logging.IntField("limit", limit),
	)
	var forms []*model.Form
	searchPattern := "%" + query + "%"
	if err := s.db.WithContext(ctx).
		Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&forms).Error; err != nil {
		return nil, fmt.Errorf("failed to search forms: %w", err)
	}
	return forms, nil
}

// GetActiveForms returns all active forms
func (s *Store) GetActiveForms(ctx context.Context) ([]*model.Form, error) {
	s.logger.Debug("getting active forms")
	var forms []*model.Form
	if err := s.db.WithContext(ctx).
		Where("active = ?", true).
		Order("created_at DESC").
		Find(&forms).Error; err != nil {
		return nil, fmt.Errorf("failed to get active forms: %w", err)
	}
	return forms, nil
}

// GetFormsByStatus returns forms by their active status
func (s *Store) GetFormsByStatus(ctx context.Context, active bool) ([]*model.Form, error) {
	s.logger.Debug("getting forms by status", logging.BoolField("active", active))
	var forms []*model.Form
	if err := s.db.WithContext(ctx).
		Where("active = ?", active).
		Order("created_at DESC").
		Find(&forms).Error; err != nil {
		return nil, fmt.Errorf("failed to get forms by status: %w", err)
	}
	return forms, nil
}

// GetFormSubmissions gets all submissions for a form
func (s *Store) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("getting form submissions", logging.StringField("form_id", formID))
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("created_at DESC").
		Find(&submissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}
	return submissions, nil
}
