package form

import (
	"context"
	"time"
)

// Form represents a form in the system
type Form struct {
	Name    string
	Fields  []Field
	Options FormOptions
}

// Field represents a form field
type Field struct {
	Name string
	Type string
}

// FormOptions represents form configuration options
type FormOptions struct {
	// Add form options as needed
}

// Response represents a form submission response
type Response struct {
	FormID      string
	Values      map[string]any
	SubmittedAt time.Time
}

// Client represents a form client interface
type Client interface {
	SubmitForm(ctx context.Context, form Form) error
	GetForm(ctx context.Context, formID string) (*Form, error)
	ListForms(ctx context.Context) ([]Form, error)
	DeleteForm(ctx context.Context, formID string) error
	UpdateForm(ctx context.Context, formID string, form Form) error
	SubmitResponse(ctx context.Context, formID string, response Response) error
	GetResponse(ctx context.Context, responseID string) (*Response, error)
	ListResponses(ctx context.Context, formID string) ([]Response, error)
	DeleteResponse(ctx context.Context, responseID string) error
}
