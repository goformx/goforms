package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/application/middleware/access"
	"github.com/goformx/goforms/internal/application/middleware/assertion"
	"github.com/goformx/goforms/internal/application/middleware/security"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/application/validation"
	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// FormAPIHandler handles API form operations
type FormAPIHandler struct {
	*FormBaseHandler
	AccessManager          *access.Manager
	RequestProcessor       FormRequestProcessor
	ResponseBuilder        FormResponseBuilder
	ErrorHandler           FormErrorHandler
	ComprehensiveValidator *validation.ComprehensiveValidator
	FormServiceHandler     *FormService
	AssertionMiddleware    *assertion.Middleware
	UserEnsurer            user.UserEnsurer
}

// NewFormAPIHandler creates a new FormAPIHandler.
func NewFormAPIHandler(
	base *BaseHandler,
	formService formdomain.Service,
	accessManager *access.Manager,
	formValidator *validation.FormValidator,
	sanitizer sanitization.ServiceInterface,
	userEnsurer user.UserEnsurer,
) *FormAPIHandler {
	// Create dependencies
	requestProcessor := NewFormRequestProcessor(sanitizer, formValidator, base.Logger)
	responseBuilder := NewFormResponseBuilder()
	errorHandler := NewFormErrorHandler(responseBuilder)
	comprehensiveValidator := validation.NewComprehensiveValidator()
	formServiceHandler := NewFormService(formService, base.Logger)
	assertionMiddleware := assertion.NewMiddleware(base.Config, base.Logger)

	return &FormAPIHandler{
		FormBaseHandler:        NewFormBaseHandler(base, formService, formValidator),
		AccessManager:          accessManager,
		RequestProcessor:       requestProcessor,
		ResponseBuilder:        responseBuilder,
		ErrorHandler:           errorHandler,
		ComprehensiveValidator: comprehensiveValidator,
		FormServiceHandler:     formServiceHandler,
		AssertionMiddleware:    assertionMiddleware,
		UserEnsurer:            userEnsurer,
	}
}

// RegisterRoutes registers API routes for forms.
func (h *FormAPIHandler) RegisterRoutes(e *echo.Echo) {
	// Laravel API routes with assertion auth
	h.RegisterLaravelRoutes(e)

	// Public /forms routes for embed (schema, validation, submit, embed HTML)
	h.RegisterPublicFormsRoutes(e)
}

// RegisterLaravelRoutes registers /api/forms routes with assertion middleware for Laravel proxy.
func (h *FormAPIHandler) RegisterLaravelRoutes(e *echo.Echo) {
	formsLaravel := e.Group(constants.PathAPIFormsLaravel)
	formsLaravel.Use(h.AssertionMiddleware.Verify())
	formsLaravel.Use(h.ensureUserMiddleware())

	formsLaravel.GET("", h.handleListForms)
	formsLaravel.POST("", h.handleCreateForm)
	formsLaravel.GET("/:id", h.handleGetForm)
	formsLaravel.PUT("/:id", h.handleUpdateForm)
	formsLaravel.DELETE("/:id", h.handleDeleteForm)
	formsLaravel.GET("/:id/submissions", h.handleListSubmissions)
	formsLaravel.GET("/:id/submissions/:sid", h.handleGetSubmission)
}

// ensureUserMiddleware returns middleware that lazily syncs the Laravel user to a Go shadow row.
// Runs after assertion verification so user_id is available in the context.
func (h *FormAPIHandler) ensureUserMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(string)
			if !ok {
				return next(c)
			}
			if err := h.UserEnsurer.EnsureUser(c.Request().Context(), userID); err != nil {
				h.Logger.Error("failed to ensure Laravel user",
					"user_id", h.Logger.SanitizeField("user_id", userID), "error", err)
				return h.HandleError(c, err, "Failed to ensure user")
			}
			return next(c)
		}
	}
}

// RegisterPublicFormsRoutes registers public routes at /forms/:id/... for cleaner embed URLs.
// These routes bypass the /api/v1 prefix and are intended for cross-origin embedding.
func (h *FormAPIHandler) RegisterPublicFormsRoutes(e *echo.Echo) {
	formsPublic := e.Group(constants.PathFormsPublic)
	formsPublic.Use(NewFormCORSMiddleware(h.FormService, h.Config.Security.CORS))

	// Apply API key middleware if enabled (same as /api/v1/forms)
	if h.Config.Security.APIKey.Enabled {
		apiKeyAuth := security.NewAPIKeyAuth(h.Logger, h.Config)
		formsPublic.Use(apiKeyAuth.Setup())
	}

	formsPublic.GET("/:id/schema", h.handleFormSchema)
	formsPublic.GET("/:id/validation", h.handleFormValidationSchema)
	formsPublic.POST("/:id/submit", h.handleFormSubmit)
	formsPublic.GET("/:id/embed", h.handleFormEmbed)
}

// Register registers the FormAPIHandler with the Echo instance.
func (h *FormAPIHandler) Register(_ *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// GET /api/v1/forms
func (h *FormAPIHandler) handleListForms(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return h.HandleForbidden(c, "User not authenticated")
	}

	// Get forms for the user
	forms, err := h.FormService.ListForms(c.Request().Context(), userID)
	if err != nil {
		h.Logger.Error("failed to list forms", "error", err)

		return h.HandleError(c, err, "Failed to list forms")
	}

	h.Logger.Debug("forms listed successfully",
		"user_id", h.Logger.SanitizeField("user_id", userID),
		"form_count", len(forms))

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildFormListResponse(c, forms); respErr != nil {
		h.Logger.Error("failed to build form list response", "error", respErr)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// GET /api/v1/forms/:id
func (h *FormAPIHandler) handleGetForm(c echo.Context) error {
	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildFormResponse(c, form); respErr != nil {
		h.Logger.Error("failed to build form response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// GET /api/v1/forms/:id/schema
func (h *FormAPIHandler) handleFormSchema(c echo.Context) error {
	form, err := h.getFormOrError(c)
	if err != nil {
		return err
	}

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildSchemaResponse(c, form.Schema); respErr != nil {
		h.Logger.Error("failed to build schema response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// GET /api/v1/forms/:id/validation
func (h *FormAPIHandler) handleFormValidationSchema(c echo.Context) error {
	form, err := h.getFormOrError(c)
	if err != nil {
		return err
	}

	if validationErr := h.validateFormSchema(c, form); validationErr != nil {
		return validationErr
	}

	// Generate client-side validation rules from form schema
	clientValidation, err := h.ComprehensiveValidator.GenerateClientValidation(form.Schema)
	if err != nil {
		h.Logger.Error("failed to generate client validation schema", "error", err, "form_id", form.ID)

		return h.wrapError("handle schema error", h.ErrorHandler.HandleSchemaError(c, err))
	}

	return response.Success(c, clientValidation)
}

// POST /api/forms - create form (assertion auth)
func (h *FormAPIHandler) handleCreateForm(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return h.HandleForbidden(c, "User not authenticated")
	}

	req, err := h.RequestProcessor.ProcessCreateRequest(c)
	if err != nil {
		return h.wrapError("handle create error", h.ErrorHandler.HandleSchemaError(c, err))
	}

	form, err := h.FormServiceHandler.CreateForm(c.Request().Context(), userID, req)
	if err != nil {
		h.Logger.Error("failed to create form", "error", err)

		return h.HandleError(c, err, "Failed to create form")
	}

	h.Logger.Debug("form created successfully", "form_id", form.ID, "user_id", h.Logger.SanitizeField("user_id", userID))

	return c.JSON(http.StatusCreated, response.APIResponse{
		Success: true,
		Message: "Form created successfully",
		Data: map[string]any{
			"form": map[string]any{
				"id":          form.ID,
				"title":       form.Title,
				"description": form.Description,
				"status":      form.Status,
				"schema":      form.Schema,
				"created_at":  form.CreatedAt.Format(time.RFC3339),
				"updated_at":  form.UpdatedAt.Format(time.RFC3339),
			},
		},
	})
}

// PUT /api/forms/:id - update form (assertion auth)
func (h *FormAPIHandler) handleUpdateForm(c echo.Context) error {
	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	req, err := h.RequestProcessor.ProcessUpdateRequest(c)
	if err != nil {
		return h.wrapError("handle update error", h.ErrorHandler.HandleSchemaError(c, err))
	}

	if updateErr := h.FormServiceHandler.UpdateForm(c.Request().Context(), form, req); updateErr != nil {
		h.Logger.Error("failed to update form", "error", updateErr, "form_id", form.ID)

		return h.HandleError(c, updateErr, "Failed to update form")
	}

	// Reload form to get updated schema if it was changed
	updatedForm, getErr := h.FormService.GetForm(c.Request().Context(), form.ID)
	if getErr != nil || updatedForm == nil {
		h.Logger.Warn("failed to reload form after update, returning pre-update data", "form_id", form.ID, "error", getErr)

		updatedForm = form
	}

	if respErr := h.ResponseBuilder.BuildFormResponse(c, updatedForm); respErr != nil {
		h.Logger.Error("failed to build form response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// DELETE /api/forms/:id - delete form (assertion auth)
func (h *FormAPIHandler) handleDeleteForm(c echo.Context) error {
	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	if deleteErr := h.FormServiceHandler.DeleteForm(c.Request().Context(), form.ID); deleteErr != nil {
		h.Logger.Error("failed to delete form", "error", deleteErr, "form_id", form.ID)

		return h.HandleError(c, deleteErr, "Failed to delete form")
	}

	return c.JSON(http.StatusNoContent, nil)
}

// GET /api/forms/:id/submissions - list submissions (assertion auth)
func (h *FormAPIHandler) handleListSubmissions(c echo.Context) error {
	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	submissions, err := h.FormServiceHandler.GetFormSubmissions(c.Request().Context(), form.ID)
	if err != nil {
		h.Logger.Error("failed to list form submissions", "error", err, "form_id", form.ID)

		return h.HandleError(c, err, "Failed to list submissions")
	}

	if respErr := h.ResponseBuilder.BuildSubmissionListResponse(c, submissions); respErr != nil {
		h.Logger.Error("failed to build submission list response", "error", respErr, "form_id", form.ID)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// GET /api/forms/:id/submissions/:sid - get submission (assertion auth)
func (h *FormAPIHandler) handleGetSubmission(c echo.Context) error {
	form, err := h.getFormWithOwnershipOrError(c)
	if err != nil {
		return err
	}

	submissionID := c.Param("sid")
	if submissionID == "" {
		return h.ResponseBuilder.BuildNotFoundResponse(c, "Submission")
	}

	submission, err := h.FormService.GetFormSubmission(c.Request().Context(), submissionID)
	if err != nil {
		h.Logger.Error("failed to get submission", "error", err, "form_id", form.ID, "submission_id", submissionID)

		return h.HandleError(c, err, "Failed to get submission")
	}

	if submission == nil || submission.FormID != form.ID {
		return h.ResponseBuilder.BuildNotFoundResponse(c, "Submission")
	}

	return c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Data: map[string]any{
			"id":           submission.ID,
			"form_id":      submission.FormID,
			"status":       submission.Status,
			"submitted_at": submission.SubmittedAt.Format(time.RFC3339),
			"data":         submission.Data,
		},
	})
}

// GET /forms/:id/embed returns a minimal HTML page for embedding the form via iframe.
// Loads Form.io from CDN and renders the form, posting to /forms/:id/submit.
func (h *FormAPIHandler) handleFormEmbed(c echo.Context) error {
	form, err := h.getFormOrError(c)
	if err != nil {
		return err
	}

	if form.Schema == nil {
		h.Logger.Warn("form schema is nil for embed", "form_id", form.ID)

		return h.wrapError("handle embed error",
			h.ErrorHandler.HandleSchemaError(c, errors.New("form schema is required")))
	}

	formID := form.ID
	schemaURL := "/forms/" + formID + "/schema"
	submitURL := "/forms/" + formID + "/submit"

	html := `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + escapeHTML(form.Title) + `</title>
  <link rel="stylesheet" href="https://cdn.form.io/formiojs/formio.full.min.css">
</head>
<body>
  <div id="formio"></div>
  <script src="https://cdn.form.io/formiojs/formio.full.min.js"></script>
  <script>
    (function() {
      var schemaUrl = '` + schemaURL + `';
      var submitUrl = '` + submitURL + `';
      var container = document.getElementById('formio');
      Formio.createForm(container, schemaUrl, {
        submit: submitUrl,
        noSubmit: false
      }).then(function(form) {
        form.on('submit', function(submission) {
          if (submission && submission.submission) {
            window.parent.postMessage({ type: 'goformx:submitted', submission: submission.submission }, '*');
          }
        });
      }).catch(function(err) {
        container.innerHTML = '<p style="color: #dc2626;">Failed to load form. Please try again.</p>';
        console.error('Form.io load error:', err);
      });
    })();
  </script>
</body>
</html>`

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")

	return c.HTML(http.StatusOK, html)
}

// escapeHTML escapes HTML special characters for safe inclusion in attribute values.
func escapeHTML(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	).Replace(s)
}

// POST /api/v1/forms/:id/submit
func (h *FormAPIHandler) handleFormSubmit(c echo.Context) error {
	formID := c.Param("id")
	h.logFormSubmissionRequest(c, formID)

	form, err := h.getFormOrError(c)
	if err != nil {
		return err
	}

	if validationErr := h.validateFormSchema(c, form); validationErr != nil {
		return validationErr
	}

	submissionData, err := h.processSubmissionRequest(c, form.ID)
	if err != nil {
		return err
	}

	if validationDataErr := h.validateSubmissionData(c, form, submissionData); validationDataErr != nil {
		return validationDataErr
	}

	submission, err := h.createAndSubmitForm(c, form, submissionData)
	if err != nil {
		return err
	}

	h.Logger.Info("Form submitted successfully", "form_id", form.ID, "submission_id", submission.ID)

	// Build response with proper error checking
	if respErr := h.ResponseBuilder.BuildSubmissionResponse(c, submission); respErr != nil {
		h.Logger.Error(
			"failed to build submission response",
			"error", respErr,
			"form_id", form.ID,
			"submission_id", submission.ID,
		)

		return h.HandleError(c, respErr, "Failed to build response")
	}

	return nil
}

// Start initializes the form API handler.
// This is called during application startup.
func (h *FormAPIHandler) Start(_ context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the form API handler.
// This is called during application shutdown.
func (h *FormAPIHandler) Stop(_ context.Context) error {
	return nil // No cleanup needed
}

// Helper methods to reduce code duplication and improve SRP

// getFormOrError retrieves a form by ID and handles common error cases
func (h *FormAPIHandler) getFormOrError(c echo.Context) (*model.Form, error) {
	form, err := h.GetFormByID(c)
	if err != nil {
		return nil, h.HandleError(c, err, "Failed to get form")
	}

	if form == nil {
		h.Logger.Error("form is nil after GetFormByID", "form_id", c.Param("id"))

		return nil, h.wrapError("handle form not found", h.ErrorHandler.HandleFormNotFoundError(c, ""))
	}

	return form, nil
}

// getFormWithOwnershipOrError retrieves a form with ownership verification
func (h *FormAPIHandler) getFormWithOwnershipOrError(c echo.Context) (*model.Form, error) {
	form, err := h.GetFormWithOwnership(c)
	if err != nil {
		return nil, h.HandleError(c, err, "Failed to get form")
	}

	if form == nil {
		h.Logger.Error("form is nil after GetFormWithOwnership", "form_id", c.Param("id"))

		return nil, h.wrapError("handle form not found", h.ErrorHandler.HandleFormNotFoundError(c, ""))
	}

	return form, nil
}

// validateFormSchema validates that form schema exists
func (h *FormAPIHandler) validateFormSchema(c echo.Context, form *model.Form) error {
	if form.Schema == nil {
		h.Logger.Warn("form schema is nil", "form_id", form.ID)

		return h.wrapError("handle submission error",
			h.ErrorHandler.HandleSchemaError(c, errors.New("form schema is required")))
	}

	return nil
}

// logFormSubmissionRequest logs the initial form submission request
func (h *FormAPIHandler) logFormSubmissionRequest(c echo.Context, formID string) {
	h.Logger.Debug("Form submission request received",
		"form_id", formID,
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"content_type", c.Request().Header.Get("Content-Type"),
		"csrf_token_present", c.Request().Header.Get("X-Csrf-Token") != "",
		"user_agent", c.Request().UserAgent())
}

// processSubmissionRequest processes and validates the submission request
func (h *FormAPIHandler) processSubmissionRequest(c echo.Context, formID string) (model.JSON, error) {
	submissionData, err := h.RequestProcessor.ProcessSubmissionRequest(c)
	if err != nil {
		h.Logger.Error("Failed to process submission request", "form_id", formID, "error", err)

		return nil, h.wrapError("handle submission error", h.ErrorHandler.HandleSubmissionError(c, err))
	}

	h.Logger.Debug("Submission data processed successfully", "form_id", formID, "data_keys", len(submissionData))

	return submissionData, nil
}

// validateSubmissionData validates submission data against form schema
func (h *FormAPIHandler) validateSubmissionData(c echo.Context, form *model.Form, submissionData model.JSON) error {
	validationResult := h.ComprehensiveValidator.ValidateForm(form.Schema, submissionData)
	if !validationResult.IsValid {
		h.Logger.Warn("Form validation failed", "form_id", form.ID, "error_count", len(validationResult.Errors))

		return h.wrapError("build multiple error response",
			h.ResponseBuilder.BuildMultipleErrorResponse(c, validationResult.Errors))
	}

	h.Logger.Debug("Form validation passed", "form_id", form.ID)

	return nil
}

// createAndSubmitForm creates and submits the form
func (h *FormAPIHandler) createAndSubmitForm(
	c echo.Context,
	form *model.Form,
	submissionData model.JSON,
) (*model.FormSubmission, error) {
	submission := &model.FormSubmission{
		FormID:      form.ID,
		Data:        submissionData,
		SubmittedAt: time.Now(),
		Status:      model.SubmissionStatusPending,
	}

	err := h.FormService.SubmitForm(c.Request().Context(), submission)
	if err != nil {
		h.Logger.Error("Failed to submit form", "form_id", form.ID, "submission_id", submission.ID, "error", err)

		return nil, h.wrapError("handle submission error", h.ErrorHandler.HandleSubmissionError(c, err))
	}

	return submission, nil
}

// wrapError provides consistent error wrapping
func (h *FormAPIHandler) wrapError(ctx string, err error) error {
	return fmt.Errorf("%s: %w", ctx, err)
}
