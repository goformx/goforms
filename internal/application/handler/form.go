package handler

import (
	"fmt"
	"net/http"
	"time"

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
	formClient     form.Client
	authMiddleware *amw.CookieAuthMiddleware
	logger         logging.Logger
}

// NewFormHandler creates a new form handler
func NewFormHandler(
	logger logging.Logger,
	formService form.Service,
	formClient form.Client,
	userService user.Service,
) (*FormHandler, error) {
	cookieAuth := amw.NewCookieAuthMiddleware(userService, logger)

	return &FormHandler{
		base:           NewBase(WithLogger(logger)),
		formService:    formService,
		formClient:     formClient,
		authMiddleware: cookieAuth,
		logger:         logger,
	}, nil
}

// Register sets up the routes for the form handler
func (h *FormHandler) Register(e *echo.Echo) {
	// Create base form group
	formGroup := e.Group("/api/v1/forms")

	// Public form submission endpoints with rate limiting
	formGroup.POST("/:formID/submit", h.handleFormSubmission, amw.RateLimit(100, 1*time.Minute)) // 100 requests per minute
	formGroup.GET("/:formID/schema", h.handleGetFormSchema)

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
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusBadRequest, "Form ID is required"),
			"invalid form submission",
		)
	}

	// Verify CSRF token if enabled
	if csrfToken := c.Get("csrf"); csrfToken != nil {
		// Log CSRF token presence for debugging
		h.logger.Debug("CSRF token found in request",
			logging.StringField("path", c.Request().URL.Path),
			logging.StringField("method", c.Request().Method),
			logging.StringField("form_id", formID),
			logging.StringField("ip", c.RealIP()))

		// Validate CSRF token
		if csrfErr := amw.ValidateCSRFToken(c); csrfErr != nil {
			h.base.LogError("CSRF token validation failed", csrfErr,
				logging.StringField("form_id", formID),
				logging.StringField("ip", c.RealIP()))
			return h.base.WrapResponseError(
				echo.NewHTTPError(http.StatusForbidden, "CSRF token validation failed"),
				"security validation failed",
			)
		}
	}

	// Get form data from request body
	var formData map[string]any
	if bindErr := c.Bind(&formData); bindErr != nil {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusBadRequest, "Invalid form data"),
			"form data binding failed",
		)
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
		h.base.LogError("failed to create form submission", err,
			logging.StringField("form_id", formID),
			logging.StringField("ip", c.RealIP()))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusBadRequest, "Invalid form submission"),
			"form submission creation failed",
		)
	}

	// Create form response
	response := form.Response{
		ID:          submission.ID,
		FormID:      formID,
		Values:      submission.Data,
		SubmittedAt: submission.SubmittedAt,
	}

	// Submit response using client
	if submitErr := h.formClient.SubmitResponse(c.Request().Context(), formID, response); submitErr != nil {
		h.base.LogError("failed to submit form response", submitErr,
			logging.StringField("form_id", formID),
			logging.StringField("submission_id", submission.ID),
			logging.StringField("ip", c.RealIP()))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit form"),
			"form submission failed",
		)
	}

	// Log successful submission
	h.logger.Info("form submitted successfully",
		logging.StringField("form_id", formID),
		logging.StringField("submission_id", submission.ID),
		logging.StringField("ip", c.RealIP()))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Form submitted successfully",
		"id":      submission.ID,
	})
}

// handleListForms handles listing all forms for the authenticated user
func (h *FormHandler) handleListForms(c echo.Context) error {
	userIDRaw := c.Get("user_id")
	if userIDRaw == nil {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated"),
			"authentication required",
		)
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		h.base.LogError("invalid user ID type", fmt.Errorf("expected uint, got %T", userIDRaw))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusInternalServerError, "Invalid user ID type"),
			"type assertion failed",
		)
	}

	forms, err := h.formService.GetUserForms(userID)
	if err != nil {
		h.base.LogError("failed to list forms", err,
			logging.UintField("user_id", userID))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusInternalServerError, "Failed to list forms"),
			"database operation failed",
		)
	}

	return c.JSON(http.StatusOK, forms)
}

// handleCreateForm handles creating a new form
func (h *FormHandler) handleCreateForm(c echo.Context) error {
	userIDRaw := c.Get("user_id")
	if userIDRaw == nil {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated"),
			"authentication required",
		)
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		h.base.LogError("invalid user ID type", fmt.Errorf("expected uint, got %T", userIDRaw))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusInternalServerError, "Invalid user ID type"),
			"type assertion failed",
		)
	}

	var formData struct {
		Title       string    `json:"title" validate:"required,max=100"`
		Description string    `json:"description" validate:"required,max=500"`
		Schema      form.JSON `json:"schema" validate:"required"`
	}

	if err := c.Bind(&formData); err != nil {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusBadRequest, "Invalid form data"),
			"form data binding failed",
		)
	}

	// Validate form data
	if err := c.Validate(&formData); err != nil {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusBadRequest, err.Error()),
			"form validation failed",
		)
	}

	createdForm, err := h.formService.CreateForm(userID, formData.Title, formData.Description, formData.Schema)
	if err != nil {
		h.base.LogError("failed to create form", err,
			logging.UintField("user_id", userID))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusInternalServerError, "Failed to create form"),
			"database operation failed",
		)
	}

	return c.JSON(http.StatusCreated, createdForm)
}

// handleGetForm handles getting a single form
func (h *FormHandler) handleGetForm(c echo.Context) error {
	formID := c.Param("formID")
	if formID == "" {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusBadRequest, "Form ID is required"),
			"invalid form request",
		)
	}

	formData, err := h.formService.GetForm(formID)
	if err != nil {
		h.base.LogError("failed to get form", err,
			logging.StringField("form_id", formID))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusInternalServerError, "Failed to get form"),
			"database operation failed",
		)
	}

	if formData == nil {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusNotFound, "Form not found"),
			"form not found",
		)
	}

	return c.JSON(http.StatusOK, formData)
}

// handleDeleteForm handles deleting a form
func (h *FormHandler) handleDeleteForm(c echo.Context) error {
	formID := c.Param("formID")
	if formID == "" {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusBadRequest, "Form ID is required"),
			"invalid form request",
		)
	}

	deleteErr := h.formService.DeleteForm(formID)
	if deleteErr != nil {
		h.base.LogError("failed to delete form", deleteErr,
			logging.StringField("form_id", formID))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete form"),
			"database operation failed",
		)
	}

	// Log successful deletion
	h.logger.Info("form deleted successfully",
		logging.StringField("form_id", formID))

	return c.NoContent(http.StatusNoContent)
}

// handleGetFormSchema handles getting a form's schema (public endpoint)
func (h *FormHandler) handleGetFormSchema(c echo.Context) error {
	formID := c.Param("formID")
	if formID == "" {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusBadRequest, "Form ID is required"),
			"invalid form request",
		)
	}

	formData, err := h.formService.GetForm(formID)
	if err != nil {
		h.base.LogError("failed to get form schema", err,
			logging.StringField("form_id", formID))
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusInternalServerError, "Failed to get form schema"),
			"database operation failed",
		)
	}

	if formData == nil {
		return h.base.WrapResponseError(
			echo.NewHTTPError(http.StatusNotFound, "Form not found"),
			"form not found",
		)
	}

	// Log schema access
	h.logger.Debug("form schema accessed",
		logging.StringField("form_id", formID),
		logging.StringField("ip", c.RealIP()))

	return c.JSON(http.StatusOK, formData.Schema)
}
