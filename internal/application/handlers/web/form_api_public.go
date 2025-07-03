package web

import (
	"github.com/goformx/goforms/internal/application/response"
	"github.com/labstack/echo/v4"
)

// Public API endpoints for forms

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
