package formops

import (
	"strconv"

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
	EnsureFormOwnership(c echo.Context, usr *user.User, formID string) (*form.Form, error)
}

// FormData represents the structure for form creation and updates
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
	var formData FormData
	if err := c.Bind(&formData); err != nil {
		s.logger.Error("failed to bind form data",
			logging.ErrorField("error", err),
			logging.StringField("path", c.Request().URL.Path),
		)
		return nil, errors.New(errors.ErrCodeValidation, "invalid form data", err)
	}

	if err := c.Validate(&formData); err != nil {
		s.logger.Error("form validation failed",
			logging.ErrorField("error", err),
			logging.StringField("path", c.Request().URL.Path),
		)
		return nil, errors.New(errors.ErrCodeValidation, "form validation failed", err)
	}

	return &formData, nil
}

// EnsureFormOwnership ensures the user owns the form
func (s *service) EnsureFormOwnership(c echo.Context, usr *user.User, formID string) (*form.Form, error) {
	frm, err := s.formService.GetForm(formID)
	if err != nil {
		s.logger.Error("failed to get form",
			logging.ErrorField("error", err),
			logging.StringField("form_id", formID),
			logging.StringField("user_id", strconv.FormatUint(uint64(usr.ID), 10)),
		)
		return nil, errors.New(errors.ErrCodeNotFound, "form not found", err)
	}

	if frm.UserID != usr.ID {
		s.logger.Error("user does not own form",
			logging.StringField("form_id", formID),
			logging.StringField("user_id", strconv.FormatUint(uint64(usr.ID), 10)),
		)
		return nil, errors.ErrForbidden
	}

	return frm, nil
}
