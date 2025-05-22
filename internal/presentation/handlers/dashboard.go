package handlers

import (
	"net/http"

	amw "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/form/model"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/web"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type BaseHandler struct {
	LogError func(msg string, err error)
}

// Handler handles dashboard-related HTTP requests
type Handler struct {
	authMiddleware *amw.CookieAuthMiddleware
	formService    form.Service
	base           *BaseHandler
}

// NewHandler creates a new dashboard handler
func NewHandler(
	userService user.Service,
	formService form.Service,
	logger logging.Logger,
	base *BaseHandler,
) (*Handler, error) {
	cookieAuth := amw.NewCookieAuthMiddleware(userService, logger)

	return &Handler{
		authMiddleware: cookieAuth,
		formService:    formService,
		base:           base,
	}, nil
}

// Register sets up the dashboard routes
func (h *Handler) Register(e *echo.Echo) {
	// Dashboard routes
	dashboard := e.Group("/dashboard")
	dashboard.Use(h.authMiddleware.RequireAuth) // Middleware to ensure user is authenticated

	dashboard.GET("", h.ShowDashboard)
	dashboard.GET("/forms/new", h.ShowNewForm)
	dashboard.POST("/forms", h.CreateForm)
	dashboard.GET("/forms/:id/edit", h.ShowEditForm)
	dashboard.PUT("/forms/:id", h.UpdateForm)
	dashboard.GET("/forms/:id/submissions", h.ShowFormSubmissions)
	dashboard.GET("/forms/:id/schema", h.GetFormSchema)
	dashboard.PUT("/forms/:id/schema", h.UpdateFormSchema)
	dashboard.DELETE("/forms/:id", h.DeleteForm)
}

// ShowDashboard displays the user's dashboard
func (h *Handler) ShowDashboard(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	// Get user's forms
	forms, err := h.formService.GetUserForms(currentUser.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch forms")
	}

	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = "" // Set empty string if token not found
	}

	// Create page data
	data := shared.PageData{
		Title:     "Dashboard - GoForms",
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
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = "" // Set empty string if token not found
	}

	// Create page data
	data := shared.PageData{
		Title:     "Create New Form - GoForms",
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
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	var formData struct {
		Title       string `json:"title" form:"title"`
		Description string `json:"description" form:"description"`
	}

	if err := c.Bind(&formData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if err := c.Validate(formData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Create a minimal Form.io schema for the form
	defaultSchema := form.JSON{
		"display":    "form",
		"components": []any{},
	}

	// Create the form
	formObj, err := h.formService.CreateForm(currentUser.ID, formData.Title, formData.Description, defaultSchema)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create form")
	}

	// Redirect to the form edit page
	return c.Redirect(http.StatusSeeOther, "/dashboard/forms/"+formObj.ID+"/edit")
}

// ShowEditForm displays the form editing page
func (h *Handler) ShowEditForm(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	formID := c.Param("id")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	// Get form data
	formObj, err := h.formService.GetForm(formID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formObj.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	// Get form submissions
	submissions, err := h.formService.GetFormSubmissions(formID)
	if err != nil {
		h.base.LogError("failed to get form submissions", err)
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
		Title:                "Edit Form - GoForms",
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
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	formID := c.Param("id")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	// Get form data
	formObj, err := h.formService.GetForm(formID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formObj.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	// Get form submissions
	submissions, err := h.formService.GetFormSubmissions(formID)
	if err != nil {
		h.base.LogError("failed to get form submissions", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get form submissions")
	}

	// Render the submissions page
	return c.Render(http.StatusOK, "form_submissions.html", map[string]any{
		"Form":        formObj,
		"Submissions": submissions,
	})
}

// GetFormSchema handles getting a form's schema
func (h *Handler) GetFormSchema(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	formID := c.Param("id")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	// Get form data
	formObj, err := h.formService.GetForm(formID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formObj.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	return c.JSON(http.StatusOK, formObj.Schema)
}

// UpdateFormSchema handles updating a form's schema
func (h *Handler) UpdateFormSchema(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	formID := c.Param("id")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	// Get form data
	formObj, err := h.formService.GetForm(formID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formObj.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	var schema form.JSON
	if bindErr := c.Bind(&schema); bindErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid schema data")
	}

	// Update form schema
	formObj.Schema = schema
	if updateErr := h.formService.UpdateForm(formObj); updateErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update form schema")
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateForm handles updating a form's basic details
func (h *Handler) UpdateForm(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	formID := c.Param("id")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	// Get existing form
	formObj, err := h.formService.GetForm(formID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formObj.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	var formData struct {
		Title       string `json:"title" form:"title"`
		Description string `json:"description" form:"description"`
	}

	if bindErr := c.Bind(&formData); bindErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	if validateErr := c.Validate(formData); validateErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, validateErr.Error())
	}

	// Update form details
	formObj.Title = formData.Title
	formObj.Description = formData.Description

	if updateErr := h.formService.UpdateForm(formObj); updateErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update form")
	}

	return c.JSON(http.StatusOK, formObj)
}

func (h *Handler) DeleteForm(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	formID := c.Param("id")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	// Get form data
	formObj, err := h.formService.GetForm(formID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formObj.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	if deleteErr := h.formService.DeleteForm(formID); deleteErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete form")
	}

	return c.NoContent(http.StatusNoContent)
}
