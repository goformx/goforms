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
		logging.StringField("title", title),
		logging.StringField("description", description),
		logging.UintField("user_id", userID),
	)

	form := model.NewForm(userID, title, description, schema)

	// Validate form
	if err := form.Validate(); err != nil {
		s.logger.Error("form validation failed",
			logging.ErrorField("error", err),
			logging.StringField("title", title),
			logging.StringField("description", description),
			logging.UintField("user_id", userID),
		)
		return nil, fmt.Errorf("form validation failed: %w", err)
	}

	// Create form in repository
	if err := s.repo.Create(ctx, form); err != nil {
		// Check for specific GORM errors
		switch {
		case errors.Is(err, gorm.ErrDuplicatedKey):
			s.logger.Error("form with this title already exists",
				logging.ErrorField("error", err),
				logging.StringField("title", title),
				logging.UintField("user_id", userID),
			)
			return nil, fmt.Errorf("form with this title already exists: %w", err)
		case errors.Is(err, gorm.ErrForeignKeyViolated):
			s.logger.Error("invalid user ID",
				logging.ErrorField("error", err),
				logging.UintField("user_id", userID),
			)
			return nil, fmt.Errorf("invalid user ID: %w", err)
		default:
			s.logger.Error("database error while creating form",
				logging.ErrorField("error", err),
				logging.StringField("title", title),
				logging.StringField("description", description),
				logging.UintField("user_id", userID),
			)
			return nil, fmt.Errorf("database error while creating form: %w", err)
		}
	}

	// Publish form created event
	if err := s.publisher.Publish(ctx, event.NewFormCreatedEvent(form)); err != nil {
		s.logger.Error("failed to publish form created event",
			logging.StringField("form_id", form.ID),
			logging.ErrorField("error", err),
		)
		// Don't return error here as the form was created successfully
	}

	s.logger.Info("form created successfully",
		logging.StringField("form_id", form.ID),
		logging.StringField("title", form.Title),
		logging.UintField("user_id", form.UserID),
	)

	return form, nil
}

func (s *service) GetForm(ctx context.Context, id string) (*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	form, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if form == nil {
		return nil, model.ErrFormNotFound
	}

	return form, nil
}

func (s *service) GetUserForms(ctx context.Context, userID uint) ([]*model.Form, error) {
	// Add timeout to context
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	// Add user ID to context
	ctx = ctxutil.WithUserID(ctx, userID)

	forms, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

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
			logging.StringField("form_id", id),
			logging.ErrorField("error", err))
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
			logging.StringField("form_id", form.ID),
			logging.ErrorField("error", err))
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
