package form

import (
	"context"
	"fmt"

	"github.com/goformx/goforms/internal/domain/common/ctxutil"
	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/mrz1836/go-sanitize"
)

type Service interface {
	CreateForm(ctx context.Context, userID string, form *model.Form) error
	GetForm(ctx context.Context, id string) (*model.Form, error)
	GetUserForms(ctx context.Context, userID string) ([]*model.Form, error)
	UpdateForm(ctx context.Context, userID string, form *model.Form) error
	DeleteForm(ctx context.Context, userID, id string) error
	GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error)
}

type service struct {
	repo      Repository
	publisher event.Publisher
	logger    logging.Logger
}

// NewService creates a new form service instance
func NewService(repo Repository, publisher event.Publisher, logger logging.Logger) *service {
	return &service{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// CreateForm creates a new form
func (s *service) CreateForm(ctx context.Context, userID string, form *model.Form) error {
	logger := s.logger.WithUserID(userID)

	// Sanitize form data before validation
	form.Title = sanitize.XSS(form.Title)
	form.Description = sanitize.XSS(form.Description)

	if err := form.Validate(); err != nil {
		logger.Error("form validation failed", "error", err)
		return domainerrors.New(domainerrors.ErrCodeInvalidInput, "create form: invalid input", err)
	}

	form.UserID = userID

	if err := s.repo.Create(ctx, form); err != nil {
		logger.Error("failed to create form", "error", err)
		return err
	}

	// Publish form created event
	if pubErr := s.publisher.Publish(ctx, event.NewFormCreatedEvent(form)); pubErr != nil {
		logger.Error("failed to publish form created event", "error", pubErr)
	}

	return nil
}

// GetForm retrieves a form by ID
func (s *service) GetForm(ctx context.Context, id string) (*model.Form, error) {
	form, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get form", "error", err)
		return nil, err
	}

	return form, nil
}

// GetUserForms retrieves all forms for a user
func (s *service) GetUserForms(ctx context.Context, userID string) ([]*model.Form, error) {
	logger := s.logger.WithUserID(userID)

	forms, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("failed to get user forms", "error", err)
		return nil, err
	}

	return forms, nil
}

// UpdateForm updates a form
func (s *service) UpdateForm(ctx context.Context, userID string, form *model.Form) error {
	logger := s.logger.WithUserID(userID)

	existingForm, err := s.repo.GetByID(ctx, form.ID)
	if err != nil {
		logger.Error("failed to get existing form", "error", err)
		return err
	}

	if existingForm.UserID != userID {
		logger.Error("user does not own form", "user_id", userID, "form_id", form.ID)
		return domainerrors.New(domainerrors.ErrCodeForbidden, "update form: user does not own form", nil)
	}

	// Sanitize form data before validation
	form.Title = sanitize.XSS(form.Title)
	form.Description = sanitize.XSS(form.Description)

	if validateErr := form.Validate(); validateErr != nil {
		logger.Error("form validation failed", "error", validateErr)
		return domainerrors.New(domainerrors.ErrCodeInvalidInput, "update form: invalid input", validateErr)
	}

	if updateErr := s.repo.Update(ctx, form); updateErr != nil {
		logger.Error("failed to update form", "error", updateErr)
		return updateErr
	}

	// Publish form updated event
	if pubErr := s.publisher.Publish(ctx, event.NewFormUpdatedEvent(form)); pubErr != nil {
		logger.Error("failed to publish form updated event", "error", pubErr)
	}

	return nil
}

// DeleteForm deletes a form
func (s *service) DeleteForm(ctx context.Context, userID, id string) error {
	logger := s.logger.WithUserID(userID)

	form, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to get form", "error", err)
		return err
	}

	if form.UserID != userID {
		logger.Error("user does not own form", "user_id", userID, "form_id", id)
		return domainerrors.New(domainerrors.ErrCodeForbidden, "delete form: user does not own form", nil)
	}

	if deleteErr := s.repo.Delete(ctx, id); deleteErr != nil {
		logger.Error("failed to delete form", "error", deleteErr)
		return deleteErr
	}

	// Publish form deleted event
	if pubErr := s.publisher.Publish(ctx, event.NewFormDeletedEvent(id)); pubErr != nil {
		logger.Error("failed to publish form deleted event", "error", pubErr)
	}

	return nil
}

// GetFormSubmissions returns all submissions for a form
func (s *service) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	ctx, cancel := ctxutil.WithDefaultTimeout(ctx)
	defer cancel()

	submissions, err := s.repo.GetFormSubmissions(ctx, formID)
	if err != nil {
		s.logger.Error("failed to get form submissions", "error", err)
		return nil, fmt.Errorf("failed to get form submissions: %w", err)
	}

	return submissions, nil
}
