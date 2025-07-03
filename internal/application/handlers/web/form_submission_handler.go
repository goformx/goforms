package web

import (
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/labstack/echo/v4"
)

// Helper methods for form submission and validation
// These are shared between public and authenticated API handlers

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

// updateFormSchema updates the form schema in the database
func (h *FormAPIHandler) updateFormSchema(c echo.Context, form *model.Form, schema model.JSON) error {
	form.Schema = schema
	if updateErr := h.FormService.UpdateForm(c.Request().Context(), form); updateErr != nil {
		h.Logger.Error("failed to update form schema", "error", updateErr)

		return h.wrapError("handle schema update error", h.ErrorHandler.HandleSchemaError(c, updateErr))
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
