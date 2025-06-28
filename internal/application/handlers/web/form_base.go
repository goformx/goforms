package web

import (
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// FormBaseHandler extends BaseHandler with form-specific functionality
type FormBaseHandler struct {
	*BaseHandler
	FormService   formdomain.Service
	FormValidator *validation.FormValidator
}

// NewFormBaseHandler creates a new form base handler
func NewFormBaseHandler(
	base *BaseHandler,
	formService formdomain.Service,
	formValidator *validation.FormValidator,
) *FormBaseHandler {
	return &FormBaseHandler{
		BaseHandler:   base,
		FormService:   formService,
		FormValidator: formValidator,
	}
}

// GetFormByID retrieves a form by ID with error handling
func (h *FormBaseHandler) GetFormByID(c echo.Context) (*model.Form, error) {
	formID, err := h.FormValidator.ValidateFormID(c)
	if err != nil {
		if handleErr := h.HandleError(c, err, "Invalid form ID"); handleErr != nil {
			h.Logger.Error("failed to handle error", "error", handleErr)
		}

		return nil, echo.NewHTTPError(constants.StatusBadRequest, "Invalid form ID")
	}

	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		h.Logger.Error("failed to get form", "form_id", formID, "error", err)

		if handleErr := h.HandleNotFound(c, "Form not found"); handleErr != nil {
			h.Logger.Error("failed to handle not found", "error", handleErr)
		}

		return nil, echo.NewHTTPError(constants.StatusNotFound, "Form not found")
	}

	return form, nil
}

// RequireFormOwnership verifies the user owns the form
func (h *FormBaseHandler) RequireFormOwnership(c echo.Context, form *model.Form) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		if handleErr := h.HandleForbidden(c, "User not authenticated"); handleErr != nil {
			h.Logger.Error("failed to handle forbidden", "error", handleErr)
		}

		return echo.NewHTTPError(constants.StatusUnauthorized, "User not authenticated")
	}

	if form.UserID != userID {
		h.Logger.Error("ownership verification failed",
			"resource_user_id", form.UserID,
			"request_user_id", userID)

		if handleErr := h.HandleForbidden(c, "You don't have permission to access this resource"); handleErr != nil {
			h.Logger.Error("failed to handle forbidden", "error", handleErr)
		}

		return echo.NewHTTPError(constants.StatusForbidden, "You don't have permission to access this resource")
	}

	return nil
}

// GetFormWithOwnership gets a form and verifies ownership in one call
func (h *FormBaseHandler) GetFormWithOwnership(c echo.Context) (*model.Form, error) {
	form, err := h.GetFormByID(c)
	if err != nil {
		return nil, err
	}

	if ownershipErr := h.RequireFormOwnership(c, form); ownershipErr != nil {
		return nil, ownershipErr
	}

	return form, nil
}
