package model

import (
	"time"

	"github.com/google/uuid"
)

// Form represents a form in the system
type Form struct {
	ID          string    `json:"id"`
	UserID      uint      `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Schema      JSON      `json:"schema"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JSON is a type alias for map[string]interface{} to represent JSON data
type JSON map[string]interface{}

// NewForm creates a new form instance
func NewForm(userID uint, title, description string, schema JSON) *Form {
	now := time.Now()
	return &Form{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Schema:      schema,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Validate validates the form
func (f *Form) Validate() error {
	if f.Title == "" {
		return ErrFormInvalid
	}
	return nil
}

// Update updates the form with new values
func (f *Form) Update(title, description string, schema JSON) {
	f.Title = title
	f.Description = description
	if schema != nil {
		f.Schema = schema
	}
	f.UpdatedAt = time.Now()
}

// Deactivate marks the form as inactive
func (f *Form) Deactivate() {
	f.Active = false
	f.UpdatedAt = time.Now()
}

// Activate marks the form as active
func (f *Form) Activate() {
	f.Active = true
	f.UpdatedAt = time.Now()
}
