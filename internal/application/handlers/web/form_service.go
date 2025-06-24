// internal/application/handlers/web/form_service.go
package web

import (
	"context"
	"strings"

	"github.com/goformx/goforms/internal/domain/common/types"
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
		"type":       "object",
		"components": []any{},
	}

	form := model.NewForm(userID, req.Title, req.Description, schema)

	if req.CorsOrigins != "" {
		form.CorsOrigins = types.StringArray(parseCSV(req.CorsOrigins))
	}

	return form, s.formService.CreateForm(ctx, form)
}

// UpdateForm updates an existing form with the given request data
func (s *FormService) UpdateForm(ctx context.Context, form *model.Form, req *FormUpdateRequest) error {
	form.Title = req.Title
	form.Description = req.Description
	form.Status = req.Status

	if req.CorsOrigins != "" {
		form.CorsOrigins = types.StringArray(parseCSV(req.CorsOrigins))
	}

	return s.formService.UpdateForm(ctx, form)
}

// DeleteForm deletes a form by ID
func (s *FormService) DeleteForm(ctx context.Context, formID string) error {
	return s.formService.DeleteForm(ctx, formID)
}

// GetFormSubmissions retrieves submissions for a form
func (s *FormService) GetFormSubmissions(ctx context.Context, formID string) ([]*model.FormSubmission, error) {
	return s.formService.ListFormSubmissions(ctx, formID)
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
