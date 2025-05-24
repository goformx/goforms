package form_operations

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Service defines the interface for form operations
type Service interface {
	// ValidateAndBindFormData validates and binds form data from the request
	ValidateAndBindFormData(c echo.Context) (*FormData, error)
	// EnsureFormOwnership ensures the user owns the form
	EnsureFormOwnership(c echo.Context, user *user.User, formID string) (*form.Form, error)
}

// FormData represents the data submitted in a form
type FormData struct {
	Title       string `json:"title" form:"title" validate:"required"`
	Description string `json:"description" form:"description" validate:"required"`
}

// service implements the form operations service
type service struct {
	formService form.Service
	logger      logging.Logger
}

// NewService creates a new form operations service
func NewService(formService form.Service, logger logging.Logger) Service {
	return &service{
		formService: formService,
		logger:      logger,
	}
}

// ValidateAndBindFormData validates and binds form data from the request
func (s *service) ValidateAndBindFormData(c echo.Context) (*FormData, error) {
	var data FormData
	if err := c.Bind(&data); err != nil {
		s.logger.Error("failed to bind form data",
			logging.ErrorField("error", err),
		)
		return nil, err
	}

	// TODO: Add validation logic here

	return &data, nil
}

// EnsureFormOwnership ensures the user owns the form
func (s *service) EnsureFormOwnership(c echo.Context, user *user.User, formID string) (*form.Form, error) {
	form, err := s.formService.GetForm(formID)
	if err != nil {
		s.logger.Error("failed to get form",
			logging.ErrorField("error", err),
			logging.StringField("form_id", formID),
		)
		return nil, err
	}

	if form.UserID != user.ID {
		s.logger.Error("user does not own form",
			logging.StringField("user_id", fmt.Sprintf("%d", user.ID)),
			logging.StringField("form_id", formID),
		)
		return nil, errors.ErrForbidden
	}

	return form, nil
}
