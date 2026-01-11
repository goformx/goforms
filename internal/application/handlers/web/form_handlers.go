// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"errors"
	"fmt"
	"net/http"
	"os"

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

	// Debug: Log CSRF token from context with configured key
	csrfConfig := h.Config.Security.CSRF
	csrfContextKey := csrfConfig.ContextKey
	if csrfContextKey == "" {
		csrfContextKey = "csrf"
	}
	csrfToken, _ := c.Get(csrfContextKey).(string)

	// Also check cookie
	csrfCookieName := csrfConfig.CookieName
	if csrfCookieName == "" {
		csrfCookieName = "_csrf"
	}
	csrfCookie, _ := c.Cookie(csrfCookieName)

	cookieTokenLength := 0
	if csrfCookie != nil && csrfCookie.Value != "" {
		cookieTokenLength = len(csrfCookie.Value)
	}

	// Use fmt.Fprintf to os.Stdout to bypass logger sanitization for debugging
	if h.Config.App.IsDevelopment() {
		fmt.Fprintf(os.Stdout, "[CSRF DEBUG] handleNew: path=%s, context_key=%q, cookie_name=%q, token_in_context=%v (len=%d), token_in_cookie=%v (len=%d)\n",
			c.Request().URL.Path,
			csrfContextKey,
			csrfCookieName,
			csrfToken != "",
			len(csrfToken),
			csrfCookie != nil && csrfCookie.Value != "",
			cookieTokenLength)
		if csrfToken != "" {
			fmt.Fprintf(os.Stdout, "[CSRF DEBUG] Token from context: %s\n", csrfToken)
		}
		if csrfCookie != nil && csrfCookie.Value != "" {
			fmt.Fprintf(os.Stdout, "[CSRF DEBUG] Token from cookie: %s\n", csrfCookie.Value)
		}
	}

	h.Logger.Debug("handleNew: CSRF token check",
		"path", c.Request().URL.Path,
		"context_key", csrfContextKey,
		"cookie_name", csrfCookieName,
		"token_in_context", csrfToken != "",
		"token_in_context_length", len(csrfToken),
		"token_in_cookie", csrfCookie != nil && csrfCookie.Value != "",
		"token_in_cookie_length", cookieTokenLength)

	data := h.NewPageData(c, "New Form")
	data.SetUser(user)

	// Debug: Log CSRF token in page data
	tokenValuePreview := ""
	if len(data.CSRFToken) > 0 && len(data.CSRFToken) <= 50 {
		tokenValuePreview = data.CSRFToken
	} else if len(data.CSRFToken) > 50 {
		tokenValuePreview = data.CSRFToken[:50] + "..."
	}

	h.Logger.Debug("handleNew: CSRF token in page data",
		"token_present", data.CSRFToken != "",
		"token_length", len(data.CSRFToken),
		"token_value_preview", tokenValuePreview)

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
		return fmt.Errorf("handle error: %w", h.ErrorHandler.HandleError(c, err))
	}

	// Create form using business logic service
	form, err := h.FormService.CreateForm(c.Request().Context(), user.ID, req)
	if err != nil {
		return h.handleFormCreationError(c, err)
	}

	// For AJAX requests, return JSON so the frontend can handle the redirect
	if c.Request().Header.Get(constants.HeaderXRequestedWith) == "XMLHttpRequest" {
		return h.ResponseBuilder.BuildSuccessResponse(c, "Form created successfully", map[string]any{
			"form_id": form.ID,
		})
	}

	// For regular form submissions, redirect to the edit page
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

	if renderErr := h.Renderer.Render(c, pages.FormSubmissions(*data)); renderErr != nil {
		return fmt.Errorf("failed to render form submissions page: %w", renderErr)
	}

	return nil
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
