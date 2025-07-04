// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

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

	if renderErr := h.Renderer.Render(c, pages.NewForm(*data)); renderErr != nil {
		return fmt.Errorf("failed to render new form page: %w", renderErr)
	}

	return nil
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
		return h.HandleError(c, err, "Failed to create form")
	}

	// Create form using business logic service
	form, err := h.FormService.CreateForm(c.Request().Context(), user.ID, req)
	if err != nil {
		return h.HandleError(c, err, "Failed to create form")
	}

	// Redirect to the edit page for the new form
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/forms/%s/edit", form.ID))
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

	if renderErr := h.Renderer.Render(c, pages.EditForm(*data, form)); renderErr != nil {
		return fmt.Errorf("failed to render edit form page: %w", renderErr)
	}

	return nil
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
		return h.HandleError(c, err, "Failed to update form")
	}

	// Update form using business logic service
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form, req); updateErr != nil {
		h.Logger.Error("failed to update form", "error", updateErr)
		return h.HandleError(c, updateErr, "Failed to update form")
	}

	// Redirect back to the edit page
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/forms/%s/edit", form.ID))
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

	// Redirect to dashboard after deletion
	return c.Redirect(http.StatusSeeOther, "/dashboard")
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

	if renderErr := h.Renderer.Render(c, pages.FormSubmissions(*data)); renderErr != nil {
		return fmt.Errorf("failed to render form submissions page: %w", renderErr)
	}

	return nil
}

// Note: handleFormCreationError removed - using standard error handling
