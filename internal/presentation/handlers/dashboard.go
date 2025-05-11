package handlers

import (
	"net/http"
	"strconv"

	amw "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/web"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

// Handler handles dashboard-related HTTP requests
type Handler struct {
	authMiddleware *amw.CookieAuthMiddleware
	formService    form.Service
}

// NewHandler creates a new dashboard handler
func NewHandler(
	userService user.Service,
	formService form.Service,
	logger logging.Logger,
) (*Handler, error) {
	cookieAuth := amw.NewCookieAuthMiddleware(userService, logger)

	return &Handler{
		authMiddleware: cookieAuth,
		formService:    formService,
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
	return c.Redirect(http.StatusSeeOther, "/dashboard/forms/"+strconv.FormatUint(uint64(formObj.ID), 10)+"/edit")
}

// ShowEditForm displays the form editing page
func (h *Handler) ShowEditForm(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	// Get form ID from URL parameter
	formID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form ID")
	}

	// Get form from service
	formObj, err := h.formService.GetForm(uint(formID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formObj.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
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
		CSRFToken:            csrfToken,
		AssetPath:            web.GetAssetPath,
		FormBuilderAssetPath: web.GetAssetPath("src/js/form-builder.ts"),
	}

	// Set content
	data.Content = pages.EditFormContent(data)

	// Render edit form page
	return pages.EditForm(data).Render(c.Request().Context(), c.Response().Writer)
}

// ShowFormSubmissions handles viewing form submissions
func (h *Handler) ShowFormSubmissions(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	// Get form ID from URL parameter
	formID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form ID")
	}

	// Get form from service
	formObj, err := h.formService.GetForm(uint(formID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formObj.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	// Get form submissions
	submissions, err := h.formService.GetFormSubmissions(uint(formID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch submissions")
	}

	// Get CSRF token from context
	csrfToken, ok := c.Get("csrf").(string)
	if !ok {
		csrfToken = "" // Set empty string if token not found
	}

	// Create page data
	data := shared.PageData{
		Title:       "Form Submissions - GoForms",
		User:        currentUser,
		Form:        formObj,
		Submissions: submissions,
		CSRFToken:   csrfToken,
		AssetPath:   web.GetAssetPath,
	}

	// Set content
	data.Content = pages.FormSubmissionsContent(data)

	// Render form submissions page
	return pages.FormSubmissions(data).Render(c.Request().Context(), c.Response().Writer)
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

	id, err := strconv.ParseUint(formID, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form ID")
	}

	formData, err := h.formService.GetForm(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get form")
	}

	if formData == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formData.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	// Return the Form.io schema directly
	return c.JSON(http.StatusOK, formData.Schema)
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

	id, err := strconv.ParseUint(formID, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form ID")
	}

	// Get existing form
	formData, err := h.formService.GetForm(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get form")
	}

	if formData == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if formData.UserID != currentUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	// Parse the new schema from request body
	var newSchema map[string]any
	if bindErr := c.Bind(&newSchema); bindErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid schema format")
	}

	// Update the form's schema directly
	formData.Schema = form.JSON(newSchema)
	if updateErr := h.formService.UpdateForm(formData); updateErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update form schema")
	}

	return c.JSON(http.StatusOK, newSchema)
}

// UpdateForm handles updating a form's basic details
func (h *Handler) UpdateForm(c echo.Context) error {
	currentUser, ok := c.Get("user").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	formID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form ID")
	}

	// Get existing form
	formObj, err := h.formService.GetForm(uint(formID))
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
