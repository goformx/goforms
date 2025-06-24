package web

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
	"github.com/labstack/echo/v4"
)

// Validation constants
const (
	MaxTitleLength       = 255
	MaxDescriptionLength = 1000
)

// FormRequestProcessorImpl implements FormRequestProcessor
type FormRequestProcessorImpl struct {
	sanitizer sanitization.ServiceInterface
	validator *validation.FormValidator
}

// NewFormRequestProcessor creates a new form request processor
func NewFormRequestProcessor(
	sanitizer sanitization.ServiceInterface,
	validator *validation.FormValidator,
) FormRequestProcessor {
	return &FormRequestProcessorImpl{
		sanitizer: sanitizer,
		validator: validator,
	}
}

// ProcessCreateRequest processes form creation requests
func (p *FormRequestProcessorImpl) ProcessCreateRequest(c echo.Context) (*FormCreateRequest, error) {
	req := &FormCreateRequest{
		Title:       p.sanitizer.String(c.FormValue("title")),
		Description: p.sanitizer.String(c.FormValue("description")),
		CorsOrigins: p.sanitizer.String(c.FormValue("cors_origins")),
		CorsMethods: p.sanitizer.String(c.FormValue("cors_methods")),
	}

	if err := p.validateCreateRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// ProcessUpdateRequest processes form update requests
func (p *FormRequestProcessorImpl) ProcessUpdateRequest(c echo.Context) (*FormUpdateRequest, error) {
	req := &FormUpdateRequest{
		Title:       p.sanitizer.String(c.FormValue("title")),
		Description: p.sanitizer.String(c.FormValue("description")),
		Status:      p.sanitizer.String(c.FormValue("status")),
		CorsOrigins: p.sanitizer.String(c.FormValue("cors_origins")),
	}

	if err := p.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// ProcessSchemaUpdateRequest processes schema update requests
func (p *FormRequestProcessorImpl) ProcessSchemaUpdateRequest(c echo.Context) (model.JSON, error) {
	var schema model.JSON
	if err := json.NewDecoder(c.Request().Body).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode schema: %w", err)
	}

	if err := p.validateSchema(schema); err != nil {
		return nil, err
	}

	return schema, nil
}

// ProcessSubmissionRequest processes form submission requests
func (p *FormRequestProcessorImpl) ProcessSubmissionRequest(c echo.Context) (model.JSON, error) {
	var submissionData model.JSON
	if err := json.NewDecoder(c.Request().Body).Decode(&submissionData); err != nil {
		return nil, fmt.Errorf("failed to decode submission data: %w", err)
	}

	if submissionData == nil {
		return nil, errors.New("submission data is required")
	}

	return submissionData, nil
}

// validateCreateRequest validates form creation request
func (p *FormRequestProcessorImpl) validateCreateRequest(req *FormCreateRequest) error {
	if req.Title == "" {
		return errors.New("title is required")
	}

	if len(req.Title) > MaxTitleLength {
		return errors.New("title too long")
	}

	if len(req.Description) > MaxDescriptionLength {
		return errors.New("description too long")
	}

	return nil
}

// validateUpdateRequest validates form update request
func (p *FormRequestProcessorImpl) validateUpdateRequest(req *FormUpdateRequest) error {
	if req.Title == "" {
		return errors.New("title is required")
	}

	if len(req.Title) > MaxTitleLength {
		return errors.New("title too long")
	}

	if len(req.Description) > MaxDescriptionLength {
		return errors.New("description too long")
	}

	// Validate status if provided
	if req.Status != "" {
		validStatuses := []string{"draft", "published", "archived"}
		isValid := false
		for _, status := range validStatuses {
			if req.Status == status {
				isValid = true
				break
			}
		}
		if !isValid {
			return errors.New("invalid form status")
		}
	}

	return nil
}

// validateSchema validates form schema
func (p *FormRequestProcessorImpl) validateSchema(schema model.JSON) error {
	if schema == nil {
		return errors.New("schema is required")
	}

	// Schema is already a map[string]any, no need for type assertion
	return nil
}
