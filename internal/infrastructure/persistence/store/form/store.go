package form

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/persistence/store/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Store implements form.Repository interface
type Store struct {
	db     *database.GormDB
	logger logging.Logger
}

// NewStore creates a new form store
func NewStore(db *database.GormDB, logger logging.Logger) form.Repository {
	logger.Debug("creating form store",
		logging.Bool("db_available", db != nil),
	)
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form
func (s *Store) Create(ctx context.Context, formModel *model.Form) error {
	if err := s.db.WithContext(ctx).Create(formModel).Error; err != nil {
		s.logger.Error("failed to create form in database",
			logging.String("form_id", formModel.ID),
			logging.Error(err),
		)
		return common.NewDatabaseError("create", "form", formModel.ID, err)
	}
	return nil
}

// GetByID retrieves a form by ID
func (s *Store) GetByID(ctx context.Context, id string) (*model.Form, error) {
	// Normalize the UUID by trimming spaces and converting to lowercase
	normalizedID := strings.TrimSpace(strings.ToLower(id))

	// Validate UUID format
	if _, err := uuid.Parse(normalizedID); err != nil {
		s.logger.Error("invalid form ID format",
			logging.String("form_id", id),
			logging.Error(err),
		)
		return nil, common.NewInvalidInputError("get", "form", id, err)
	}

	s.logger.Debug("getting form by id",
		logging.String("form_id", normalizedID),
	)

	var formModel model.Form
	if err := s.db.WithContext(ctx).Where("uuid = ?", normalizedID).First(&formModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("form not found",
				logging.String("form_id", normalizedID),
			)
			return nil, common.NewNotFoundError("get", "form", normalizedID)
		}
		s.logger.Error("database error while getting form",
			logging.String("form_id", normalizedID),
			logging.Error(err),
		)
		return nil, common.NewDatabaseError("get", "form", normalizedID, err)
	}

	s.logger.Debug("form retrieved successfully",
		logging.String("form_id", formModel.ID),
		logging.String("title", formModel.Title),
	)
	return &formModel, nil
}

// GetByUserID retrieves all forms for a user
func (s *Store) GetByUserID(ctx context.Context, userID uint) ([]*model.Form, error) {
	s.logger.Debug("getting forms by user id",
		logging.Uint("user_id", userID),
	)

	var forms []*model.Form
	result := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&forms)
	if result.Error != nil {
		s.logger.Error("database error while getting user forms",
			logging.Uint("user_id", userID),
			logging.Error(result.Error),
		)
		return nil, common.NewDatabaseError("get_by_user", "form", fmt.Sprintf("user:%d", userID), result.Error)
	}

	s.logger.Debug("successfully retrieved user forms",
		logging.Uint("user_id", userID),
		logging.Int("form_count", len(forms)),
	)

	return forms, nil
}

// Update updates a form
func (s *Store) Update(ctx context.Context, formModel *model.Form) error {
	s.logger.Debug("updating form", logging.String("form_id", formModel.ID))
	result := s.db.WithContext(ctx).Model(&model.Form{}).Where("uuid = ?", formModel.ID).Updates(formModel)
	if result.Error != nil {
		return common.NewDatabaseError("update", "form", formModel.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return common.NewNotFoundError("update", "form", formModel.ID)
	}
	return nil
}

// Delete deletes a form
func (s *Store) Delete(ctx context.Context, id string) error {
	// Normalize the UUID by trimming spaces and converting to lowercase
	normalizedID := strings.TrimSpace(strings.ToLower(id))

	// Validate UUID format
	if _, err := uuid.Parse(normalizedID); err != nil {
		s.logger.Error("invalid form ID format",
			logging.String("form_id", id),
			logging.Error(err),
		)
		return common.NewInvalidInputError("delete", "form", id, err)
	}

	s.logger.Debug("deleting form",
		logging.String("form_id", normalizedID),
	)

	result := s.db.WithContext(ctx).Where("uuid = ?", normalizedID).Delete(&model.Form{})
	if result.Error != nil {
		s.logger.Error("database error while deleting form",
			logging.String("form_id", normalizedID),
			logging.Error(result.Error),
		)
		return common.NewDatabaseError("delete", "form", normalizedID, result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Debug("form not found for deletion",
			logging.String("form_id", normalizedID),
		)
		return common.NewNotFoundError("delete", "form", normalizedID)
	}
	return nil
}

// List returns a paginated list of forms
func (s *Store) List(ctx context.Context, offset, limit int) ([]*model.Form, error) {
	s.logger.Debug("listing forms", logging.Int("offset", offset), logging.Int("limit", limit))
	var forms []*model.Form
	if err := s.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&forms).Error; err != nil {
		return nil, common.NewDatabaseError("list", "form", "", err)
	}
	return forms, nil
}

// Count returns the total number of forms
func (s *Store) Count(ctx context.Context) (int, error) {
	s.logger.Debug("counting forms")
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Form{}).Count(&count).Error; err != nil {
		return 0, common.NewDatabaseError("count", "form", "", err)
	}
	return int(count), nil
}

// Search searches forms by title or description
func (s *Store) Search(ctx context.Context, query string, offset, limit int) ([]*model.Form, error) {
	s.logger.Debug("searching forms",
		logging.String("query", query),
		logging.Int("offset", offset),
		logging.Int("limit", limit),
	)
	var forms []*model.Form
	searchPattern := "%" + query + "%"
	if err := s.db.WithContext(ctx).
		Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&forms).Error; err != nil {
		return nil, common.NewDatabaseError("search", "form", query, err)
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
		return nil, common.NewDatabaseError("get_active", "form", "", err)
	}
	return forms, nil
}

// GetFormsByStatus returns forms by their active status
func (s *Store) GetFormsByStatus(ctx context.Context, active bool) ([]*model.Form, error) {
	s.logger.Debug("getting forms by status", logging.Bool("active", active))
	var forms []*model.Form
	if err := s.db.WithContext(ctx).
		Where("active = ?", active).
		Order("created_at DESC").
		Find(&forms).Error; err != nil {
		return nil, common.NewDatabaseError("get_by_status", "form", fmt.Sprintf("status:%v", active), err)
	}
	return forms, nil
}

// GetFormSubmissions retrieves all submissions for a form
func (s *Store) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	s.logger.Debug("getting form submissions", logging.String("form_id", formID))
	var submissions []*model.FormSubmission
	if err := s.db.WithContext(ctx).
		Where("form_id = ?", formID).
		Order("created_at DESC").
		Find(&submissions).Error; err != nil {
		return nil, common.NewDatabaseError("get_submissions", "form", formID, err)
	}
	return submissions, nil
}
