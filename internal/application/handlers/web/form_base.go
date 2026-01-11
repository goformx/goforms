package web

import (
	"fmt"

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

// GetFormByID retrieves a form by ID without ownership verification
func (h *FormBaseHandler) GetFormByID(c echo.Context) (*model.Form, error) {
	formID := c.Param("id")
	logger := h.Logger.WithComponent("form_base").With("form_id", formID)

	logger.Debug("getting form by ID")

	form, err := h.FormService.GetForm(c.Request().Context(), formID)
	if err != nil {
		logger.Debug("failed to get form by ID", "error", err)

		return nil, fmt.Errorf("get form by ID: %w", err)
	}

	if form == nil {
		logger.Debug("form not found")

		return nil, fmt.Errorf("get form by ID: %w", h.HandleNotFound(c, "Form not found"))
	}

	logger.Debug("form retrieved successfully", "title", form.Title)

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
