package web

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/labstack/echo/v4"
)

// FormAPIHandler handles API form operations
type FormAPIHandler struct {
	*FormBaseHandler
	AccessManager *access.AccessManager
}

func NewFormAPIHandler(
	base *BaseHandler,
	formService formdomain.Service,
	accessManager *access.AccessManager,
	formValidator *validation.FormValidator,
) *FormAPIHandler {
	return &FormAPIHandler{
		FormBaseHandler: NewFormBaseHandler(base, formService, formValidator),
		AccessManager:   accessManager,
	}
}

func (h *FormAPIHandler) RegisterRoutes(e *echo.Echo) {
	// API routes with access control
	api := e.Group(constants.PathAPIv1)
	formsAPI := api.Group("/forms")
	formsAPI.Use(access.Middleware(h.AccessManager, h.Logger))
	formsAPI.GET("/:id/schema", h.handleFormSchema)
	formsAPI.PUT("/:id/schema", h.handleFormSchemaUpdate)

	// Public API routes (no authentication required)
	// These are for embedded forms on external websites
	publicAPI := e.Group(constants.PathAPIv1)
	publicFormsAPI := publicAPI.Group("/forms")
	publicFormsAPI.GET("/:id/schema", h.handleFormSchema)
	publicFormsAPI.POST("/:id/submit", h.HandleFormSubmit)
}

// Register satisfies the Handler interface
func (h *FormAPIHandler) Register(e *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// GET /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchema(c echo.Context) error {
	form, err := h.GetFormByID(c)
	if err != nil {
		return h.HandleError(c, err, "Failed to get form schema")
	}

	// Set content type for JSON response
	c.Response().Header().Set("Content-Type", "application/json")

	return c.JSON(constants.StatusOK, form.Schema)
}

// PUT /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchemaUpdate(c echo.Context) error {
	_, err := h.RequireAuthenticatedUser(c)
	if err != nil {
		return h.HandleError(c, err, "Authentication required")
	}

	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return h.HandleError(c, err, "Unauthorized or form not found")
	}

	// Parse schema from request body
	schema, decodeErr := decodeSchema(c)
	if decodeErr != nil {
		return h.HandleError(c, decodeErr, "Failed to decode schema")
	}

	// Update form schema
	form.Schema = schema
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form); updateErr != nil {
		return h.HandleError(c, updateErr, "Failed to update form schema")
	}

	return c.JSON(constants.StatusOK, form.Schema)
}

// POST /api/v1/forms/:id/submit
func (h *FormAPIHandler) HandleFormSubmit(c echo.Context) error {
	form, err := h.GetFormByID(c)
	if err != nil {
		return h.HandleError(c, err, "Failed to get form for submission")
	}

	// Parse submission data
	var submissionData model.JSON
	if decodeErr := json.NewDecoder(c.Request().Body).Decode(&submissionData); decodeErr != nil {
		return h.HandleError(c, decodeErr, "Invalid submission data")
	}

	// Create submission
	submission := &model.FormSubmission{
		FormID:      form.ID,
		Data:        submissionData,
		SubmittedAt: time.Now(),
		Status:      model.SubmissionStatusPending,
	}

	// Submit form
	err = h.FormService.SubmitForm(c.Request().Context(), submission)
	if err != nil {
		return h.HandleError(c, err, "Failed to submit form")
	}

	return c.JSON(constants.StatusOK, map[string]any{
		"success": true,
		"message": "Form submitted successfully",
		"data": map[string]any{
			"submission_id": submission.ID,
			"status":        submission.Status,
		},
	})
}

// decodeSchema decodes the form schema from the request body
func decodeSchema(c echo.Context) (model.JSON, error) {
	var schema model.JSON
	decodeErr := json.NewDecoder(c.Request().Body).Decode(&schema)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode schema: %w", decodeErr)
	}
	return schema, nil
}

// Start initializes the form API handler.
// This is called during application startup.
func (h *FormAPIHandler) Start(ctx context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the form API handler.
// This is called during application shutdown.
func (h *FormAPIHandler) Stop(ctx context.Context) error {
	return nil // No cleanup needed
}
