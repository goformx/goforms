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
		logging.String("title", title),
		logging.String("description", description),
		logging.Uint("user_id", userID),
	)

	form := model.NewForm(userID, title, description, schema)

	// Validate form
	if err := form.Validate(); err != nil {
		s.logger.Error("form validation failed",
			logging.Error(err),
			logging.String("title", title),
			logging.String("description", description),
			logging.Uint("user_id", userID),
		)
		return nil, fmt.Errorf("form validation failed: %w", err)
	}

	// Create form in repository
	if err := s.repo.Create(ctx, form); err != nil {
		// Check for specific GORM errors
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			s.logger.Error("form with this title already exists",
				logging.Error(err),
				logging.String("title", title),
				logging.Uint("user_id", userID),
			)
			return nil, fmt.Errorf("form with this title already exists: %w", err)
		case errors.Is(err, gorm.ErrForeignKeyViolated):
			s.logger.Error("invalid user ID",
				logging.Error(err),
				logging.Uint("user_id", userID),
			)
			return nil, fmt.Errorf("invalid user ID: %w", err)
		default:
			s.logger.Error("database error while creating form",
				logging.Error(err),
				logging.String("title", title),
				logging.String("description", description),
				logging.Uint("user_id", userID),
			)
			return nil, fmt.Errorf("database error while creating form: %w", err)
		}
	}

	// Publish form created event
	if err := s.publisher.Publish(ctx, event.NewFormCreatedEvent(form)); err != nil {
		s.logger.Error("failed to publish form created event",
			logging.String("form_id", form.ID),
			logging.Error(err),
		)
		// Don't return error here as the form was created successfully
	}

	s.logger.Info("form created successfully",
		logging.String("form_id", form.ID),
		logging.String("title", form.Title),
		logging.Uint("user_id", form.UserID),
	)

	return form, nil
}

func (s *service) GetForm(ctx context.Context, id string) (*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	s.logger.Debug("attempting to get form",
		logging.String("form_id", id),
		logging.String("operation", "get_form"),
	)

	form, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get form from repository",
			logging.String("form_id", id),
			logging.Error(err),
			logging.String("error_type", "repository_error"),
			logging.String("error_details", fmt.Sprintf("%+v", err)),
			logging.String("operation", "get_form"),
		)
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	if form == nil {
		s.logger.Debug("form not found",
			logging.String("form_id", id),
			logging.String("operation", "get_form"),
		)
		return nil, model.ErrFormNotFound
	}

	s.logger.Debug("form retrieved successfully",
		logging.String("form_id", form.ID),
		logging.String("title", form.Title),
		logging.Uint("user_id", form.UserID),
		logging.String("operation", "get_form"),
	)

	return form, nil
}

func (s *service) GetUserForms(ctx context.Context, userID uint) ([]*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Add user ID to context
	ctx = ctxutil.WithUserID(ctx, userID)

	s.logger.Debug("attempting to get user forms",
		logging.Uint("user_id", userID),
	)

	forms, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		// Log error with minimal context
		s.logger.Error("failed to get user forms",
			logging.Uint("user_id", userID),
			logging.Error(err),
			logging.String("error_message", err.Error()),
			logging.String("error_type", fmt.Sprintf("%T", err)),
		)

		// Return the original error without wrapping to preserve error details
		return nil, err
	}

	s.logger.Debug("user forms retrieved successfully",
		logging.Uint("user_id", userID),
		logging.Int("form_count", len(forms)),
	)

	return forms, nil
}

func (s *service) DeleteForm(ctx context.Context, id string) error {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete form: %w", err)
	}

	if err := s.publisher.Publish(ctx, event.NewFormDeletedEvent(id)); err != nil {
		s.logger.Error("failed to publish form deleted event",
			logging.String("form_id", id),
			logging.Error(err))
	}

	return nil
}

func (s *service) UpdateForm(ctx context.Context, form *model.Form) error {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	if form == nil {
		return model.ErrFormInvalid
	}

	// Add user ID to context
	ctx = ctxutil.WithUserID(ctx, form.UserID)

	if err := form.Validate(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, form); err != nil {
		return fmt.Errorf("failed to update form: %w", err)
	}

	if err := s.publisher.Publish(ctx, event.NewFormUpdatedEvent(form)); err != nil {
		s.logger.Error("failed to publish form updated event",
			logging.String("form_id", form.ID),
			logging.Error(err))
	}

	return nil
}

// GetFormSubmissions returns all submissions for a form
func (s *service) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	return s.repo.GetFormSubmissions(ctx, formID)
}
