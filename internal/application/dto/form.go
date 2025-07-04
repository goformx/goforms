package dto

import (
	"time"

	"github.com/goformx/goforms/internal/domain/form/model"
)

// CreateFormRequest represents a form creation request
type CreateFormRequest struct {
	Title       string                 `json:"title" validate:"required,min=1,max=255"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema" validate:"required"`
	UserID      string                 `json:"user_id" validate:"required"`
	Status      string                 `json:"status"`
}

// UpdateFormRequest represents a form update request
type UpdateFormRequest struct {
	ID          string                 `json:"id" validate:"required"`
	Title       string                 `json:"title" validate:"required,min=1,max=255"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema" validate:"required"`
	UserID      string                 `json:"user_id" validate:"required"`
	Status      string                 `json:"status"`
}

// FormResponse represents a form response
type FormResponse struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	UserID      string                 `json:"user_id"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// FormListResponse represents a list of forms
type FormListResponse struct {
	Forms []FormResponse `json:"forms"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// DeleteFormRequest represents a form deletion request
type DeleteFormRequest struct {
	ID     string `json:"id" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}

// DeleteFormResponse represents a successful form deletion response
type DeleteFormResponse struct {
	Message string `json:"message"`
}

// SubmitFormRequest represents a form submission request
type SubmitFormRequest struct {
	FormID string                 `json:"form_id" validate:"required"`
	Data   map[string]interface{} `json:"data" validate:"required"`
	UserID string                 `json:"user_id,omitempty"`
}

// SubmitFormResponse represents a successful form submission response
type SubmitFormResponse struct {
	SubmissionID string                 `json:"submission_id"`
	FormID       string                 `json:"form_id"`
	Data         map[string]interface{} `json:"data"`
	SubmittedAt  time.Time              `json:"submitted_at"`
}

// FormSchemaResponse represents a form schema response
type FormSchemaResponse struct {
	ID     string                 `json:"id"`
	Schema map[string]interface{} `json:"schema"`
}

// FormValidationSchemaResponse represents a form validation schema response
type FormValidationSchemaResponse struct {
	FormID string                 `json:"form_id"`
	Schema map[string]interface{} `json:"schema"`
	Rules  map[string]interface{} `json:"rules"`
	Fields []string               `json:"fields"`
}

// FormError represents form-related errors
type FormError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// ConvertFormToResponse converts a domain form to a response DTO
func ConvertFormToResponse(form *model.Form) FormResponse {
	return FormResponse{
		ID:          form.ID,
		Title:       form.Title,
		Description: form.Description,
		Schema:      form.Schema,
		UserID:      form.UserID,
		Status:      form.Status,
		CreatedAt:   form.CreatedAt,
		UpdatedAt:   form.UpdatedAt,
	}
}

// ConvertFormListToResponse converts a list of domain forms to response DTOs
func ConvertFormListToResponse(forms []*model.Form) []FormResponse {
	responses := make([]FormResponse, len(forms))
	for i, form := range forms {
		responses[i] = ConvertFormToResponse(form)
	}

	return responses
}
