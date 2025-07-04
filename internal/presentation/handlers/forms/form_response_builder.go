package forms

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// FormResponseBuilder builds form-related responses
type FormResponseBuilder struct {
	config       *config.Config
	assetManager web.AssetManagerInterface
	renderer     view.Renderer
	logger       logging.Logger
}

// NewFormResponseBuilder creates a new form response builder
func NewFormResponseBuilder(
	cfg *config.Config,
	assetManager web.AssetManagerInterface,
	renderer view.Renderer,
	logger logging.Logger,
) *FormResponseBuilder {
	return &FormResponseBuilder{
		config:       cfg,
		assetManager: assetManager,
		renderer:     renderer,
		logger:       logger,
	}
}

// BuildNewFormResponse builds a response for the new form page
func (b *FormResponseBuilder) BuildNewFormResponse(c echo.Context, user *entities.User) error {
	b.logger.Info("new form page rendered", "user_id", user.ID)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":    user.ID,
					"email": user.Email,
					"role":  user.Role,
				},
			},
		})
	}

	// For web requests, render new form page
	return b.renderNewForm(c, user)
}

// BuildEditFormResponse builds a response for the edit form page
func (b *FormResponseBuilder) BuildEditFormResponse(c echo.Context, user *entities.User, form *model.Form) error {
	b.logger.Info("edit form page rendered", "user_id", user.ID, "form_id", form.ID)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":    user.ID,
					"email": user.Email,
					"role":  user.Role,
				},
				"form": form,
			},
		})
	}

	// For web requests, render edit form page
	return b.renderEditForm(c, user, form)
}

// BuildCreateFormSuccessResponse builds a successful form creation response
func (b *FormResponseBuilder) BuildCreateFormSuccessResponse(c echo.Context, form *model.Form) error {
	b.logger.Info("form created successfully", "form_id", form.ID, "title", form.Title)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"status":  "success",
			"message": "Form created successfully",
			"data": map[string]interface{}{
				"form": form,
			},
		})
	}

	// For web requests, redirect to edit form page
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/forms/%s/edit", form.ID))
}

// BuildUpdateFormSuccessResponse builds a successful form update response
func (b *FormResponseBuilder) BuildUpdateFormSuccessResponse(c echo.Context, form *model.Form) error {
	b.logger.Info("form updated successfully", "form_id", form.ID, "title", form.Title)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": "Form updated successfully",
			"data": map[string]interface{}{
				"form": form,
			},
		})
	}

	// For web requests, redirect to edit form page
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/forms/%s/edit", form.ID))
}

// BuildDeleteFormSuccessResponse builds a successful form deletion response
func (b *FormResponseBuilder) BuildDeleteFormSuccessResponse(c echo.Context) error {
	b.logger.Info("form deleted successfully")

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": "Form deleted successfully",
		})
	}

	// For web requests, redirect to dashboard
	return c.Redirect(http.StatusSeeOther, constants.PathDashboard)
}

// BuildFormSubmissionsResponse builds a response for form submissions page
func (b *FormResponseBuilder) BuildFormSubmissionsResponse(
	c echo.Context,
	user *entities.User,
	form *model.Form,
	submissions []*model.FormSubmission,
) error {
	b.logger.Info("form submissions page rendered",
		"user_id", user.ID,
		"form_id", form.ID,
		"submissions_count", len(submissions))

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":    user.ID,
					"email": user.Email,
					"role":  user.Role,
				},
				"form":        form,
				"submissions": submissions,
				"count":       len(submissions),
			},
		})
	}

	// For web requests, render form submissions page
	return b.renderFormSubmissions(c, user, form, submissions)
}

// BuildFormErrorResponse builds an error response for form operations
func (b *FormResponseBuilder) BuildFormErrorResponse(c echo.Context, errorMessage string, statusCode int) error {
	b.logger.Warn("form operation failed", "error", errorMessage, "status_code", statusCode)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(statusCode, map[string]interface{}{
			"status":  "error",
			"message": errorMessage,
		})
	}

	// For web requests, render appropriate page with error
	if strings.Contains(errorMessage, "authentication") || strings.Contains(errorMessage, "unauthorized") {
		return c.Redirect(http.StatusSeeOther, constants.PathLogin)
	}

	// For other errors, redirect to dashboard with error
	return c.Redirect(http.StatusSeeOther, constants.PathDashboard)
}

// BuildAuthenticationErrorResponse builds an authentication error response
func (b *FormResponseBuilder) BuildAuthenticationErrorResponse(c echo.Context) error {
	b.logger.Warn("authentication required for form access")

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"status":  "error",
			"message": "Authentication required",
		})
	}

	// For web requests, redirect to login
	return c.Redirect(http.StatusSeeOther, constants.PathLogin)
}

// BuildFormNotFoundResponse builds a form not found response
func (b *FormResponseBuilder) BuildFormNotFoundResponse(c echo.Context) error {
	b.logger.Warn("form not found")

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "Form not found",
		})
	}

	// For web requests, redirect to dashboard
	return c.Redirect(http.StatusSeeOther, constants.PathDashboard)
}

// BuildValidationErrorResponse builds a validation error response
func (b *FormResponseBuilder) BuildValidationErrorResponse(c echo.Context, field, message string) error {
	b.logger.Warn("validation error", "field", field, "message", message)

	// Check if this is an API request
	if b.isAPIRequest(c) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Validation failed",
			"errors": map[string]string{
				field: message,
			},
		})
	}

	// For web requests, redirect back with error
	return c.Redirect(http.StatusSeeOther, c.Request().Referer())
}

// Helper methods

// isAPIRequest checks if the request is an API request
func (b *FormResponseBuilder) isAPIRequest(c echo.Context) bool {
	accept := c.Request().Header.Get("Accept")
	contentType := c.Request().Header.Get("Content-Type")

	return strings.Contains(accept, "application/json") ||
		strings.Contains(contentType, "application/json") ||
		strings.HasPrefix(c.Request().URL.Path, "/api/")
}

// renderNewForm renders the new form page
func (b *FormResponseBuilder) renderNewForm(c echo.Context, user *entities.User) error {
	// Create page data for the new form template
	pageData := view.NewPageData(b.config, b.assetManager, c, "New Form")
	pageData = pageData.WithUser(user)

	// Render the new form page using the generated template
	newFormComponent := pages.NewForm(*pageData)

	return b.renderer.Render(c, newFormComponent)
}

// renderEditForm renders the edit form page
func (b *FormResponseBuilder) renderEditForm(c echo.Context, user *entities.User, form *model.Form) error {
	// Create page data for the edit form template
	pageData := view.NewPageData(b.config, b.assetManager, c, "Edit Form")
	pageData = pageData.WithUser(user).WithForm(form)

	// Render the edit form page using the generated template
	editFormComponent := pages.EditForm(*pageData, form)

	return b.renderer.Render(c, editFormComponent)
}

// renderFormSubmissions renders the form submissions page
func (b *FormResponseBuilder) renderFormSubmissions(
	c echo.Context,
	user *entities.User,
	form *model.Form,
	submissions []*model.FormSubmission,
) error {
	// Create page data for the form submissions template
	pageData := view.NewPageData(b.config, b.assetManager, c, "Form Submissions")
	pageData = pageData.WithUser(user).WithForm(form).WithSubmissions(submissions)

	// Render the form submissions page using the generated template
	formSubmissionsComponent := pages.FormSubmissions(*pageData)

	return b.renderer.Render(c, formSubmissionsComponent)
}
