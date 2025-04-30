package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

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
	forms     map[uint]*form.Form      // key is formID
	responses map[string]form.Response // key is responseID
	nextID    uint                     // for generating IDs
}

// NewClient creates a new form client
func NewClient() *Client {
	return &Client{
		forms:     make(map[uint]*form.Form),
		responses: make(map[string]form.Response),
		nextID:    1,
	}
}

// generateID generates a new unique ID
func (c *Client) generateID() uint {
	id := c.nextID
	c.nextID++
	return id
}

// SubmitForm submits a new form
func (c *Client) SubmitForm(ctx context.Context, f form.Form) error {
	if f.Title == "" || f.Schema == nil {
		return ErrInvalidInput
	}
	formID := c.generateID()
	f.ID = formID
	c.forms[formID] = &f
	return nil
}

// GetForm retrieves a form by ID
func (c *Client) GetForm(ctx context.Context, formID string) (*form.Form, error) {
	if formID == "" {
		return nil, ErrInvalidInput
	}
	id, err := strconv.ParseUint(formID, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	f, exists := c.forms[uint(id)]
	if !exists {
		return nil, ErrFormNotFound
	}
	return f, nil
}

// ListForms lists all forms
func (c *Client) ListForms(ctx context.Context) ([]form.Form, error) {
	forms := make([]form.Form, 0, len(c.forms))
	for _, f := range c.forms {
		forms = append(forms, *f)
	}
	return forms, nil
}

// DeleteForm deletes a form by ID
func (c *Client) DeleteForm(ctx context.Context, formID string) error {
	if formID == "" {
		return ErrInvalidInput
	}
	id, err := strconv.ParseUint(formID, 10, 64)
	if err != nil {
		return ErrInvalidInput
	}
	if _, exists := c.forms[uint(id)]; !exists {
		return ErrFormNotFound
	}
	delete(c.forms, uint(id))
	return nil
}

// UpdateForm updates an existing form
func (c *Client) UpdateForm(ctx context.Context, formID string, f form.Form) error {
	if formID == "" {
		return ErrInvalidInput
	}
	id, err := strconv.ParseUint(formID, 10, 64)
	if err != nil {
		return ErrInvalidInput
	}
	if _, exists := c.forms[uint(id)]; !exists {
		return ErrFormNotFound
	}
	f.ID = uint(id)
	c.forms[uint(id)] = &f
	return nil
}

// SubmitResponse submits a form response
func (c *Client) SubmitResponse(ctx context.Context, formID string, response form.Response) error {
	if formID == "" || response.Values == nil {
		return ErrInvalidInput
	}
	responseID := fmt.Sprintf("resp-%d", c.nextID)
	response.ID = responseID
	response.FormID = formID
	response.SubmittedAt = time.Now()
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
