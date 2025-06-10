package form

import (
	"context"
	"errors"
	"fmt"

	"github.com/goformx/goforms/internal/domain/common/ctxutil"
	"github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"gorm.io/gorm"
)

type service struct {
	repo      Repository
	publisher event.Publisher
	logger    logging.Logger
}

// NewService creates a new form service instance
func NewService(repo Repository, publisher event.Publisher, logger logging.Logger) Service {
	logger.Debug("creating form service",
		"repo_available", repo != nil,
	)
	return &service{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

func (s *service) CreateForm(
	ctx context.Context,
	userID uint,
	title, description string,
	schema model.JSON,
) (*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Add user ID to context
	ctx = ctxutil.WithUserID(ctx, userID)

	// Log form creation attempt
	s.logger.Debug("creating form",
		"title", title,
		"description", description,
		"user_id", userID,
		"operation", "create_form",
	)

	form := model.NewForm(userID, title, description, schema)

	// Validate form
	if err := form.Validate(); err != nil {
		s.logger.Error("form validation failed",
			"error", err,
			"title", title,
			"description", description,
			"user_id", userID,
			"operation", "create_form",
			"error_type", "validation_error",
			"error_details", fmt.Sprintf("%+v", err),
		)
		return nil, fmt.Errorf("form validation failed: %w", err)
	}

	// Create form in repository
	if err := s.repo.Create(ctx, form); err != nil {
		// Check for specific GORM errors
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			s.logger.Error("form with this title already exists",
				"error", err,
				"title", title,
				"user_id", userID,
				"operation", "create_form",
				"error_type", "duplicate_key",
				"error_details", fmt.Sprintf("%+v", err),
			)
			return nil, fmt.Errorf("form with this title already exists: %w", err)
		case errors.Is(err, gorm.ErrForeignKeyViolated):
			s.logger.Error("invalid user ID",
				"error", err,
				"user_id", userID,
				"operation", "create_form",
				"error_type", "foreign_key_violation",
				"error_details", fmt.Sprintf("%+v", err),
			)
			return nil, fmt.Errorf("invalid user ID: %w", err)
		default:
			s.logger.Error("database error while creating form",
				"error", err,
				"title", title,
				"description", description,
				"user_id", userID,
				"operation", "create_form",
				"error_type", "database_error",
				"error_details", fmt.Sprintf("%+v", err),
			)
			return nil, fmt.Errorf("database error while creating form: %w", err)
		}
	}

	// Publish form created event
	if err := s.publisher.Publish(ctx, event.NewFormCreatedEvent(form)); err != nil {
		s.logger.Error("failed to publish form created event",
			"form_id", form.ID,
			"error", err,
			"operation", "create_form",
			"error_type", "event_publish_error",
			"error_details", fmt.Sprintf("%+v", err),
		)
		// Don't return error here as the form was created successfully
	}

	s.logger.Info("form created successfully",
		"form_id", form.ID,
		"title", form.Title,
		"user_id", form.UserID,
		"operation", "create_form",
	)

	return form, nil
}

func (s *service) GetForm(ctx context.Context, id string) (*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	s.logger.Debug("attempting to get form",
		"form_id", id,
		"operation", "get_form",
	)

	form, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get form from repository",
			"form_id", id,
			"error", err,
			"error_type", "repository_error",
			"error_details", fmt.Sprintf("%+v", err),
			"operation", "get_form",
		)
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	if form == nil {
		s.logger.Debug("form not found",
			"form_id", id,
			"operation", "get_form",
		)
		return nil, model.ErrFormNotFound
	}

	s.logger.Debug("form retrieved successfully",
		"form_id", form.ID,
		"title", form.Title,
		"user_id", form.UserID,
		"operation", "get_form",
	)

	return form, nil
}

// GetUserForms retrieves all forms for a given user
func (s *service) GetUserForms(ctx context.Context, userID uint) ([]*model.Form, error) {
	s.logger.Debug("get user forms request received",
		"operation", "get_user_forms",
		"user_id", userID,
	)

	forms, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("form service failed to get user forms",
			"operation", "get_user_forms",
			"user_id", userID,
			"error", err,
		)
		return nil, err
	}

	s.logger.Debug("user forms retrieved successfully",
		"operation", "get_user_forms",
		"user_id", userID,
		"form_count", len(forms),
	)

	return forms, nil
}

func (s *service) DeleteForm(ctx context.Context, id string) error {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete form",
			"form_id", id,
			"error", err,
			"operation", "delete_form",
			"error_type", "repository_error",
			"error_details", fmt.Sprintf("%+v", err),
		)
		return fmt.Errorf("failed to delete form: %w", err)
	}

	if err := s.publisher.Publish(ctx, event.NewFormDeletedEvent(id)); err != nil {
		s.logger.Error("failed to publish form deleted event",
			"form_id", id,
			"error", err,
			"operation", "delete_form",
			"error_type", "event_publish_error",
			"error_details", fmt.Sprintf("%+v", err),
		)
	}

	return nil
}

func (s *service) UpdateForm(ctx context.Context, form *model.Form) error {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	if form == nil {
		s.logger.Error("invalid form",
			"operation", "update_form",
			"error_type", "validation_error",
			"error_details", "form is nil",
		)
		return model.ErrFormInvalid
	}

	// Add user ID to context
	ctx = ctxutil.WithUserID(ctx, form.UserID)

	if err := form.Validate(); err != nil {
		s.logger.Error("form validation failed",
			"form_id", form.ID,
			"error", err,
			"operation", "update_form",
			"error_type", "validation_error",
			"error_details", fmt.Sprintf("%+v", err),
		)
		return err
	}

	if err := s.repo.Update(ctx, form); err != nil {
		s.logger.Error("failed to update form",
			"form_id", form.ID,
			"error", err,
			"operation", "update_form",
			"error_type", "repository_error",
			"error_details", fmt.Sprintf("%+v", err),
		)
		return fmt.Errorf("failed to update form: %w", err)
	}

	if err := s.publisher.Publish(ctx, event.NewFormUpdatedEvent(form)); err != nil {
		s.logger.Error("failed to publish form updated event",
			"form_id", form.ID,
			"error", err,
			"operation", "update_form",
			"error_type", "event_publish_error",
			"error_details", fmt.Sprintf("%+v", err),
		)
	}

	return nil
}

// GetFormSubmissions returns all submissions for a form
func (s *service) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	s.logger.Debug("getting form submissions",
		"form_id", formID,
		"operation", "get_form_submissions",
	)

	submissions, err := s.repo.GetFormSubmissions(ctx, formID)
	if err != nil {
		s.logger.Error("failed to get form submissions",
			"form_id", formID,
			"error", err,
			"operation", "get_form_submissions",
			"error_type", "repository_error",
			"error_details", fmt.Sprintf("%+v", err),
		)
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}

	s.logger.Debug("form submissions retrieved successfully",
		"form_id", formID,
		"submission_count", len(submissions),
		"operation", "get_form_submissions",
	)

	return submissions, nil
}
