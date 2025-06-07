package form

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

var (
	ErrFormNotFound = errors.New("form not found")
	ErrFormInvalid  = errors.New("invalid form data")
)

// formStore implements form.Repository interface
type formStore struct {
	db     *database.Database
	logger logging.Logger
}

// NewStore creates a new form store
func NewStore(db *database.Database, logger logging.Logger) form.Repository {
	logger.Debug("creating form store",
		logging.BoolField("db_available", db != nil),
	)
	return &formStore{
		db:     db,
		logger: logger,
	}
}

// Create creates a new form
func (s *formStore) Create(ctx context.Context, f *model.Form) error {
	s.logger.Debug("Create called", logging.StringField("form_id", f.ID))

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`INSERT INTO forms (uuid, user_id, title, description, schema, active, created_at, updated_at) 
		VALUES (%s, %s, %s, %s, %s, %s, %s, %s)`,
		s.db.GetPlaceholder(1),
		s.db.GetPlaceholder(2),
		s.db.GetPlaceholder(3),
		s.db.GetPlaceholder(4),
		s.db.GetPlaceholder(5),
		s.db.GetPlaceholder(6),
		s.db.GetPlaceholder(7),
		s.db.GetPlaceholder(8),
	)

	schemaBytes, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	_, err = tx.ExecContext(ctx, query,
		f.ID,
		f.UserID,
		f.Title,
		f.Description,
		schemaBytes,
		f.Active,
		f.CreatedAt,
		f.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create form: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a form by ID
func (s *formStore) GetByID(ctx context.Context, id string) (*model.Form, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE id = %s
	`, s.db.GetPlaceholder(1))

	var form model.Form
	err := s.db.GetContext(ctx, &form, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFormNotFound
		}
		s.logger.Error("failed to get form by id",
			logging.String("operation", "get_form_by_id"),
			logging.String("form_id", id),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to get form by id: %w", err)
	}

	return &form, nil
}

// GetByUserID retrieves forms by user ID
func (s *formStore) GetByUserID(ctx context.Context, userID uint) ([]*model.Form, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE user_id = %s
		ORDER BY created_at DESC
	`, s.db.GetPlaceholder(1))

	var forms []*model.Form
	err := s.db.SelectContext(ctx, &forms, query, userID)
	if err != nil {
		s.logger.Error("failed to get forms by user id",
			logging.StringField("operation", "get_forms_by_user_id"),
			logging.StringField("user_id", fmt.Sprintf("%d", userID)),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to get forms by user id: %w", err)
	}

	return forms, nil
}

// Update updates a form
func (s *formStore) Update(ctx context.Context, f *model.Form) error {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`
		UPDATE forms
		SET title = %s, description = %s, schema = %s, active = %s, updated_at = %s
		WHERE id = %s
	`,
		s.db.GetPlaceholder(1),
		s.db.GetPlaceholder(2),
		s.db.GetPlaceholder(3),
		s.db.GetPlaceholder(4),
		s.db.GetPlaceholder(5),
		s.db.GetPlaceholder(6),
	)

	schemaBytes, err := json.Marshal(f.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	result, err := tx.ExecContext(ctx, query,
		f.Title,
		f.Description,
		schemaBytes,
		f.Active,
		time.Now(),
		f.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update form: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrFormNotFound
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete deletes a form
func (s *formStore) Delete(ctx context.Context, id string) error {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf("DELETE FROM forms WHERE id = %s", s.db.GetPlaceholder(1))
	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrFormNotFound
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// List returns a paginated list of forms
func (s *formStore) List(ctx context.Context, offset, limit int) ([]*model.Form, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s
	`, s.db.GetPlaceholder(1), s.db.GetPlaceholder(2))

	var forms []*model.Form
	err := s.db.SelectContext(ctx, &forms, query, limit, offset)
	if err != nil {
		s.logger.Error("failed to list forms",
			logging.String("operation", "list_forms"),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to list forms: %w", err)
	}

	return forms, nil
}

// Count returns the total number of forms
func (s *formStore) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM forms")
	if err != nil {
		s.logger.Error("failed to count forms",
			logging.String("operation", "count_forms"),
			logging.Error(err),
		)
		return 0, fmt.Errorf("failed to count forms: %w", err)
	}

	return count, nil
}

// Search searches forms by title or description
func (s *formStore) Search(ctx context.Context, query string, offset, limit int) ([]*model.Form, error) {
	searchQuery := fmt.Sprintf(`
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE title ILIKE %s OR description ILIKE %s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s
	`, s.db.GetPlaceholder(1), s.db.GetPlaceholder(2), s.db.GetPlaceholder(3), s.db.GetPlaceholder(4))

	searchPattern := "%" + query + "%"
	var forms []*model.Form
	err := s.db.SelectContext(ctx, &forms, searchQuery, searchPattern, searchPattern, limit, offset)
	if err != nil {
		s.logger.Error("failed to search forms",
			logging.String("operation", "search_forms"),
			logging.String("query", query),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to search forms: %w", err)
	}

	return forms, nil
}

// GetActiveForms returns all active forms
func (s *formStore) GetActiveForms(ctx context.Context) ([]*model.Form, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE active = %s
		ORDER BY created_at DESC
	`, s.db.GetPlaceholder(1))

	var forms []*model.Form
	err := s.db.SelectContext(ctx, &forms, query, true)
	if err != nil {
		s.logger.Error("failed to get active forms",
			logging.String("operation", "get_active_forms"),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to get active forms: %w", err)
	}

	return forms, nil
}

// GetFormsByStatus returns forms by their active status
func (s *formStore) GetFormsByStatus(ctx context.Context, active bool) ([]*model.Form, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, title, description, schema, active, created_at, updated_at
		FROM forms
		WHERE active = %s
		ORDER BY created_at DESC
	`, s.db.GetPlaceholder(1))

	var forms []*model.Form
	err := s.db.SelectContext(ctx, &forms, query, active)
	if err != nil {
		s.logger.Error("failed to get forms by status",
			logging.StringField("operation", "get_forms_by_status"),
			logging.StringField("active", fmt.Sprintf("%v", active)),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to get forms by status: %w", err)
	}

	return forms, nil
}

// GetFormSubmissions gets all submissions for a form
func (s *formStore) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	query := fmt.Sprintf(`
		SELECT id, form_id, data, created_at
		FROM form_submissions
		WHERE form_id = %s
		ORDER BY created_at DESC
	`, s.db.GetPlaceholder(1))

	var submissions []*model.FormSubmission
	err := s.db.SelectContext(ctx, &submissions, query, formID)
	if err != nil {
		s.logger.Error("failed to get form submissions",
			logging.String("operation", "get_form_submissions"),
			logging.String("form_id", formID),
			logging.Error(err),
		)
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}

	return submissions, nil
}
