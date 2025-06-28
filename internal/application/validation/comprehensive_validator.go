// Package validation provides comprehensive form validation utilities
// for validating form schemas, submissions, and generating client-side rules.
package validation

import (
	"errors"

	"github.com/goformx/goforms/internal/domain/form/model"
)

// ComprehensiveValidator provides comprehensive form validation
type ComprehensiveValidator struct {
	fieldValidator *FieldValidator
	schemaParser   *SchemaParser
}

// NewComprehensiveValidator creates a new comprehensive form validator
func NewComprehensiveValidator() *ComprehensiveValidator {
	return &ComprehensiveValidator{
		fieldValidator: NewFieldValidator(),
		schemaParser:   NewSchemaParser(),
	}
}

// ValidateForm validates a form submission against its schema
func (v *ComprehensiveValidator) ValidateForm(schema, submission model.JSON) Result {
	result := Result{
		IsValid: true,
		Errors:  []Error{},
	}

	// Extract components from schema
	components, ok := v.schemaParser.ExtractComponents(schema)
	if !ok {
		result.IsValid = false
		result.Errors = append(result.Errors, Error{
			Field:   "schema",
			Message: "Invalid form schema: missing components",
			Rule:    "",
		})

		return result
	}

	// Validate each component
	for _, component := range components {
		if componentMap, componentOk := component.(map[string]any); componentOk {
			fieldErrors := v.validateComponent(componentMap, submission)
			result.Errors = append(result.Errors, fieldErrors...)
		}
	}

	// Check if any errors occurred
	if len(result.Errors) > 0 {
		result.IsValid = false
	}

	return result
}

// validateComponent validates a single form component
func (v *ComprehensiveValidator) validateComponent(component map[string]any, submission model.JSON) []Error {
	// Extract component key
	key, ok := v.schemaParser.ExtractComponentKey(component)
	if !ok {
		return []Error{}
	}

	// Get field value from submission
	fieldValue, exists := submission[key]
	if !exists {
		fieldValue = nil
	}

	// Extract validation rules
	validation := v.schemaParser.ExtractValidationRules(component)

	// Validate field using field validator
	return v.fieldValidator.ValidateField(key, fieldValue, &validation)
}

// GenerateClientValidation generates client-side validation rules from schema
func (v *ComprehensiveValidator) GenerateClientValidation(schema model.JSON) (map[string]any, error) {
	clientRules := make(map[string]any)

	components, ok := v.schemaParser.ExtractComponents(schema)
	if !ok {
		return nil, errors.New("invalid schema: missing components")
	}

	for _, component := range components {
		if componentMap, componentOk := component.(map[string]any); componentOk {
			key, keyOk := v.schemaParser.ExtractComponentKey(componentMap)
			if !keyOk {
				continue
			}

			validation := v.schemaParser.ExtractValidationRules(componentMap)
			clientRules[key] = v.schemaParser.ConvertToClientRules(&validation)
		}
	}

	return clientRules, nil
}
