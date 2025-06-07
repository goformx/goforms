package formstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/jmoiron/sqlx"
)

// Store implements form.Repository interface
type store struct {
	db     *sqlx.DB
	logger logging.Logger
}

// NewStore creates a new form store
func NewStore(db *sqlx.DB, logger logging.Logger) form.Repository {
	return &store{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form
func (s *store) Create(ctx context.Context, form *model.Form) error {
	query := `
		INSERT INTO forms (
			id, user_id, title, description, schema, active, created_at, updated_at
		) VALUES (
			:id, :user_id, :title, :description, :schema, :active, :created_at, :updated_at
		)
	`

	form.CreatedAt = time.Now()
	form.UpdatedAt = form.CreatedAt

	_, err := s.db.NamedExecContext(ctx, query, form)
	if err != nil {
		s.logger.Error("failed to create form",
			logging.String("operation", "create_form"),
			logging.String("form_id", form.ID),
			logging.Error(err),
		)
		return err
	}

	return nil
}

// GetByID retrieves a form by ID
func (s *store) GetByID(ctx context.Context, id string) (*model.Form, error) {
	query := `
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE id = $1
	`

	var form model.Form
	err := s.db.GetContext(ctx, &form, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.logger.Error("failed to get form by id",
			logging.String("operation", "get_form_by_id"),
			logging.String("form_id", id),
			logging.Error(err),
		)
		return nil, err
	}

	return &form, nil
}

// GetByUserID retrieves forms by user ID
func (s *store) GetByUserID(ctx context.Context, userID uint) ([]*model.Form, error) {
	query := `
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var forms []*model.Form
	err := s.db.SelectContext(ctx, &forms, query, userID)
	if err != nil {
		s.logger.Error("failed to get forms by user id",
			logging.String("operation", "get_forms_by_user_id"),
			logging.String("user_id", fmt.Sprintf("%d", userID)),
			logging.Error(err),
		)
		return nil, err
	}

	return forms, nil
}

// Update updates a form
func (s *store) Update(ctx context.Context, form *model.Form) error {
	query := `
		UPDATE forms
		SET title = :title,
			description = :description,
			schema = :schema,
			active = :active,
			updated_at = :updated_at
		WHERE id = :id
	`

	form.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, form)
	if err != nil {
		s.logger.Error("failed to update form",
			logging.String("operation", "update_form"),
			logging.String("form_id", form.ID),
			logging.Error(err),
		)
		return err
	}

	return nil
}

// Delete deletes a form
func (s *store) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM forms
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		s.logger.Error("failed to delete form",
			logging.String("operation", "delete_form"),
			logging.String("form_id", id),
			logging.Error(err),
		)
		return err
	}

	return nil
}

// GetFormSubmissions retrieves form submissions
func (s *store) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	query := `
		SELECT id, form_id, data, created_at
		FROM form_submissions
		WHERE form_id = $1
		ORDER BY created_at DESC
	`

	var submissions []*model.FormSubmission
	err := s.db.SelectContext(ctx, &submissions, query, formID)
	if err != nil {
		s.logger.Error("failed to get form submissions",
			logging.String("operation", "get_form_submissions"),
			logging.String("form_id", formID),
			logging.Error(err),
		)
		return nil, err
	}

	return submissions, nil
}
