package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/jonesrussell/goforms/internal/domain/form"
)

var (
	// ErrFormNotFound is returned when a form is not found
	ErrFormNotFound = errors.New("form not found")
	// ErrResponseNotFound is returned when a response is not found
	ErrResponseNotFound = errors.New("response not found")
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// Client implements the form.Client interface
type Client struct {
	forms     map[string]form.Form     // key is formID
	responses map[string]form.Response // key is responseID
	nextID    int                      // for generating IDs
}

// NewClient creates a new form client
func NewClient() *Client {
	return &Client{
		forms:     make(map[string]form.Form),
		responses: make(map[string]form.Response),
		nextID:    1,
	}
}

// generateID generates a new unique ID
func (c *Client) generateID() string {
	id := c.nextID
	c.nextID++
	return fmt.Sprintf("id-%d", id)
}

// SubmitForm submits a new form
func (c *Client) SubmitForm(ctx context.Context, f form.Form) error {
	if f.Name == "" || len(f.Fields) == 0 {
		return ErrInvalidInput
	}
	formID := c.generateID()
	f.ID = formID
	c.forms[formID] = f
	return nil
}

// GetForm retrieves a form by ID
func (c *Client) GetForm(ctx context.Context, formID string) (*form.Form, error) {
	if formID == "" {
		return nil, ErrInvalidInput
	}
	f, exists := c.forms[formID]
	if !exists {
		return nil, ErrFormNotFound
	}
	return &f, nil
}

// ListForms lists all forms
func (c *Client) ListForms(ctx context.Context) ([]form.Form, error) {
	forms := make([]form.Form, 0, len(c.forms))
	for _, f := range c.forms {
		forms = append(forms, f)
	}
	return forms, nil
}

// DeleteForm deletes a form by ID
func (c *Client) DeleteForm(ctx context.Context, formID string) error {
	if formID == "" {
		return ErrInvalidInput
	}
	if _, exists := c.forms[formID]; !exists {
		return ErrFormNotFound
	}
	delete(c.forms, formID)
	return nil
}

// UpdateForm updates an existing form
func (c *Client) UpdateForm(ctx context.Context, formID string, f form.Form) error {
	if formID == "" {
		return ErrInvalidInput
	}
	if _, exists := c.forms[formID]; !exists {
		return ErrFormNotFound
	}
	f.ID = formID
	c.forms[formID] = f
	return nil
}

// SubmitResponse submits a form response
func (c *Client) SubmitResponse(ctx context.Context, formID string, response form.Response) error {
	if formID == "" {
		return ErrInvalidInput
	}
	if _, exists := c.forms[formID]; !exists {
		return ErrFormNotFound
	}
	responseID := c.generateID()
	response.ID = responseID
	response.FormID = formID
	c.responses[responseID] = response
	return nil
}

// GetResponse retrieves a response by ID
func (c *Client) GetResponse(ctx context.Context, responseID string) (*form.Response, error) {
	if responseID == "" {
		return nil, ErrInvalidInput
	}
	r, exists := c.responses[responseID]
	if !exists {
		return nil, ErrResponseNotFound
	}
	return &r, nil
}

// ListResponses lists all responses for a form
func (c *Client) ListResponses(ctx context.Context, formID string) ([]form.Response, error) {
	if formID == "" {
		return nil, ErrInvalidInput
	}
	if _, exists := c.forms[formID]; !exists {
		return nil, ErrFormNotFound
	}
	responses := make([]form.Response, 0)
	for _, r := range c.responses {
		if r.FormID == formID {
			responses = append(responses, r)
		}
	}
	return responses, nil
}

// DeleteResponse deletes a response by ID
func (c *Client) DeleteResponse(ctx context.Context, responseID string) error {
	if responseID == "" {
		return ErrInvalidInput
	}
	if _, exists := c.responses[responseID]; !exists {
		return ErrResponseNotFound
	}
	delete(c.responses, responseID)
	return nil
}
