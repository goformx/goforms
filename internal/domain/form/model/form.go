package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"database/sql/driver"

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
	ID          string         `json:"id" gorm:"column:uuid;primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string         `json:"user_id" gorm:"not null;index;type:uuid"`
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

// Scan implements the sql.Scanner interface for JSON
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = JSON{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	result := make(map[string]any)
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	*j = JSON(result)
	return nil
}

// Value implements the driver.Valuer interface for JSON
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// NewForm creates a new form instance
func NewForm(userID string, title, description string, schema JSON) *Form {
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

// validateProperty validates a single form property
func validateProperty(name string, prop any) error {
	property, isMap := prop.(map[string]any)
	if !isMap {
		return fmt.Errorf("invalid property format for '%s': must be an object", name)
	}

	// Check for required property fields
	if _, exists := property["type"]; !exists {
		return fmt.Errorf("missing type for property '%s'", name)
	}

	// Validate property type
	propType, isString := property["type"].(string)
	if !isString {
		return fmt.Errorf("invalid type format for property '%s'", name)
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

	return nil
}

// validateSchema validates the form schema
func (f *Form) validateSchema() error {
	// Check for required schema fields
	requiredFields := []string{"type"}
	for _, field := range requiredFields {
		if _, exists := f.Schema[field]; !exists {
			return fmt.Errorf("missing required schema field: %s", field)
		}
	}

	// Validate schema type
	schemaType, typeOk := f.Schema["type"].(string)
	if !typeOk || schemaType != "object" {
		return errors.New("invalid schema type: must be 'object'")
	}

	// Check for either properties or components
	hasProperties := false
	hasComponents := false

	if properties, propsOk := f.Schema["properties"].(map[string]any); propsOk {
		hasProperties = true
		// Validate each property
		for name, prop := range properties {
			if err := validateProperty(name, prop); err != nil {
				return err
			}
		}
	}

	if components, compsOk := f.Schema["components"].([]any); compsOk {
		hasComponents = true
		// Components array is valid even if empty
		_ = components
	}

	if !hasProperties && !hasComponents {
		return errors.New("schema must contain either properties or components")
	}

	return nil
}

// Validate validates the form
func (f *Form) Validate() error {
	// Validate title
	if f.Title == "" {
		return ErrFormTitleRequired
	}

	if len(f.Title) < MinTitleLength || len(f.Title) > MaxTitleLength {
		return fmt.Errorf("title must be between %d and %d characters", MinTitleLength, MaxTitleLength)
	}

	// Validate description length
	if len(f.Description) > MaxDescriptionLength {
		return fmt.Errorf("description must not exceed %d characters", MaxDescriptionLength)
	}

	return f.validateSchema()
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
