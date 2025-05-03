package handlers

import (
	"net/http"
	"strconv"

	amw "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
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
	}

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
	}

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
	form, err := h.formService.CreateForm(currentUser.ID, formData.Title, formData.Description, defaultSchema)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create form")
	}

	// Redirect to the form edit page
	return c.Redirect(http.StatusSeeOther, "/dashboard/forms/"+strconv.FormatUint(uint64(form.ID), 10)+"/edit")
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
	form, err := h.formService.GetForm(uint(formID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Verify form belongs to current user
	if form.UserID != currentUser.ID {
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
		Form:      form,
		CSRFToken: csrfToken,
	}

	// Render edit form page
	return pages.EditForm(data).Render(c.Request().Context(), c.Response().Writer)
}
