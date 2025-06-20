package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/common/types"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/labstack/echo/v4"
)

// Default CORS settings for forms
const (
	DefaultCorsMethods = "GET,POST,OPTIONS"
	DefaultCorsHeaders = "Content-Type,Accept,Origin"
)

// FormWebHandler handles web UI form operations
type FormWebHandler struct {
	*FormBaseHandler
	Sanitizer sanitization.ServiceInterface
}

func NewFormWebHandler(
	base *BaseHandler,
	formService formdomain.Service,
	formValidator *validation.FormValidator,
	sanitizer sanitization.ServiceInterface,
) *FormWebHandler {
	return &FormWebHandler{
		FormBaseHandler: NewFormBaseHandler(base, formService, formValidator),
		Sanitizer:       sanitizer,
	}
}

func (h *FormWebHandler) RegisterRoutes(e *echo.Echo, accessManager *access.AccessManager) {
	forms := e.Group(constants.PathForms)
	forms.Use(access.Middleware(accessManager, h.Logger))

	forms.GET("/new", h.handleNew)
	forms.POST("", h.handleCreate)
	forms.GET("/:id/edit", h.handleEdit)
	forms.POST("/:id/edit", h.handleUpdate)
	forms.DELETE("/:id", h.handleDelete)
	forms.GET("/:id/submissions", h.handleSubmissions)
}

// Register satisfies the Handler interface
func (h *FormWebHandler) Register(e *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// Start satisfies the Handler interface
func (h *FormWebHandler) Start(ctx context.Context) error {
	return nil // No initialization needed
}

// Stop satisfies the Handler interface
func (h *FormWebHandler) Stop(ctx context.Context) error {
	return nil // No cleanup needed
}

func (h *FormWebHandler) handleNew(c echo.Context) error {
	user, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	data := h.BuildPageData(c, "New Form")
	data.User = user
	return h.Renderer.Render(c, pages.NewForm(data))
}

func (h *FormWebHandler) handleCreate(c echo.Context) error {
	user, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	// Get and sanitize form data
	title := h.Sanitizer.String(c.FormValue("title"))
	description := h.Sanitizer.String(c.FormValue("description"))

	// CORS config (comma-separated values from form input)
	corsOrigins := h.Sanitizer.String(c.FormValue("cors_origins"))

	// Parse comma-separated values into string slices
	origins := types.StringArray(parseCSV(corsOrigins))

	// Use sensible defaults for methods and headers
	methods := types.StringArray(parseCSV(DefaultCorsMethods))
	headers := types.StringArray(parseCSV(DefaultCorsHeaders))

	// Create a valid initial schema
	schema := model.JSON{
		"type":       "object",
		"components": []any{},
	}

	// Create the form
	form := model.NewForm(user.ID, title, description, schema)
	form.CorsOrigins = origins
	form.CorsMethods = methods
	form.CorsHeaders = headers

	err = h.FormService.CreateForm(c.Request().Context(), form)
	if err != nil {
		h.Logger.Error("failed to create form", "error", err)

		// Check for specific validation errors
		switch {
		case errors.Is(err, model.ErrFormTitleRequired):
			return h.HandleError(c, err, "Form title is required")
		case errors.Is(err, model.ErrFormSchemaRequired):
			return h.HandleError(c, err, "Form schema is required")
		default:
			return h.HandleError(c, err, "Failed to create form")
		}
	}

	return c.Redirect(constants.StatusSeeOther, fmt.Sprintf("/forms/%s/edit", form.ID))
}

func (h *FormWebHandler) handleEdit(c echo.Context) error {
	user, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	// Debug logging for form data
	h.Logger.Debug("Form edit page loaded",
		"form_id", form.ID,
		"form_title", form.Title,
		"form_status", form.Status,
	)

	data := h.BuildPageData(c, "Edit Form")
	data.User = user
	data.Form = form
	data.FormBuilderAssetPath = h.AssetManager.AssetPath("src/js/form-builder.ts")

	return pages.EditForm(data, form).Render(c.Request().Context(), c.Response().Writer)
}

func (h *FormWebHandler) handleUpdate(c echo.Context) error {
	_, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	// Update form fields
	form.Title = h.Sanitizer.String(c.FormValue("title"))
	form.Description = h.Sanitizer.String(c.FormValue("description"))
	form.Status = h.Sanitizer.String(c.FormValue("status"))

	// CORS config (comma-separated values from form input)
	corsOrigins := h.Sanitizer.String(c.FormValue("cors_origins"))

	// Parse comma-separated values into string slices
	origins := types.StringArray(parseCSV(corsOrigins))

	// Use sensible defaults for methods and headers
	methods := types.StringArray(parseCSV(DefaultCorsMethods))
	headers := types.StringArray(parseCSV(DefaultCorsHeaders))

	form.CorsOrigins = origins
	form.CorsMethods = methods
	form.CorsHeaders = headers

	err = h.FormService.UpdateForm(c.Request().Context(), form)
	if err != nil {
		h.Logger.Error("failed to update form", "error", err)
		return h.HandleError(c, err, "Failed to update form")
	}

	// Return success response instead of redirect
	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "Form updated successfully",
		"form_id": form.ID,
	})
}

func (h *FormWebHandler) handleDelete(c echo.Context) error {
	_, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	err = h.FormService.DeleteForm(c.Request().Context(), form.ID)
	if err != nil {
		h.Logger.Error("failed to delete form", "error", err)
		return h.HandleError(c, err, "Failed to delete form")
	}

	return c.NoContent(constants.StatusNoContent)
}

func (h *FormWebHandler) handleSubmissions(c echo.Context) error {
	user, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return err
	}

	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return err
	}

	submissions, err := h.FormService.ListFormSubmissions(c.Request().Context(), form.ID)
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

// parseCSV parses a comma-separated string into a slice of strings, trimming whitespace and skipping empty values
func parseCSV(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
