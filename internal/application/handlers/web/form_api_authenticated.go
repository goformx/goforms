package web

import (
	"github.com/labstack/echo/v4"
)

// Authenticated API endpoints for forms

// GET /api/v1/forms
func (h *FormAPIHandler) handleListForms(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return h.HandleForbidden(c, "User not authenticated")
	}

	// Get forms for the user
	forms, err := h.FormService.ListForms(c.Request().Context(), userID)
	if err != nil {
		h.Logger.Error("failed to list forms", "error", err)

		return h.HandleError(c, err, "Failed to list forms")
	}

	h.Logger.Debug("forms listed successfully",
		"user_id", h.Logger.SanitizeField("user_id", userID),
		"form_count", len(forms))

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildFormListResponse(c, forms); respErr != nil {
		h.Logger.Error("failed to build form list response", "error", respErr)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// GET /api/v1/forms/:id
func (h *FormAPIHandler) handleGetForm(c echo.Context) error {
	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildFormResponse(c, form); respErr != nil {
		h.Logger.Error("failed to build form response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// PUT /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchemaUpdate(c echo.Context) error {
	_, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return h.wrapError("handle ownership error", h.ErrorHandler.HandleOwnershipError(c, err))
	}

	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	// Process and validate schema update request
	schema, err := h.RequestProcessor.ProcessSchemaUpdateRequest(c)
	if err != nil {
		return h.wrapError("handle schema error", h.ErrorHandler.HandleSchemaError(c, err))
	}

	// Update form schema
	if updateErr := h.updateFormSchema(c, form, schema); updateErr != nil {
		return updateErr
	}

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildSchemaResponse(c, form.Schema); respErr != nil {
		h.Logger.Error("failed to build schema response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}
