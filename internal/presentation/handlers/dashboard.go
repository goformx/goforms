package handlers

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/services/formops"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// FormData represents the structure for form creation and updates
type FormData struct {
	Title       string `json:"title" form:"title" validate:"required"`
	Description string `json:"description" form:"description" validate:"required"`
}

// Handler handles dashboard-related HTTP requests
type Handler struct {
	FormHandler       *FormHandler
	SubmissionHandler *SubmissionHandler
	SchemaHandler     *SchemaHandler
	logger            logging.Logger
}

// NewHandler creates a new dashboard handler
func NewHandler(
	userService user.Service,
	formService form.Service,
	logger logging.Logger,
) (*Handler, error) {
	base := NewBaseHandler(
		formService,
		logger,
	)

	// Create form operations service
	formOperations := formops.NewService(formService, logger)

	return &Handler{
		FormHandler:       NewFormHandler(formService, formOperations, logger, base),
		SubmissionHandler: NewSubmissionHandler(formService, logger, base),
		SchemaHandler:     NewSchemaHandler(formService, logger, base),
		logger:            logger,
	}, nil
}

// getAuthenticatedUser retrieves and validates the authenticated user from the context
func (h *Handler) getAuthenticatedUser(c echo.Context) (*user.User, error) {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}
	return currentUser, nil
}

// getOwnedForm retrieves a form and verifies ownership
func (h *Handler) getOwnedForm(c echo.Context, currentUser *user.User) (*form.Form, error) {
	formID := c.Param("id")
	if formID == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	formObj, err := h.FormHandler.formService.GetForm(formID)
	if err != nil {
		h.logger.Error("Failed to get form", err)
		return nil, echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	if formObj.UserID != currentUser.ID {
		return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	return formObj, nil
}

// handleError is a helper function to consistently handle and log errors
func (h *Handler) handleError(err error, status int, message string) error {
	h.logger.Error(message, err)
	return echo.NewHTTPError(status, message)
}

// Register sets up the dashboard routes
func (h *Handler) Register(e *echo.Echo) {
	dashboard := e.Group("/dashboard")
	// You may want to add middleware here if needed
	dashboard.GET("", h.ShowDashboard)

	h.FormHandler.Register(e)
	h.SubmissionHandler.Register(e)
	h.SchemaHandler.Register(e)
}

// ShowDashboard displays the user's dashboard
func (h *Handler) ShowDashboard(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil || currentUser == nil {
		h.logger.Error("ShowDashboard: user is nil or authentication failed", logging.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	// Get user's forms
	forms, err := h.FormHandler.formService.GetUserForms(currentUser.ID)
	if err != nil {
		return h.handleError(err, http.StatusInternalServerError, "Failed to fetch forms")
	}

	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = "" // Set empty string if token not found
	}

	// Create page data
	data := shared.PageData{
		Title:     "Dashboard - GoFormX",
		User:      currentUser,
		Forms:     forms,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}

	// Set content
	data.Content = pages.DashboardContent(data)

	// Render dashboard page
	return pages.Dashboard(data).Render(c.Request().Context(), c.Response().Writer)
}

// ShowNewForm displays the form creation page
func (h *Handler) ShowNewForm(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil || currentUser == nil {
		h.logger.Error("ShowNewForm: user is nil or authentication failed", logging.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = "" // Set empty string if token not found
	}

	// Create page data
	data := shared.PageData{
		Title:     "Create New Form - GoFormX",
		User:      currentUser,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}

	// Set content
	data.Content = pages.NewFormContent(data)

	// Render new form page
	return pages.NewForm(data).Render(c.Request().Context(), c.Response().Writer)
}

// CreateForm handles form creation
func (h *Handler) CreateForm(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	var formData FormData
	if bindErr := c.Bind(&formData); bindErr != nil {
		return h.handleError(bindErr, http.StatusBadRequest, "Invalid form data")
	}

	if validateErr := c.Validate(&formData); validateErr != nil {
		return h.handleError(validateErr, http.StatusUnprocessableEntity, "Form validation failed")
	}

	// Create a minimal Form.io schema for the form
	defaultSchema := form.JSON{
		"display":    "form",
		"components": []any{},
	}

	// Create the form
	formObj, createErr := h.FormHandler.formService.CreateForm(
		currentUser.ID,
		formData.Title,
		formData.Description,
		defaultSchema,
	)
	if createErr != nil {
		return h.handleError(createErr, http.StatusInternalServerError, "Failed to create form")
	}

	// Redirect to the form edit page
	return c.Redirect(http.StatusSeeOther, "/dashboard/forms/"+formObj.ID+"/edit")
}

// ShowEditForm displays the form editing page
func (h *Handler) ShowEditForm(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	// Get form submissions
	submissions, err := h.FormHandler.formService.GetFormSubmissions(formObj.ID)
	if err != nil {
		h.logger.Error("failed to get form submissions", err)
		// Don't return error, just show empty submissions
		submissions = []*model.FormSubmission{}
	}

	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = "" // Set empty string if token not found
	}

	// Create page data
	data := shared.PageData{
		Title:                "Edit Form - GoFormX",
		User:                 currentUser,
		Form:                 formObj,
		Submissions:          submissions,
		CSRFToken:            csrfToken,
		AssetPath:            web.GetAssetPath,
		FormBuilderAssetPath: web.GetAssetPath("src/js/form-builder.ts"),
	}

	// Render the edit form page
	return pages.EditForm(data).Render(c.Request().Context(), c.Response().Writer)
}

// ShowFormSubmissions handles viewing form submissions
func (h *Handler) ShowFormSubmissions(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	// Get form submissions
	submissions, err := h.FormHandler.formService.GetFormSubmissions(formObj.ID)
	if err != nil {
		return h.handleError(err, http.StatusInternalServerError, "Failed to get form submissions")
	}

	// Render the submissions page
	return c.Render(http.StatusOK, "form_submissions.html", map[string]any{
		"Form":        formObj,
		"Submissions": submissions,
	})
}

// GetFormSchema handles getting a form's schema
func (h *Handler) GetFormSchema(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, formObj.Schema)
}

// UpdateFormSchema handles updating a form's schema
func (h *Handler) UpdateFormSchema(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	// Bind the schema data directly to form.JSON
	var schema form.JSON
	if bindErr := c.Bind(&schema); bindErr != nil {
		return h.handleError(bindErr, http.StatusBadRequest, "Invalid schema data")
	}

	// Validate schema structure
	if schema == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Schema data is required")
	}

	// Ensure required fields are present with default values
	if _, ok := schema["display"]; !ok {
		schema["display"] = "form"
	}
	if _, ok := schema["components"]; !ok {
		schema["components"] = []any{}
	}

	// Validate components array
	if _, ok := schema["components"].([]any); !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Components must be an array")
	}

	// Update form schema while preserving other fields
	formObj.Schema = schema

	// Ensure we preserve all form fields
	formObj.UserID = currentUser.ID // Ensure user ID is set correctly
	formObj.Active = true           // Ensure form is active

	if updateErr := h.FormHandler.formService.UpdateForm(formObj); updateErr != nil {
		return h.handleError(updateErr, http.StatusInternalServerError, "Failed to update form schema")
	}

	return c.JSON(http.StatusOK, formObj)
}

// UpdateForm handles updating a form's basic details
func (h *Handler) UpdateForm(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	var formData FormData
	if bindErr := c.Bind(&formData); bindErr != nil {
		return h.handleError(bindErr, http.StatusBadRequest, "Invalid form data")
	}

	if validateErr := c.Validate(&formData); validateErr != nil {
		return h.handleError(validateErr, http.StatusUnprocessableEntity, "Form validation failed")
	}

	// Update form details
	formObj.Title = formData.Title
	formObj.Description = formData.Description

	if updateErr := h.FormHandler.formService.UpdateForm(formObj); updateErr != nil {
		return h.handleError(updateErr, http.StatusInternalServerError, "Failed to update form")
	}

	return c.JSON(http.StatusOK, formObj)
}

// DeleteForm handles form deletion
func (h *Handler) DeleteForm(c echo.Context) error {
	currentUser, err := h.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	if deleteErr := h.FormHandler.formService.DeleteForm(formObj.ID); deleteErr != nil {
		return h.handleError(deleteErr, http.StatusInternalServerError, "Failed to delete form")
	}

	return c.NoContent(http.StatusNoContent)
}
