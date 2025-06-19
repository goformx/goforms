package web

import (
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/labstack/echo/v4"
)

// FormBaseHandler extends BaseHandler with form-specific functionality
type FormBaseHandler struct {
	*BaseHandler
	FormService formdomain.Service
}

// NewFormBaseHandler creates a new form base handler
func NewFormBaseHandler(base *BaseHandler, formService formdomain.Service) *FormBaseHandler {
	return &FormBaseHandler{
		BaseHandler: base,
		FormService: formService,
	}
}

// GetFormByID retrieves a form by ID with error handling
func (h *FormBaseHandler) GetFormByID(c echo.Context) (*model.Form, error) {
	formID, err := h.ValidateFormID(c)
	if err != nil {
		return nil, h.HandleValidationError(c, err.Error())
	}

	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", "form_id", formID, "error", err)
		return nil, h.HandleNotFound(c, "Form not found")
	}

	return form, nil
}

// RequireFormOwnership verifies the user owns the form
func (h *FormBaseHandler) RequireFormOwnership(c echo.Context, form *model.Form) error {
	return h.ValidateUserOwnership(c, form.UserID)
}

// GetFormWithOwnership gets a form and verifies ownership in one call
func (h *FormBaseHandler) GetFormWithOwnership(c echo.Context) (*model.Form, error) {
	form, err := h.GetFormByID(c)
	if err != nil {
		return nil, err
	}

	if err := h.RequireFormOwnership(c, form); err != nil {
		return nil, err
	}

	return form, nil
}

// HandleFormError handles form-specific errors
func (h *FormBaseHandler) HandleFormError(c echo.Context, err error, message string) error {
	h.Logger.Error("form operation failed", "error", err, "message", message)
	return h.HandleError(c, err, message)
}

// HandleFormValidationError handles form validation errors
func (h *FormBaseHandler) HandleFormValidationError(c echo.Context, message string) error {
	return h.HandleValidationError(c, message)
}
