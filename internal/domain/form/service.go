package form

import (
	"context"
	"fmt"

	"github.com/goformx/goforms/internal/domain/common/ctxutil"
	domainerrors "github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form/event"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
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
	logger.Debug("creating form service",
		"repo_available", repo != nil,
	)
	return &service{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// CreateForm creates a new form
func (s *service) CreateForm(ctx context.Context, userID string, form *model.Form) error {
	logger := s.logger.WithUserID(userID)

	if err := form.Validate(); err != nil {
		logger.Error("form validation failed", "error", err)
		return domainerrors.New(domainerrors.ErrCodeInvalidInput, "create form: invalid input", err)
	}

	form.UserID = userID

	if err := s.repo.Create(ctx, form); err != nil {
		logger.Error("failed to create form", "error", err)
		return err
	}

	return nil
}

// GetForm retrieves a form by ID
func (s *service) GetForm(ctx context.Context, id string) (*model.Form, error) {
	logger := s.logger

	// Get form
	form, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to get form", "error", err)
		return nil, err
	}

	return form, nil
}

// GetUserForms retrieves all forms for a user
func (s *service) GetUserForms(ctx context.Context, userID string) ([]*model.Form, error) {
	logger := s.logger.WithUserID(userID)

	// Get forms
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

	if validateErr := form.Validate(); validateErr != nil {
		logger.Error("form validation failed", "error", validateErr)
		return domainerrors.New(domainerrors.ErrCodeInvalidInput, "update form: invalid input", validateErr)
	}

	if updateErr := s.repo.Update(ctx, form); updateErr != nil {
		logger.Error("failed to update form", "error", updateErr)
		return updateErr
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

	return nil
}

// GetFormSubmissions returns all submissions for a form
func (s *service) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
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
