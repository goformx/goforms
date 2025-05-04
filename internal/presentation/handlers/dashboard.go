package handlers

import (
	"net/http"
	"strconv"

	amw "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/web"
	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
	"github.com/jonesrussell/goforms/internal/presentation/templates/shared"
	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	authMiddleware *amw.CookieAuthMiddleware
	formService    form.Service
}

func NewDashboardHandler(userService user.Service, formService form.Service) (*DashboardHandler, error) {
	authMiddleware, err := amw.NewCookieAuthMiddleware(userService)
	if err != nil {
		return nil, err
	}

	return &DashboardHandler{
		authMiddleware: authMiddleware,
		formService:    formService,
	}, nil
}

func (h *DashboardHandler) Register(e *echo.Echo) {
	// Dashboard routes
	dashboard := e.Group("/dashboard")
	dashboard.Use(h.authMiddleware.RequireAuth) // Middleware to ensure user is authenticated

	dashboard.GET("", h.ShowDashboard)
	dashboard.GET("/forms/new", h.ShowNewForm)
	dashboard.POST("/forms", h.CreateForm)
	dashboard.GET("/forms/:id/edit", h.ShowEditForm)
	dashboard.GET("/forms/:id/submissions", h.ShowFormSubmissions)
	dashboard.GET("/forms/:id/schema", h.GetFormSchema)
	dashboard.PUT("/forms/:id/schema", h.UpdateFormSchema)
}

func (h *DashboardHandler) ShowDashboard(c echo.Context) error {
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

func (h *DashboardHandler) ShowNewForm(c echo.Context) error {
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

func (h *DashboardHandler) CreateForm(c echo.Context) error {
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

	// Create a default schema for the form
	defaultSchema := form.JSON{
		"type": "object",
		"properties": map[string]any{
			"fields": []map[string]any{},
		},
		"required": []string{},
	}

	// Create the form
	formObj, err := h.formService.CreateForm(currentUser.ID, formData.Title, formData.Description, defaultSchema)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create form")
	}

	// Redirect to the form edit page
	return c.Redirect(http.StatusSeeOther, "/dashboard/forms/"+strconv.FormatUint(uint64(formObj.ID), 10)+"/edit")
}

func (h *DashboardHandler) ShowEditForm(c echo.Context) error {
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
		Title:     "Edit Form - GoForms",
		User:      currentUser,
		Form:      formObj,
		CSRFToken: csrfToken,
		AssetPath: web.GetAssetPath,
	}

	// Set content
	data.Content = pages.EditFormContent(data)

	// Render edit form page
	return pages.EditForm(data).Render(c.Request().Context(), c.Response().Writer)
}

// ShowFormSubmissions handles viewing form submissions
func (h *DashboardHandler) ShowFormSubmissions(c echo.Context) error {
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
func (h *DashboardHandler) GetFormSchema(c echo.Context) error {
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

	// Convert the form schema to the expected format
	schema := map[string]interface{}{
		"id":     formData.ID,
		"fields": []interface{}{}, // Default to empty fields array if schema is nil
	}

	// If schema exists and has fields, use them
	if formData.Schema != nil {
		if fields, ok := formData.Schema["fields"].([]interface{}); ok {
			schema["fields"] = fields
		}
	}

	return c.JSON(http.StatusOK, schema)
}

// UpdateFormSchema handles updating a form's schema
func (h *DashboardHandler) UpdateFormSchema(c echo.Context) error {
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
	var newSchema map[string]interface{}
	if err := c.Bind(&newSchema); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid schema format")
	}

	// Update the form's schema
	formData.Schema = form.JSON(newSchema)
	if err := h.formService.UpdateForm(formData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update form schema")
	}

	return c.JSON(http.StatusOK, newSchema)
}
