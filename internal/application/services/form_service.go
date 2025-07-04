package services

import (
	"context"
	"fmt"

	"github.com/goformx/goforms/internal/application/dto"
	"github.com/goformx/goforms/internal/domain/common/interfaces"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
)

// FormUseCaseService handles form use cases
type FormUseCaseService struct {
	formService form.Service
	logger      interfaces.Logger
}

// NewFormUseCaseService creates a new form use case service
func NewFormUseCaseService(
	formService form.Service,
	logger interfaces.Logger,
) *FormUseCaseService {
	return &FormUseCaseService{
		formService: formService,
		logger:      logger,
	}
}

// CreateForm handles form creation use case
func (s *FormUseCaseService) CreateForm(ctx context.Context, request *dto.CreateFormRequest) (*dto.FormResponse, error) {
	s.logger.Info("processing create form request", "user_id", request.UserID, "title", request.Title)

	// Convert DTO to domain model
	formData := &model.Form{
		Title:       request.Title,
		Description: request.Description,
		Schema:      request.Schema,
		UserID:      request.UserID,
		Status:      request.Status,
	}

	// Call domain service
	err := s.formService.CreateForm(ctx, formData)
	if err != nil {
		s.logger.Error("failed to create form", "user_id", request.UserID, "error", err)

		return nil, fmt.Errorf("failed to create form: %w", err)
	}

	// Build response
	response := dto.ConvertFormToResponse(formData)

	s.logger.Info("form created successfully", "form_id", formData.ID, "user_id", request.UserID)

	return &response, nil
}

// UpdateForm handles form update use case
func (s *FormUseCaseService) UpdateForm(ctx context.Context, request *dto.UpdateFormRequest) (*dto.FormResponse, error) {
	s.logger.Info("processing update form request", "form_id", request.ID, "user_id", request.UserID)

	// Get existing form
	existingForm, err := s.formService.GetForm(ctx, request.ID)
	if err != nil {
		s.logger.Error("failed to get form for update", "form_id", request.ID, "error", err)

		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	// Check ownership
	if existingForm.UserID != request.UserID {
		s.logger.Warn("unauthorized form update attempt", "form_id", request.ID, "user_id", request.UserID, "form_owner", existingForm.UserID)

		return nil, fmt.Errorf("unauthorized: you don't have permission to update this form")
	}

	// Update form data
	existingForm.Title = request.Title
	existingForm.Description = request.Description
	existingForm.Schema = request.Schema
	existingForm.Status = request.Status

	// Call domain service
	err = s.formService.UpdateForm(ctx, existingForm)
	if err != nil {
		s.logger.Error("failed to update form", "form_id", request.ID, "error", err)

		return nil, fmt.Errorf("failed to update form: %w", err)
	}

	// Build response
	response := dto.ConvertFormToResponse(existingForm)

	s.logger.Info("form updated successfully", "form_id", request.ID, "user_id", request.UserID)

	return &response, nil
}

// DeleteForm handles form deletion use case
func (s *FormUseCaseService) DeleteForm(ctx context.Context, request *dto.DeleteFormRequest) (*dto.DeleteFormResponse, error) {
	s.logger.Info("processing delete form request", "form_id", request.ID, "user_id", request.UserID)

	// Get existing form
	existingForm, err := s.formService.GetForm(ctx, request.ID)
	if err != nil {
		s.logger.Error("failed to get form for deletion", "form_id", request.ID, "error", err)

		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	// Check ownership
	if existingForm.UserID != request.UserID {
		s.logger.Warn("unauthorized form deletion attempt", "form_id", request.ID, "user_id", request.UserID, "form_owner", existingForm.UserID)

		return nil, fmt.Errorf("unauthorized: you don't have permission to delete this form")
	}

	// Call domain service
	err = s.formService.DeleteForm(ctx, request.ID)
	if err != nil {
		s.logger.Error("failed to delete form", "form_id", request.ID, "error", err)

		return nil, fmt.Errorf("failed to delete form: %w", err)
	}

	response := &dto.DeleteFormResponse{
		Message: "Form deleted successfully",
	}

	s.logger.Info("form deleted successfully", "form_id", request.ID, "user_id", request.UserID)

	return response, nil
}

// GetForm handles form retrieval use case
func (s *FormUseCaseService) GetForm(ctx context.Context, formID string) (*dto.FormResponse, error) {
	s.logger.Debug("processing get form request", "form_id", formID)

	// Call domain service
	formData, err := s.formService.GetForm(ctx, formID)
	if err != nil {
		s.logger.Error("failed to get form", "form_id", formID, "error", err)

		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	// Build response
	response := dto.ConvertFormToResponse(formData)

	return &response, nil
}

// ListForms handles form listing use case
func (s *FormUseCaseService) ListForms(ctx context.Context, userID string, pagination *dto.PaginationRequest) (*dto.FormListResponse, error) {
	s.logger.Debug("processing list forms request", "user_id", userID, "page", pagination.Page, "limit", pagination.Limit)

	// Call domain service
	forms, err := s.formService.ListForms(ctx, userID)
	if err != nil {
		s.logger.Error("failed to list forms", "user_id", userID, "error", err)

		return nil, fmt.Errorf("failed to list forms: %w", err)
	}

	// Convert to response DTOs
	formResponses := dto.ConvertFormListToResponse(forms)

	// TODO: Get total count from domain service
	total := len(formResponses) // This should come from a separate count query

	response := &dto.FormListResponse{
		Forms: formResponses,
		Total: total,
		Page:  pagination.Page,
		Limit: pagination.Limit,
	}

	return response, nil
}

// SubmitForm handles form submission use case
func (s *FormUseCaseService) SubmitForm(ctx context.Context, request *dto.SubmitFormRequest) (*dto.SubmitFormResponse, error) {
	s.logger.Info("processing form submission", "form_id", request.FormID, "user_id", request.UserID)

	// Create submission model
	submission := &model.FormSubmission{
		FormID: request.FormID,
		Data:   model.JSON(request.Data),
	}

	// Call domain service
	err := s.formService.SubmitForm(ctx, submission)
	if err != nil {
		s.logger.Error("failed to submit form", "form_id", request.FormID, "error", err)

		return nil, fmt.Errorf("failed to submit form: %w", err)
	}

	response := &dto.SubmitFormResponse{
		SubmissionID: submission.ID,
		FormID:       request.FormID,
		Data:         submission.Data,
		SubmittedAt:  submission.CreatedAt,
	}

	s.logger.Info("form submitted successfully", "submission_id", submission.ID, "form_id", request.FormID)

	return response, nil
}

// GetFormSchema handles form schema retrieval use case
func (s *FormUseCaseService) GetFormSchema(ctx context.Context, formID string) (*dto.FormSchemaResponse, error) {
	s.logger.Debug("processing get form schema request", "form_id", formID)

	// Call domain service
	formData, err := s.formService.GetForm(ctx, formID)
	if err != nil {
		s.logger.Error("failed to get form schema", "form_id", formID, "error", err)

		return nil, fmt.Errorf("failed to get form schema: %w", err)
	}

	response := &dto.FormSchemaResponse{
		ID:     formData.ID,
		Schema: formData.Schema,
	}

	return response, nil
}
