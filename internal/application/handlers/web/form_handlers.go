// internal/application/handlers/web/form_handlers.go
package web

import (
	"errors"
	"net/http"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/labstack/echo/v4"
)

// HTTP handler methods for form operations
// These methods handle the actual HTTP request/response logic

// handleNew displays the new form creation page
func (h *FormWebHandler) handleNew(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	data := h.BuildPageData(c, "New Form")
	data.User = user
	return h.Renderer.Render(c, pages.NewForm(data))
}

// handleCreate processes form creation requests
func (h *FormWebHandler) handleCreate(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	// Process and validate request
	req, err := h.RequestProcessor.ProcessCreateRequest(c)
	if err != nil {
		return h.ErrorHandler.HandleValidationError(c, err)
	}

	// Create form using business logic service
	form, err := h.FormService.CreateForm(c.Request().Context(), user.ID, req)
	if err != nil {
		return h.handleFormCreationError(c, err)
	}

	return h.ResponseBuilder.BuildSuccessResponse(c, "Form created successfully", map[string]any{
		"form_id": form.ID,
	})
}

// handleEdit displays the form editing page
func (h *FormWebHandler) handleEdit(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	// Log form access for debugging
	h.FormService.LogFormAccess(form)

	data := h.BuildPageData(c, "Edit Form")
	data.User = user
	data.Form = form
	data.FormBuilderAssetPath = h.AssetManager.AssetPath("src/js/form-builder.ts")

	return pages.EditForm(data, form).Render(c.Request().Context(), c.Response().Writer)
}

// handleUpdate processes form update requests
func (h *FormWebHandler) handleUpdate(c echo.Context) error {
	_, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	// Process and validate request
	req, err := h.RequestProcessor.ProcessUpdateRequest(c)
	if err != nil {
		return h.ErrorHandler.HandleValidationError(c, err)
	}

	// Update form using business logic service
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form, req); updateErr != nil {
		h.Logger.Error("failed to update form", "error", updateErr)
		return h.HandleError(c, updateErr, "Failed to update form")
	}

	return h.ResponseBuilder.BuildSuccessResponse(c, "Form updated successfully", map[string]any{
		"form_id": form.ID,
	})
}

// handleDelete processes form deletion requests
func (h *FormWebHandler) handleDelete(c echo.Context) error {
	_, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	if deleteErr := h.FormService.DeleteForm(c.Request().Context(), form.ID); deleteErr != nil {
		h.Logger.Error("failed to delete form", "error", deleteErr)
		return h.HandleError(c, deleteErr, "Failed to delete form")
	}

	return c.NoContent(constants.StatusNoContent)
}

// handleSubmissions displays form submissions
func (h *FormWebHandler) handleSubmissions(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	submissions, err := h.FormService.GetFormSubmissions(c.Request().Context(), form.ID)
	if err != nil {
		h.Logger.Error("failed to get form submissions", "error", err)
		return h.HandleError(c, err, "Failed to get form submissions")
	}

	data := h.BuildPageData(c, "Form Submissions")
	data.User = user
	data.Form = form
	data.Submissions = submissions

	return h.Renderer.Render(c, pages.FormSubmissions(data))
}

// handleFormCreationError handles form creation errors
func (h *FormWebHandler) handleFormCreationError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrFormTitleRequired):
		return h.ResponseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form title is required")
	case errors.Is(err, model.ErrFormSchemaRequired):
		return h.ResponseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form schema is required")
	default:
		return h.ResponseBuilder.BuildErrorResponse(c, http.StatusInternalServerError, "Failed to create form")
	}
}
