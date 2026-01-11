// Package model contains domain models and error definitions for forms.
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
	// MinTitleLength is the minimum length for a form title
	MinTitleLength = 3
	// MaxTitleLength is the maximum length for a form title
	MaxTitleLength = 100
	// MaxDescriptionLength is the maximum length for a form description
	MaxDescriptionLength = 500
	// MaxFields is the maximum number of fields allowed in a form
	MaxFields = 50
)

var (
	// ErrInvalidJSON represents an invalid JSON error
	ErrInvalidJSON = errors.New("invalid JSON")
)

// Field represents a form field
type Field struct {
	ID        string    `gorm:"primaryKey"             json:"id"`
	FormID    string    `gorm:"not null"               json:"form_id"`
	Label     string    `gorm:"size:100;not null"      json:"label"`
	Type      string    `gorm:"size:20;not null"       json:"type"`
	Required  bool      `gorm:"not null;default:false" json:"required"`
	Options   []string  `gorm:"type:json"              json:"options"`
	CreatedAt time.Time `gorm:"not null"               json:"created_at"`
	UpdatedAt time.Time `gorm:"not null"               json:"updated_at"`
}

// Validate validates the field
func (f *Field) Validate() error {
	if f.Label == "" {
		return errors.New("label is required")
	}

	if f.Type == "" {
		return errors.New("type is required")
	}

	return nil
}

// Form represents a form in the system
type Form struct {
	ID          string         `gorm:"column:uuid;primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID      string         `gorm:"not null;index;type:uuid"                                   json:"user_id"`
	Title       string         `gorm:"not null;size:100"                                          json:"title"`
	Description string         `gorm:"size:500"                                                   json:"description"`
	Schema      JSON           `gorm:"type:jsonb;not null"                                        json:"schema"`
	Active      bool           `gorm:"not null;default:true"                                      json:"active"`
	CreatedAt   time.Time      `gorm:"not null;autoCreateTime"                                    json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null;autoUpdateTime"                                    json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                                      json:"-"`
	Fields      []Field        `gorm:"foreignKey:FormID"                                          json:"fields"`
	Status      string         `gorm:"size:20;not null;default:'draft'"                           json:"status"`

	// CORS settings for form embedding
	CorsOrigins JSON `gorm:"type:json" json:"cors_origins"`
	CorsMethods JSON `gorm:"type:json" json:"cors_methods"`
	CorsHeaders JSON `gorm:"type:json" json:"cors_headers"`
}

// GetID returns the form's ID
func (f *Form) GetID() string {
	return f.ID
}

// SetID sets the form's ID
func (f *Form) SetID(id string) {
	f.ID = id
}

// TableName specifies the table name for the Form model
func (f *Form) TableName() string {
	return "forms"
}

// BeforeCreate is a GORM hook that runs before creating a form
func (f *Form) BeforeCreate(_ *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}

	if !f.Active {
		f.Active = true
	}

	if f.Status == "" {
		f.Status = "draft"
	}

	// Ensure CORS fields are properly initialized
	if f.CorsOrigins == nil {
		f.CorsOrigins = JSON{}
	}

	if f.CorsMethods == nil {
		f.CorsMethods = JSON{}
	}

	if f.CorsHeaders == nil {
		f.CorsHeaders = JSON{}
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a form
func (f *Form) BeforeUpdate(_ *gorm.DB) error {
	f.UpdatedAt = time.Now()

	return nil
}

// BeforeSave is a GORM hook that runs before saving a form
func (f *Form) BeforeSave(_ *gorm.DB) error {
	// Ensure CORS fields are properly initialized
	if f.CorsOrigins == nil {
		f.CorsOrigins = JSON{}
	}

	if f.CorsMethods == nil {
		f.CorsMethods = JSON{}
	}

	if f.CorsHeaders == nil {
		f.CorsHeaders = JSON{}
	}

	return nil
}

// JSON is a custom type for handling JSON data
type JSON map[string]any

// Scan implements the sql.Scanner interface for JSON
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = nil

		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}

	// First try to unmarshal as an object
	var result map[string]any

	err := json.Unmarshal(bytes, &result)
	if err == nil {
		*j = JSON(result)

		return nil
	}

	// If that fails, try to unmarshal as an array and convert to object
	var arrayResult []any

	err = json.Unmarshal(bytes, &arrayResult)
	if err != nil {
		return fmt.Errorf("unmarshal JSON scan value: %w", err)
	}

	// Convert array to object with "data" key
	*j = JSON{"data": arrayResult}

	return nil
}

// Value implements the driver.Valuer interface for JSON
func (j *JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, ErrInvalidJSON
	}

	data, err := json.Marshal(*j)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON value: %w", err)
	}

	return data, nil
}

// MarshalJSON implements the json.Marshaler interface
func (j *JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return nil, ErrInvalidJSON
	}

	data, err := json.Marshal(*j)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON to bytes: %w", err)
	}

	return data, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return ErrInvalidJSON
	}

	if err := json.Unmarshal(data, (*map[string]any)(j)); err != nil {
		return fmt.Errorf("unmarshal JSON from bytes: %w", err)
	}

	return nil
}

// NewForm creates a new form instance
func NewForm(userID, title, description string, schema JSON) *Form {
	now := time.Now()

	return &Form{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       title,
		Description: description,
		Schema:      schema,
		Active:      true,
		Status:      "draft",
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   gorm.DeletedAt{},
		Fields:      []Field{},
		CorsOrigins: JSON{},
		CorsMethods: JSON{},
		CorsHeaders: JSON{},
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
	// Validate required schema fields
	if err := f.validateRequiredSchemaFields(); err != nil {
		return err
	}

	// Validate schema type
	if err := f.validateSchemaType(); err != nil {
		return err
	}

	// Validate schema content
	if err := f.validateSchemaContent(); err != nil {
		return err
	}

	return nil
}

// validateRequiredSchemaFields validates that all required schema fields are present
func (f *Form) validateRequiredSchemaFields() error {
	// Accept either JSON Schema format (type) or Form.io format (display)
	_, hasType := f.Schema["type"]
	_, hasDisplay := f.Schema["display"]

	if !hasType && !hasDisplay {
		return errors.New("schema must have 'type' or 'display' field")
	}

	return nil
}

// validateSchemaType validates that the schema type is correct
func (f *Form) validateSchemaType() error {
	// Accept Form.io format (display: form) or JSON Schema format (type: object)
	if display, ok := f.Schema["display"].(string); ok && display == "form" {
		return nil
	}

	schemaType, typeOk := f.Schema["type"].(string)
	if !typeOk || schemaType != "object" {
		return errors.New("invalid schema: must have 'type: object' or 'display: form'")
	}

	return nil
}

// validateSchemaContent validates the content of the schema (properties or components)
func (f *Form) validateSchemaContent() error {
	hasProperties, propErr := f.validateProperties()
	hasComponents := f.validateComponents()

	if propErr != nil {
		return propErr
	}

	if !hasProperties && !hasComponents {
		return errors.New("schema must contain either properties or components")
	}

	return nil
}

// validateProperties validates the properties section of the schema
func (f *Form) validateProperties() (bool, error) {
	properties, propsOk := f.Schema["properties"].(map[string]any)
	if !propsOk {
		return false, nil
	}

	// Validate each property
	for name, prop := range properties {
		if err := validateProperty(name, prop); err != nil {
			return false, err
		}
	}

	return true, nil
}

// validateComponents validates the components section of the schema
func (f *Form) validateComponents() bool {
	_, compsOk := f.Schema["components"].([]any)

	return compsOk
}

// Validate validates the form
func (f *Form) Validate() error {
	if f.Title == "" {
		return errors.New("title is required")
	}

	if len(f.Title) < MinTitleLength {
		return fmt.Errorf("title must be between %d and %d characters", MinTitleLength, MaxTitleLength)
	}

	if len(f.Title) > MaxTitleLength {
		return fmt.Errorf("title must be between %d and %d characters", MinTitleLength, MaxTitleLength)
	}

	if len(f.Description) > MaxDescriptionLength {
		return fmt.Errorf("description must not exceed %d characters", MaxDescriptionLength)
	}

	if len(f.Fields) > MaxFields {
		return fmt.Errorf("form cannot have more than %d fields", MaxFields)
	}

	for i := range f.Fields {
		if err := f.Fields[i].Validate(); err != nil {
			return fmt.Errorf("invalid field: %w", err)
		}
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

// extractStringSlice extracts a string slice from JSON array
func extractStringSlice(data JSON, key string) []string {
	var result []string
	if data == nil {
		return result
	}

	// First try to get the value directly by key
	if arr, ok := data[key].([]any); ok {
		for _, item := range arr {
			if str, strOk := item.(string); strOk {
				result = append(result, str)
			}
		}

		return result
	}

	// If not found by key, check if the data itself is an array (stored under "data" key)
	if arr, ok := data["data"].([]any); ok {
		for _, item := range arr {
			if str, strOk := item.(string); strOk {
				result = append(result, str)
			}
		}

		return result
	}

	return result
}

// GetCorsConfig returns the CORS configuration for this form
func (f *Form) GetCorsConfig() (origins, methods, headers []string) {
	origins = extractStringSlice(f.CorsOrigins, "origins")
	methods = extractStringSlice(f.CorsMethods, "methods")
	headers = extractStringSlice(f.CorsHeaders, "headers")

	return origins, methods, headers
}

// SetCorsConfig sets the CORS configuration for this form
func (f *Form) SetCorsConfig(origins, methods, headers []string) {
	f.CorsOrigins = JSON{"origins": origins}
	f.CorsMethods = JSON{"methods": methods}
	f.CorsHeaders = JSON{"headers": headers}
}
