// Package web provides HTTP handlers for web-based functionality including
// authentication, form management, and user interface components.
package web

import (
	"context"
	"fmt"
	"strings"

	formdomain "github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// FormService handles form-related business logic
type FormService struct {
	formService formdomain.Service
	logger      logging.Logger
}

// NewFormService creates a new FormService instance
func NewFormService(formService formdomain.Service, logger logging.Logger) *FormService {
	return &FormService{
		formService: formService,
		logger:      logger,
	}
}

// CreateForm creates a new form with the given request data
func (s *FormService) CreateForm(ctx context.Context, userID string, req *FormCreateRequest) (*model.Form, error) {
	schema := model.JSON{
		"type": "object",
		"components": []any{
			map[string]any{
				"type":  "button",
				"key":   "submit",
				"label": "Submit",
				"input": true,
			},
		},
	}

	form := model.NewForm(userID, req.Title, "", schema)

	if err := s.formService.CreateForm(ctx, form); err != nil {
		return nil, fmt.Errorf("create form: %w", err)
	}

	return form, nil
}

// UpdateForm updates an existing form with the given request data
func (s *FormService) UpdateForm(ctx context.Context, form *model.Form, req *FormUpdateRequest) error {
	form.Title = req.Title
	form.Description = req.Description
	form.Status = req.Status

	if req.CorsOrigins != "" {
		form.CorsOrigins = model.JSON{"origins": parseCSV(req.CorsOrigins)}
	}

	if req.Schema != nil {
		form.Schema = req.Schema
	}

	if err := s.formService.UpdateForm(ctx, form); err != nil {
		return fmt.Errorf("update form: %w", err)
	}

	return nil
}

// DeleteForm deletes a form by ID
func (s *FormService) DeleteForm(ctx context.Context, formID string) error {
	if err := s.formService.DeleteForm(ctx, formID); err != nil {
		return fmt.Errorf("delete form: %w", err)
	}

	return nil
}

// GetFormSubmissions retrieves submissions for a form
func (s *FormService) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	submissions, err := s.formService.ListFormSubmissions(ctx, formID)
	if err != nil {
		return nil, fmt.Errorf("list form submissions: %w", err)
	}

	return submissions, nil
}

// LogFormAccess logs form access for debugging
func (s *FormService) LogFormAccess(form *model.Form) {
	s.logger.Debug("Form access",
		"form_id", form.ID,
		"form_title", form.Title,
		"form_status", form.Status,
	)
}

// parseCSV parses a comma-separated string into a slice of strings, trimming whitespace and skipping empty values
func parseCSV(input string) []string {
	if input == "" {
		return []string{} // Return empty slice instead of nil
	}

	parts := strings.Split(input, ",")

	var result []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
