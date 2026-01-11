// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/romsar/gonertia"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/presentation/inertia"
)

// HTTP handler methods for form operations
// These methods handle the actual HTTP request/response logic

// handleNew displays the new form creation page
func (h *FormWebHandler) handleNew(c echo.Context) error {
	_, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	return h.Inertia.Render(c, "Forms/New", inertia.Props{
		"title": "Create New Form",
	})
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

	// For AJAX/Inertia requests, return JSON so the frontend can handle the redirect
	if c.Request().Header.Get(constants.HeaderXRequestedWith) == "XMLHttpRequest" || gonertia.IsInertiaRequest(c.Request()) {
		return h.ResponseBuilder.BuildSuccessResponse(c, "Form created successfully", map[string]any{
			"form_id": form.ID,
		})
	}

	// For regular form submissions, redirect to the edit page
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/forms/%s/edit", form.ID))
}

// handleEdit displays the form editing page
func (h *FormWebHandler) handleEdit(c echo.Context) error {
	_, err := h.AuthHelper.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.AuthHelper.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	// Log form access for debugging
	h.FormService.LogFormAccess(form)

	return h.Inertia.Render(c, "Forms/Edit", inertia.Props{
		"title": "Edit Form",
		"form": map[string]any{
			"id":          form.ID,
			"title":       form.Title,
			"description": form.Description,
			"status":      form.Status,
			"corsOrigins": form.CorsOrigins,
			"createdAt":   form.CreatedAt,
			"updatedAt":   form.UpdatedAt,
		},
	})
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
	_, err := h.AuthHelper.RequireAuthenticatedUser(c)
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

	// Convert submissions to serializable format
	submissionsList := make([]map[string]any, len(submissions))
	for i, s := range submissions {
		submissionsList[i] = map[string]any{
			"id":        s.ID,
			"data":      s.Data,
			"status":    s.Status,
			"createdAt": s.CreatedAt,
			"updatedAt": s.UpdatedAt,
		}
	}

	return h.Inertia.Render(c, "Forms/Submissions", inertia.Props{
		"title": "Form Submissions",
		"form": map[string]any{
			"id":    form.ID,
			"title": form.Title,
		},
		"submissions": submissionsList,
	})
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
