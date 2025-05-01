package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	amw "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/form/model"
	"github.com/jonesrussell/goforms/internal/domain/user"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// FormHandler handles form-related requests
type FormHandler struct {
	base           Base
	formService    form.Service
	formClient     form.Client
	authMiddleware *amw.CookieAuthMiddleware
}

// NewFormHandler creates a new form handler
func NewFormHandler(
	logger logging.Logger,
	formService form.Service,
	formClient form.Client,
	userService user.Service,
) (*FormHandler, error) {
	authMiddleware, err := amw.NewCookieAuthMiddleware(userService)
	if err != nil {
		return nil, err
	}

	return &FormHandler{
		base:           NewBase(WithLogger(logger)),
		formService:    formService,
		formClient:     formClient,
		authMiddleware: authMiddleware,
	}, nil
}

// Register sets up the routes for the form handler
func (h *FormHandler) Register(e *echo.Echo) {
	// Public form submission endpoints
	formGroup := e.Group("/v1/forms")
	formGroup.POST("/:formID/submit", h.handleFormSubmission)

	// Protected admin endpoints
	adminGroup := e.Group("/v1/forms")
	adminGroup.Use(h.authMiddleware.RequireAuth)
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
		h.base.Logger.Debug("CSRF token found in request",
			logging.String("path", c.Request().URL.Path),
			logging.String("method", c.Request().Method))

		// Validate CSRF token
		if err := amw.ValidateCSRFToken(c); err != nil {
			h.base.LogError("CSRF token validation failed", err)
			return echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed")
		}
	}

	// Get form data from request body
	var formData map[string]any
	if err := c.Bind(&formData); err != nil {
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

	// Create form response
	response := form.Response{
		ID:          submission.ID,
		FormID:      submission.FormID,
		Values:      submission.Data,
		SubmittedAt: submission.SubmittedAt,
	}

	// Submit response using client
	if submitErr := h.formClient.SubmitResponse(c.Request().Context(), formID, response); submitErr != nil {
		h.base.LogError("failed to submit form response", submitErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit form response")
	}

	// Set security headers
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().Header().Set("X-Frame-Options", "DENY")
	c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
	c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"id":           submission.ID,
			"form_id":      submission.FormID,
			"values":       submission.Data,
			"submitted_at": submission.SubmittedAt,
		},
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

	id, err := strconv.ParseUint(formID, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form ID")
	}

	formData, err := h.formService.GetForm(uint(id))
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

	id, err := strconv.ParseUint(formID, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid form ID")
	}

	deleteErr := h.formService.DeleteForm(uint(id))
	if deleteErr != nil {
		h.base.LogError("failed to delete form", deleteErr)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete form")
	}

	return c.NoContent(http.StatusNoContent)
}
