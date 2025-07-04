package forms

import (
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/entities"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/web"
	"github.com/goformx/goforms/internal/presentation/handlers"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// FormHandler handles all form-related routes
// Implements httpiface.Handler
type FormHandler struct {
	handlers.BaseHandler
	formService     form.Service
	sessionManager  *session.Manager
	renderer        view.Renderer
	config          *config.Config
	assetManager    web.AssetManagerInterface
	logger          logging.Logger
	requestParser   *FormRequestParser
	responseBuilder *FormResponseBuilder
}

// NewFormHandler creates a new FormHandler and registers all form routes
func NewFormHandler(
	formService form.Service,
	sessionManager *session.Manager,
	renderer view.Renderer,
	config *config.Config,
	assetManager web.AssetManagerInterface,
	logger logging.Logger,
) *FormHandler {
	h := &FormHandler{
		BaseHandler:     *handlers.NewBaseHandler("forms"),
		formService:     formService,
		sessionManager:  sessionManager,
		renderer:        renderer,
		config:          config,
		assetManager:    assetManager,
		logger:          logger,
		requestParser:   NewFormRequestParser(),
		responseBuilder: NewFormResponseBuilder(config, assetManager, renderer, logger),
	}

	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/forms/new",
		Handler: h.NewForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "POST",
		Path:    "/forms",
		Handler: h.CreateForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/forms/:id/edit",
		Handler: h.EditForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "PUT",
		Path:    "/forms/:id",
		Handler: h.UpdateForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "DELETE",
		Path:    "/forms/:id",
		Handler: h.DeleteForm,
	})
	h.AddRoute(httpiface.Route{
		Method:  "GET",
		Path:    "/forms/:id/submissions",
		Handler: h.FormSubmissions,
	})

	return h
}

// NewForm handles GET /forms/new
func (h *FormHandler) NewForm(ctx httpiface.Context) error {
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")
		return fmt.Errorf("internal server error: context conversion failed")
	}

	user, err := h.getUserFromSession(echoCtx)
	if err != nil {
		h.logger.Warn("authentication required for new form access", "error", err)
		return h.responseBuilder.BuildAuthenticationErrorResponse(echoCtx)
	}

	return h.responseBuilder.BuildNewFormResponse(echoCtx, user)
}

// CreateForm handles POST /forms
func (h *FormHandler) CreateForm(ctx httpiface.Context) error {
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")
		return fmt.Errorf("internal server error: context conversion failed")
	}

	user, err := h.getUserFromSession(echoCtx)
	if err != nil {
		h.logger.Warn("authentication required for form creation", "error", err)
		return h.responseBuilder.BuildAuthenticationErrorResponse(echoCtx)
	}

	// Parse form creation data
	formData, err := h.requestParser.ParseCreateForm(echoCtx)
	if err != nil {
		h.logger.Error("failed to parse create form request", "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Invalid request format", 400)
	}

	// Validate form data
	if err := h.requestParser.ValidateCreateForm(formData); err != nil {
		return h.responseBuilder.BuildValidationErrorResponse(echoCtx, "form", err.Error())
	}

	// Set user ID and create form
	formData.UserID = user.ID
	formData.Status = "draft" // Default status

	// Create form using service
	if err := h.formService.CreateForm(echoCtx.Request().Context(), formData); err != nil {
		h.logger.Error("failed to create form", "user_id", user.ID, "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Failed to create form. Please try again.", 500)
	}

	return h.responseBuilder.BuildCreateFormSuccessResponse(echoCtx, formData)
}

// EditForm handles GET /forms/:id/edit
func (h *FormHandler) EditForm(ctx httpiface.Context) error {
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")
		return fmt.Errorf("internal server error: context conversion failed")
	}

	user, err := h.getUserFromSession(echoCtx)
	if err != nil {
		h.logger.Warn("authentication required for edit form access", "error", err)
		return h.responseBuilder.BuildAuthenticationErrorResponse(echoCtx)
	}

	// Parse form ID
	formID, err := h.requestParser.ParseFormID(echoCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Invalid form ID", 400)
	}

	// Get form from service
	form, err := h.formService.GetForm(echoCtx.Request().Context(), formID)
	if err != nil {
		h.logger.Error("failed to get form", "form_id", formID, "error", err)
		return h.responseBuilder.BuildFormNotFoundResponse(echoCtx)
	}

	// Check form ownership
	if form.UserID != user.ID {
		h.logger.Warn("unauthorized form access", "user_id", user.ID, "form_user_id", form.UserID, "form_id", formID)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "You don't have permission to edit this form", 403)
	}

	return h.responseBuilder.BuildEditFormResponse(echoCtx, user, form)
}

// UpdateForm handles PUT /forms/:id
func (h *FormHandler) UpdateForm(ctx httpiface.Context) error {
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")
		return fmt.Errorf("internal server error: context conversion failed")
	}

	user, err := h.getUserFromSession(echoCtx)
	if err != nil {
		h.logger.Warn("authentication required for form update", "error", err)
		return h.responseBuilder.BuildAuthenticationErrorResponse(echoCtx)
	}

	// Parse form ID
	formID, err := h.requestParser.ParseFormID(echoCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Invalid form ID", 400)
	}

	// Get existing form
	existingForm, err := h.formService.GetForm(echoCtx.Request().Context(), formID)
	if err != nil {
		h.logger.Error("failed to get form for update", "form_id", formID, "error", err)
		return h.responseBuilder.BuildFormNotFoundResponse(echoCtx)
	}

	// Check form ownership
	if existingForm.UserID != user.ID {
		h.logger.Warn("unauthorized form update", "user_id", user.ID, "form_user_id", existingForm.UserID, "form_id", formID)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "You don't have permission to update this form", 403)
	}

	// Parse update data
	updateData, err := h.requestParser.ParseUpdateForm(echoCtx)
	if err != nil {
		h.logger.Error("failed to parse update form request", "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Invalid request format", 400)
	}

	// Validate update data
	if err := h.requestParser.ValidateUpdateForm(updateData); err != nil {
		return h.responseBuilder.BuildValidationErrorResponse(echoCtx, "form", err.Error())
	}

	// Update form fields
	existingForm.Title = updateData.Title
	existingForm.Description = updateData.Description
	if updateData.Status != "" {
		existingForm.Status = updateData.Status
	}
	if updateData.Schema != nil {
		existingForm.Schema = updateData.Schema
	}

	// Update form using service
	if err := h.formService.UpdateForm(echoCtx.Request().Context(), existingForm); err != nil {
		h.logger.Error("failed to update form", "form_id", formID, "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Failed to update form. Please try again.", 500)
	}

	return h.responseBuilder.BuildUpdateFormSuccessResponse(echoCtx, existingForm)
}

// DeleteForm handles DELETE /forms/:id
func (h *FormHandler) DeleteForm(ctx httpiface.Context) error {
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")
		return fmt.Errorf("internal server error: context conversion failed")
	}

	user, err := h.getUserFromSession(echoCtx)
	if err != nil {
		h.logger.Warn("authentication required for form deletion", "error", err)
		return h.responseBuilder.BuildAuthenticationErrorResponse(echoCtx)
	}

	// Parse form ID
	formID, err := h.requestParser.ParseFormID(echoCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Invalid form ID", 400)
	}

	// Get form to check ownership
	form, err := h.formService.GetForm(echoCtx.Request().Context(), formID)
	if err != nil {
		h.logger.Error("failed to get form for deletion", "form_id", formID, "error", err)
		return h.responseBuilder.BuildFormNotFoundResponse(echoCtx)
	}

	// Check form ownership
	if form.UserID != user.ID {
		h.logger.Warn("unauthorized form deletion", "user_id", user.ID, "form_user_id", form.UserID, "form_id", formID)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "You don't have permission to delete this form", 403)
	}

	// Delete form using service
	if err := h.formService.DeleteForm(echoCtx.Request().Context(), formID); err != nil {
		h.logger.Error("failed to delete form", "form_id", formID, "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Failed to delete form. Please try again.", 500)
	}

	return h.responseBuilder.BuildDeleteFormSuccessResponse(echoCtx)
}

// FormSubmissions handles GET /forms/:id/submissions
func (h *FormHandler) FormSubmissions(ctx httpiface.Context) error {
	echoCtx, ok := ctx.Request().(echo.Context)
	if !ok {
		h.logger.Error("failed to get echo context from httpiface.Context")
		return fmt.Errorf("internal server error: context conversion failed")
	}

	user, err := h.getUserFromSession(echoCtx)
	if err != nil {
		h.logger.Warn("authentication required for form submissions access", "error", err)
		return h.responseBuilder.BuildAuthenticationErrorResponse(echoCtx)
	}

	// Parse form ID
	formID, err := h.requestParser.ParseFormID(echoCtx)
	if err != nil {
		h.logger.Error("failed to parse form ID", "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Invalid form ID", 400)
	}

	// Get form to check ownership
	form, err := h.formService.GetForm(echoCtx.Request().Context(), formID)
	if err != nil {
		h.logger.Error("failed to get form for submissions", "form_id", formID, "error", err)
		return h.responseBuilder.BuildFormNotFoundResponse(echoCtx)
	}

	// Check form ownership
	if form.UserID != user.ID {
		h.logger.Warn("unauthorized form submissions access", "user_id", user.ID, "form_user_id", form.UserID, "form_id", formID)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "You don't have permission to view submissions for this form", 403)
	}

	// Get form submissions
	submissions, err := h.formService.ListFormSubmissions(echoCtx.Request().Context(), formID)
	if err != nil {
		h.logger.Error("failed to get form submissions", "form_id", formID, "error", err)
		return h.responseBuilder.BuildFormErrorResponse(echoCtx, "Failed to load form submissions. Please try again.", 500)
	}

	return h.responseBuilder.BuildFormSubmissionsResponse(echoCtx, user, form, submissions)
}

// getUserFromSession extracts user information from the session
func (h *FormHandler) getUserFromSession(c echo.Context) (*entities.User, error) {
	// Get session cookie
	cookie, err := c.Cookie(h.sessionManager.GetCookieName())
	if err != nil {
		return nil, fmt.Errorf("no session cookie found")
	}

	// Get session from manager
	session, exists := h.sessionManager.GetSession(cookie.Value)
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	// Create user entity from session data
	user := &entities.User{
		ID:    session.UserID,
		Email: session.Email,
		Role:  session.Role,
	}

	return user, nil
}
