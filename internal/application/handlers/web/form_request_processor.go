package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/validation"
	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
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
	logger    logging.Logger
}

// NewFormRequestProcessor creates a new form request processor
func NewFormRequestProcessor(
	sanitizer sanitization.ServiceInterface,
	validator *validation.FormValidator,
	logger logging.Logger,
) FormRequestProcessor {
	return &FormRequestProcessorImpl{
		sanitizer: sanitizer,
		validator: validator,
		logger:    logger.WithComponent("form_request_processor"),
	}
}

// ProcessCreateRequest processes form creation requests
func (p *FormRequestProcessorImpl) ProcessCreateRequest(c echo.Context) (*FormCreateRequest, error) {
	req := &FormCreateRequest{}

	// Try to bind JSON first (Inertia sends JSON)
	if err := c.Bind(req); err != nil {
		// Fallback to form values
		req.Title = p.sanitizer.String(c.FormValue("title"))
	} else {
		// Sanitize bound values
		req.Title = p.sanitizer.String(req.Title)
	}

	if err := p.validateCreateRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// ProcessUpdateRequest processes form update requests
func (p *FormRequestProcessorImpl) ProcessUpdateRequest(c echo.Context) (*FormUpdateRequest, error) {
	req := &FormUpdateRequest{}

	// Try to bind JSON first (Inertia sends JSON)
	if err := c.Bind(req); err != nil {
		// Fallback to form values
		req.Title = p.sanitizer.String(c.FormValue("title"))
		req.Description = p.sanitizer.String(c.FormValue("description"))
		req.Status = p.sanitizer.String(c.FormValue("status"))
		req.CorsOrigins = p.sanitizer.String(c.FormValue("cors_origins"))
	} else {
		// Sanitize bound values
		req.Title = p.sanitizer.String(req.Title)
		req.Description = p.sanitizer.String(req.Description)
		req.Status = p.sanitizer.String(req.Status)
		req.CorsOrigins = p.sanitizer.String(req.CorsOrigins)
	}

	// Validate CORS origins when publishing
	if req.Status == "published" && strings.TrimSpace(req.CorsOrigins) == "" {
		return nil, errors.New("CORS origins are required when publishing a form")
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
	logger := p.logger.WithOperation("process_submission")

	// Log request details for debugging
	logger.Debug("processing submission request",
		"content_type", c.Request().Header.Get("Content-Type"),
		"content_length", c.Request().ContentLength,
		"method", c.Request().Method)

	var submissionData model.JSON
	if err := json.NewDecoder(c.Request().Body).Decode(&submissionData); err != nil {
		logger.Debug("failed to decode submission data", "error", err)

		return nil, fmt.Errorf("failed to decode submission data: %w", err)
	}

	logger.Debug("submission data decoded", "data_keys", len(submissionData))

	if submissionData == nil {
		logger.Debug("submission data is nil")

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

		// Require CORS origins when publishing
		if req.Status == "published" && req.CorsOrigins == "" {
			return errors.New("CORS origins are required when publishing a form")
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
