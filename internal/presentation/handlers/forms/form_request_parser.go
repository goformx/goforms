package forms

import (
	"fmt"
	"strings"

	"github.com/goformx/goforms/internal/domain/form/model"
	"github.com/labstack/echo/v4"
)

// FormRequestParser parses form-related requests.
type FormRequestParser struct{}

// NewFormRequestParser creates a new form request parser
func NewFormRequestParser() *FormRequestParser {
	return &FormRequestParser{}
}

// ParseCreateForm parses form creation data from the request (JSON or form)
func (p *FormRequestParser) ParseCreateForm(c echo.Context) (*model.Form, error) {
	contentType := c.Request().Header.Get("Content-Type")

	var form *model.Form

	if strings.Contains(contentType, "application/json") {
		var data struct {
			Title       string     `json:"title"`
			Description string     `json:"description"`
			Schema      model.JSON `json:"schema"`
		}
		if err := c.Bind(&data); err != nil {
			return nil, fmt.Errorf("failed to bind create form: %w", err)
		}

		form = &model.Form{
			Title:       data.Title,
			Description: data.Description,
			Schema:      data.Schema,
		}
	} else {
		form = &model.Form{
			Title:       c.FormValue("title"),
			Description: c.FormValue("description"),
			Schema:      model.JSON{}, // Default empty schema
		}
	}

	// Sanitize inputs
	form.Title = strings.TrimSpace(form.Title)
	form.Description = strings.TrimSpace(form.Description)

	return form, nil
}

// ParseUpdateForm parses form update data from the request (JSON or form)
func (p *FormRequestParser) ParseUpdateForm(c echo.Context) (*model.Form, error) {
	contentType := c.Request().Header.Get("Content-Type")

	var form *model.Form

	if strings.Contains(contentType, "application/json") {
		var data struct {
			Title       string     `json:"title"`
			Description string     `json:"description"`
			Schema      model.JSON `json:"schema"`
			Status      string     `json:"status"`
		}
		if err := c.Bind(&data); err != nil {
			return nil, fmt.Errorf("failed to bind update form: %w", err)
		}

		form = &model.Form{
			Title:       data.Title,
			Description: data.Description,
			Schema:      data.Schema,
			Status:      data.Status,
		}
	} else {
		form = &model.Form{
			Title:       c.FormValue("title"),
			Description: c.FormValue("description"),
			Status:      c.FormValue("status"),
			Schema:      model.JSON{}, // Will be updated separately
		}
	}

	// Sanitize inputs
	form.Title = strings.TrimSpace(form.Title)
	form.Description = strings.TrimSpace(form.Description)
	form.Status = strings.TrimSpace(form.Status)

	return form, nil
}

// ParseFormID parses form ID from URL parameters
func (p *FormRequestParser) ParseFormID(c echo.Context) (string, error) {
	formID := c.Param("id")
	if formID == "" {
		return "", fmt.Errorf("form ID is required")
	}

	return formID, nil
}

// ValidateCreateForm validates form creation data
func (p *FormRequestParser) ValidateCreateForm(form *model.Form) error {
	if form.Title == "" {
		return fmt.Errorf("title is required")
	}

	if len(form.Title) < 3 {
		return fmt.Errorf("title must be at least 3 characters")
	}

	if len(form.Title) > 100 {
		return fmt.Errorf("title must not exceed 100 characters")
	}

	if len(form.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}

	return nil
}

// ValidateUpdateForm validates form update data
func (p *FormRequestParser) ValidateUpdateForm(form *model.Form) error {
	if form.Title == "" {
		return fmt.Errorf("title is required")
	}

	if len(form.Title) < 3 {
		return fmt.Errorf("title must be at least 3 characters")
	}

	if len(form.Title) > 100 {
		return fmt.Errorf("title must not exceed 100 characters")
	}

	if len(form.Description) > 500 {
		return fmt.Errorf("description must not exceed 500 characters")
	}

	// Validate status if provided
	if form.Status != "" {
		validStatuses := []string{"draft", "published", "archived"}
		isValid := false

		for _, status := range validStatuses {
			if form.Status == status {
				isValid = true

				break
			}
		}

		if !isValid {
			return fmt.Errorf("invalid status: must be one of %v", validStatuses)
		}
	}

	return nil
}
