package model

import (
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/domain/common/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	MinTitleLength       = 3
	MaxTitleLength       = 100
	MaxDescriptionLength = 500
)

// Form represents a form in the system
type Form struct {
	ID          string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Title       string         `json:"title" gorm:"not null;size:100"`
	Description string         `json:"description" gorm:"size:500"`
	Schema      JSON           `json:"schema" gorm:"type:jsonb;not null"`
	Active      bool           `json:"active" gorm:"not null;default:true"`
	CreatedAt   time.Time      `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"not null;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for the Form model
func (f *Form) TableName() string {
	return "forms"
}

// BeforeCreate is a GORM hook that runs before creating a form
func (f *Form) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	if !f.Active {
		f.Active = true
	}
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a form
func (f *Form) BeforeUpdate(tx *gorm.DB) error {
	f.UpdatedAt = time.Now()
	return nil
}

// JSON is a type alias for map[string]any to represent JSON data
type JSON map[string]any

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
		return ErrFormTitleRequired
	}

	if len(f.Title) < MinTitleLength {
		return errors.New(errors.ErrCodeValidation, "form title must be at least 3 characters long", nil)
	}

	if len(f.Title) > MaxTitleLength {
		return errors.New(errors.ErrCodeValidation, "form title must not exceed 100 characters", nil)
	}

	if f.Description != "" && len(f.Description) > MaxDescriptionLength {
		return errors.New(errors.ErrCodeValidation, "form description must not exceed 500 characters", nil)
	}

	if f.Schema == nil {
		return ErrFormSchemaRequired
	}

	if len(f.Schema) == 0 {
		return errors.New(errors.ErrCodeValidation, "form schema cannot be empty", nil)
	}

	// Validate schema structure
	if err := f.validateSchema(); err != nil {
		return errors.Wrap(err, errors.ErrCodeValidation, "invalid form schema")
	}

	return nil
}

// validateSchema validates the form schema structure
func (f *Form) validateSchema() error {
	// Check for required schema fields
	requiredFields := []string{"type", "properties"}
	for _, field := range requiredFields {
		if _, exists := f.Schema[field]; !exists {
			return fmt.Errorf("missing required schema field: %s", field)
		}
	}

	// Validate schema type
	schemaType, ok := f.Schema["type"].(string)
	if !ok || schemaType != "object" {
		return errors.New(errors.ErrCodeValidation, "invalid schema type: must be 'object'", nil)
	}

	// Validate properties
	properties, ok := f.Schema["properties"].(map[string]any)
	if !ok {
		return errors.New(errors.ErrCodeValidation, "invalid properties format: must be an object", nil)
	}

	if len(properties) == 0 {
		return errors.New(errors.ErrCodeValidation, "schema must contain at least one property", nil)
	}

	// Validate each property
	for name, prop := range properties {
		property, isMap := prop.(map[string]any)
		if !isMap {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("invalid property format for '%s': must be an object", name), nil)
		}

		// Check for required property fields
		if _, exists := property["type"]; !exists {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("missing type for property '%s'", name), nil)
		}

		// Validate property type
		propType, isString := property["type"].(string)
		if !isString {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("invalid type format for property '%s'", name), nil)
		}

		// Validate property type value
		validTypes := map[string]bool{
			"string":  true,
			"number":  true,
			"integer": true,
			"boolean": true,
			"array":   true,
			"object":  true,
		}

		if !validTypes[propType] {
			return fmt.Errorf("invalid type '%s' for property '%s'", propType, name)
		}
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
