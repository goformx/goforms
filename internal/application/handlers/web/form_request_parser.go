// internal/application/handlers/web/form_request_parser.go
package web

import (
	"strings"

	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/labstack/echo/v4"
)

// FormRequestParser handles parsing and sanitizing form requests
type FormRequestParser struct {
	sanitizer sanitization.ServiceInterface
}

// NewFormRequestParser creates a new FormRequestParser
func NewFormRequestParser(sanitizer sanitization.ServiceInterface) *FormRequestParser {
	return &FormRequestParser{
		sanitizer: sanitizer,
	}
}

// CreateFormRequest represents a form creation request
type CreateFormRequest struct {
	Title       string
	Description string
	CorsOrigins model.JSON
}

// UpdateFormRequest represents a form update request
type UpdateFormRequest struct {
	Title       string
	Description string
	Status      string
	CorsOrigins model.JSON
}

// ParseCreateFormRequest parses and sanitizes a form creation request
func (p *FormRequestParser) ParseCreateFormRequest(c echo.Context) *CreateFormRequest {
	req := &CreateFormRequest{
		Title:       p.sanitizer.String(c.FormValue("title")),
		Description: p.sanitizer.String(c.FormValue("description")),
	}

	if corsOrigins := p.sanitizer.String(c.FormValue("cors_origins")); corsOrigins != "" {
		origins := p.parseCSV(corsOrigins)
		req.CorsOrigins = model.JSON{"origins": origins}
	}

	return req
}

// ParseUpdateFormRequest parses and sanitizes a form update request
func (p *FormRequestParser) ParseUpdateFormRequest(c echo.Context) *UpdateFormRequest {
	req := &UpdateFormRequest{
		Title:       p.sanitizer.String(c.FormValue("title")),
		Description: p.sanitizer.String(c.FormValue("description")),
		Status:      p.sanitizer.String(c.FormValue("status")),
	}

	if corsOrigins := p.sanitizer.String(c.FormValue("cors_origins")); corsOrigins != "" {
		origins := p.parseCSV(corsOrigins)
		req.CorsOrigins = model.JSON{"origins": origins}
	}

	return req
}

// CreateFormFromRequest creates a form model from a create request
func (p *FormRequestParser) CreateFormFromRequest(userID string, req *CreateFormRequest) *model.Form {
	// Create a valid initial schema
	schema := model.JSON{
		"type":       "object",
		"components": []any{},
	}

	form := model.NewForm(userID, req.Title, req.Description, schema)

	// Only set CORS origins if provided
	if req.CorsOrigins != nil {
		form.CorsOrigins = req.CorsOrigins
	}

	return form
}

// ApplyUpdateToForm applies update request to an existing form
func (p *FormRequestParser) ApplyUpdateToForm(form *model.Form, req *UpdateFormRequest) {
	form.Title = req.Title
	form.Description = req.Description
	form.Status = req.Status

	// Only set CORS origins if provided
	if req.CorsOrigins != nil {
		form.CorsOrigins = req.CorsOrigins
	}
}

// parseCSV parses a comma-separated string into a slice of strings,
// trimming whitespace and skipping empty values
func (p *FormRequestParser) parseCSV(input string) []string {
	if input == "" {
		return []string{} // Return empty slice instead of nil
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
