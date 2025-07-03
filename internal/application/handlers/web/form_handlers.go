// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
)

// HTTP handler methods for form operations
// These methods handle the actual HTTP request/response logic

// handleNew displays the new form creation page
func (h *FormWebHandler) handleNew(c echo.Context) error {
	user, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	data := h.NewPageData(c, "New Form")
	data.SetUser(user)

	return fmt.Errorf("render new form: %w", h.Renderer.Render(c, pages.NewForm(*data)))
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
		return fmt.Errorf("handle error: %w", h.ErrorHandler.HandleError(c, err))
	}

	// Create form using business logic service
	form, err := h.FormService.CreateForm(c.Request().Context(), user.ID, req)
	if err != nil {
		return h.handleFormCreationError(c, err)
	}

	return fmt.Errorf("build success response: %w",
		h.ResponseBuilder.BuildSuccessResponse(c, "Form created successfully", map[string]any{
			"form_id": form.ID,
		}))
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

	data := h.NewPageData(c, "Edit Form").
		WithForm(form).
		WithFormBuilderAssetPath(h.AssetManager.AssetPath("src/js/pages/form-builder.ts"))

	data.SetUser(user)

	return fmt.Errorf("render edit form: %w", h.Renderer.Render(c, pages.EditForm(*data, form)))
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
		return fmt.Errorf("handle error: %w", h.ErrorHandler.HandleError(c, err))
	}

	// Update form using business logic service
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form, req); updateErr != nil {
		h.Logger.Error("failed to update form", "error", updateErr)

		return h.HandleError(c, updateErr, "Failed to update form")
	}

	return fmt.Errorf("build success response: %w",
		h.ResponseBuilder.BuildSuccessResponse(c, "Form updated successfully", map[string]any{
			"form_id": form.ID,
		}))
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

	return fmt.Errorf("no content response: %w", c.NoContent(constants.StatusNoContent))
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

	data := h.NewPageData(c, "Form Submissions").
		WithForm(form).
		WithSubmissions(submissions)

	data.SetUser(user)

	return fmt.Errorf("render submissions: %w", h.Renderer.Render(c, pages.FormSubmissions(*data)))
}

// handleFormCreationError handles form creation errors
func (h *FormWebHandler) handleFormCreationError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrFormTitleRequired):
		return fmt.Errorf("build error response: %w",
			h.ResponseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form title is required"))
	case errors.Is(err, model.ErrFormSchemaRequired):
		return fmt.Errorf("build error response: %w",
			h.ResponseBuilder.BuildErrorResponse(c, http.StatusBadRequest, "Form schema is required"))
	default:
		return fmt.Errorf("build error response: %w",
			h.ResponseBuilder.BuildErrorResponse(c, http.StatusInternalServerError, "Failed to create form"))
	}
}
