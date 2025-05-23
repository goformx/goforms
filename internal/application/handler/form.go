package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	amw "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// FormHandler handles form-related requests
type FormHandler struct {
	base           Base
	formService    form.Service
	authMiddleware *amw.CookieAuthMiddleware
	logger         logging.Logger
}

// NewFormHandler creates a new form handler
func NewFormHandler(
	logger logging.Logger,
	formService form.Service,
	userService user.Service,
) (*FormHandler, error) {
	cookieAuth := amw.NewCookieAuthMiddleware(userService, logger)

	return &FormHandler{
		base:           NewBase(WithLogger(logger)),
		formService:    formService,
		authMiddleware: cookieAuth,
		logger:         logger,
	}, nil
}

// Register sets up the routes for the form handler
func (h *FormHandler) Register(e *echo.Echo) {
	// Create base form group
	formGroup := e.Group("/api/v1/forms")

	// Public form submission endpoints
	formGroup.POST("/:formID/submit", h.handleFormSubmission)

	// Protected admin endpoints - apply auth middleware to a subgroup
	adminGroup := formGroup.Group("", h.authMiddleware.RequireAuth)
	adminGroup.GET("", h.handleListForms)
	adminGroup.POST("", h.handleCreateForm)
	adminGroup.GET("/:formID", h.handleGetForm)
	adminGroup.DELETE("/:formID", h.handleDeleteForm)
}

// handleFormSubmission handles form submissions from external websites
func (h *FormHandler) handleFormSubmission(c echo.Context) error {
	formID := c.Param("formID")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	// Verify CSRF token if enabled
	if csrfToken := c.Get("csrf"); csrfToken != nil {
		// Log CSRF token presence for debugging
		h.logger.Debug("CSRF token found in request",
			logging.StringField("path", c.Request().URL.Path),
			logging.StringField("method", c.Request().Method))

		// Validate CSRF token
		if csrfErr := amw.ValidateCSRFToken(c); csrfErr != nil {
			h.base.LogError("CSRF token validation failed", csrfErr)
			return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
		}
	}

	// Get form data from request body
	var formData map[string]any
	if bindErr := c.Bind(&formData); bindErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	// Get metadata (IP, User-Agent, etc.)
	metadata := map[string]string{
		"ip":         c.RealIP(),
		"user_agent": c.Request().UserAgent(),
		"origin":     c.Request().Header.Get("Origin"),
		"referer":    c.Request().Header.Get("Referer"),
	}

	// Create form submission
	submission, err := model.NewFormSubmission(formID, formData, metadata)
	if err != nil {
		h.base.LogError("failed to create form submission", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form submission")
	}

	// Get form to verify it exists
	if _, err := h.formService.GetForm(formID); err != nil {
		h.base.LogError("failed to get form", err)
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	// Store the submission
	_, err = h.formService.GetFormSubmissions(formID)
	if err != nil {
		h.base.LogError("failed to store form submission", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit form")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Form submitted successfully",
		"id":      submission.ID,
	})
}

// handleListForms handles listing all forms for the authenticated user
func (h *FormHandler) handleListForms(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid user ID type")
	}
	forms, err := h.formService.GetUserForms(userID)
	if err != nil {
		h.base.LogError("failed to list forms", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list forms")
	}
	return c.JSON(http.StatusOK, forms)
}

// handleCreateForm handles creating a new form
func (h *FormHandler) handleCreateForm(c echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid user ID type")
	}

	var formData struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Schema      form.JSON `json:"schema"`
	}

	if err := c.Bind(&formData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form data")
	}

	createdForm, err := h.formService.CreateForm(userID, formData.Title, formData.Description, formData.Schema)
	if err != nil {
		h.base.LogError("failed to create form", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create form")
	}

	return c.JSON(http.StatusCreated, createdForm)
}

// handleGetForm handles getting a single form
func (h *FormHandler) handleGetForm(c echo.Context) error {
	formID := c.Param("formID")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	formData, err := h.formService.GetForm(formID)
	if err != nil {
		h.base.LogError("failed to get form", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get form")
	}

	if formData == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Form not found")
	}

	return c.JSON(http.StatusOK, formData)
}

// handleDeleteForm handles deleting a form
func (h *FormHandler) handleDeleteForm(c echo.Context) error {
	formID := c.Param("formID")
	if formID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Form ID is required")
	}

	deleteErr := h.formService.DeleteForm(formID)
	if deleteErr != nil {
		h.base.LogError("failed to delete form", deleteErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete form")
	}

	return c.NoContent(http.StatusNoContent)
}
