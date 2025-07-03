package validation

import (
	"fmt"
	"regexp"
	"strconv"
)

// FieldValidator handles field-specific validation logic
type FieldValidator struct {
	emailRegex *regexp.Regexp
	urlRegex   *regexp.Regexp
	phoneRegex *regexp.Regexp
	dateRegex  *regexp.Regexp
}

// NewFieldValidator creates a new field validator
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{
		emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		urlRegex:   regexp.MustCompile(`^https?://[^\s/$.?#].\S*$`),
		phoneRegex: regexp.MustCompile(`^\+?[1-9]\d{0,15}$`),
		dateRegex:  regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
	}
}

// ValidateField validates a single field against its rules
func (v *FieldValidator) ValidateField(fieldName string, value any, rules *FieldValidation) []Error {
	var errors []Error

	// Required field validation
	if requiredErrors := v.validateRequired(fieldName, value, rules); len(requiredErrors) > 0 {
		errors = append(errors, requiredErrors...)
	}

	// Skip further validation if field is empty and not required
	if value == nil || value == "" {
		return errors
	}

	// Type validation
	if typeErrors := v.ValidateFieldType(fieldName, value, rules.Type); typeErrors != nil {
		errors = append(errors, *typeErrors)
	}

	// String-specific validations
	if strErrors := v.validateStringField(fieldName, value, rules); len(strErrors) > 0 {
		errors = append(errors, strErrors...)
	}

	// Numeric validations
	if numErrors := v.validateNumericField(fieldName, value, rules); len(numErrors) > 0 {
		errors = append(errors, numErrors...)
	}

	// Pattern validation
	if patternErrors := v.validatePattern(fieldName, value, rules.Pattern); len(patternErrors) > 0 {
		errors = append(errors, patternErrors...)
	}

	// Options validation
	if optionsErrors := v.validateOptions(fieldName, value, rules.Options); len(optionsErrors) > 0 {
		errors = append(errors, optionsErrors...)
	}

	// Custom rules validation
	if customErrors := v.validateCustomRules(fieldName, value, rules.CustomRules); len(customErrors) > 0 {
		errors = append(errors, customErrors...)
	}

	return errors
}

// validateRequired validates if a required field has a value
func (v *FieldValidator) validateRequired(fieldName string, value any, rules *FieldValidation) []Error {
	if rules.Required && (value == nil || value == "") {
		return []Error{{
			Field:   fieldName,
			Message: rules.getMessage("required", "This field is required"),
			Rule:    "required",
		}}
	}

	return nil
}

// validateStringField validates string-specific rules
func (v *FieldValidator) validateStringField(fieldName string, value any, rules *FieldValidation) []Error {
	var errors []Error

	if strValue, ok := value.(string); ok {
		if rules.MinLength > 0 && len(strValue) < rules.MinLength {
			errors = append(errors, Error{
				Field:   fieldName,
				Message: rules.getMessage("minLength", fmt.Sprintf("Minimum length is %d characters", rules.MinLength)),
				Rule:    "minLength",
			})
		}

		if rules.MaxLength > 0 && len(strValue) > rules.MaxLength {
			errors = append(errors, Error{
				Field:   fieldName,
				Message: rules.getMessage("maxLength", fmt.Sprintf("Maximum length is %d characters", rules.MaxLength)),
				Rule:    "maxLength",
			})
		}
	}

	return errors
}

// validateNumericField validates numeric-specific rules
func (v *FieldValidator) validateNumericField(fieldName string, value any, rules *FieldValidation) []Error {
	var errors []Error

	if numValue, ok := v.toFloat64(value); ok {
		if rules.Min != 0 && numValue < rules.Min {
			errors = append(errors, Error{
				Field:   fieldName,
				Message: rules.getMessage("min", fmt.Sprintf("Minimum value is %g", rules.Min)),
				Rule:    "min",
			})
		}

		if rules.Max != 0 && numValue > rules.Max {
			errors = append(errors, Error{
				Field:   fieldName,
				Message: rules.getMessage("max", fmt.Sprintf("Maximum value is %g", rules.Max)),
				Rule:    "max",
			})
		}
	}

	return errors
}

// validatePattern validates that a value matches a regex pattern
func (v *FieldValidator) validatePattern(fieldName string, value any, pattern string) []Error {
	if pattern == "" {
		return nil
	}

	if strValue, ok := value.(string); ok {
		matched, err := regexp.MatchString(pattern, strValue)
		if err != nil {
			return []Error{{
				Field:   fieldName,
				Message: fmt.Sprintf("Invalid regex pattern: %v", err),
				Rule:    "pattern",
			}}
		}

		if !matched {
			return []Error{{
				Field:   fieldName,
				Message: "Value does not match required pattern",
				Rule:    "pattern",
			}}
		}
	}

	return nil
}

// validateOptions validates that a value is in the allowed options
func (v *FieldValidator) validateOptions(fieldName string, value any, options []string) []Error {
	if len(options) == 0 {
		return nil
	}

	if strValue, ok := value.(string); ok {
		for _, option := range options {
			if strValue == option {
				return nil
			}
		}

		return []Error{{
			Field:   fieldName,
			Message: "Invalid option selected",
			Rule:    "options",
		}}
	}

	return nil
}

// validateCustomRules validates custom validation rules
func (v *FieldValidator) validateCustomRules(fieldName string, value any, rules []Rule) []Error {
	var errors []Error

	for _, rule := range rules {
		if ruleError := v.validateCustomRule(fieldName, value, rule); ruleError != nil {
			errors = append(errors, *ruleError)
		}
	}

	return errors
}

// ValidateFieldType validates the type of a field
func (v *FieldValidator) ValidateFieldType(fieldName string, value any, fieldType string) *Error {
	switch fieldType {
	case "email":
		return v.validateEmail(fieldName, value)
	case "url":
		return v.validateURL(fieldName, value)
	case "phoneNumber":
		return v.validatePhoneNumber(fieldName, value)
	case "date":
		return v.validateDate(fieldName, value)
	case "number":
		return v.validateNumber(fieldName, value)
	case "integer":
		return v.validateInteger(fieldName, value)
	}

	return nil
}

// validateEmail validates email format
func (v *FieldValidator) validateEmail(fieldName string, value any) *Error {
	if strValue, ok := value.(string); ok {
		if !v.emailRegex.MatchString(strValue) {
			return &Error{
				Field:   fieldName,
				Message: "Invalid email format",
				Rule:    "email",
			}
		}
	}

	return nil
}

// validateURL validates URL format
func (v *FieldValidator) validateURL(fieldName string, value any) *Error {
	if strValue, ok := value.(string); ok {
		if !v.urlRegex.MatchString(strValue) {
			return &Error{
				Field:   fieldName,
				Message: "Invalid URL format",
				Rule:    "url",
			}
		}
	}

	return nil
}

// validatePhoneNumber validates phone number format
func (v *FieldValidator) validatePhoneNumber(fieldName string, value any) *Error {
	if strValue, ok := value.(string); ok {
		if !v.phoneRegex.MatchString(strValue) {
			return &Error{
				Field:   fieldName,
				Message: "Invalid phone number format",
				Rule:    "phoneNumber",
			}
		}
	}

	return nil
}

// validateDate validates date format
func (v *FieldValidator) validateDate(fieldName string, value any) *Error {
	if strValue, ok := value.(string); ok {
		if !v.dateRegex.MatchString(strValue) {
			return &Error{
				Field:   fieldName,
				Message: "Invalid date format (YYYY-MM-DD)",
				Rule:    "date",
			}
		}
	}

	return nil
}

// validateNumber validates number format
func (v *FieldValidator) validateNumber(fieldName string, value any) *Error {
	if _, ok := v.toFloat64(value); !ok {
		return &Error{
			Field:   fieldName,
			Message: "Value must be a number",
			Rule:    "number",
		}
	}

	return nil
}

// validateInteger validates integer format
func (v *FieldValidator) validateInteger(fieldName string, value any) *Error {
	if floatValue, ok := v.toFloat64(value); ok {
		if floatValue != float64(int(floatValue)) {
			return &Error{
				Field:   fieldName,
				Message: "Value must be an integer",
				Rule:    "integer",
			}
		}
	} else {
		return &Error{
			Field:   fieldName,
			Message: "Value must be an integer",
			Rule:    "integer",
		}
	}

	return nil
}

// validateCustomRule validates a custom validation rule
func (v *FieldValidator) validateCustomRule(fieldName string, value any, rule Rule) *Error {
	switch rule.Type {
	case "regex":
		return v.validateRegexRule(fieldName, value, rule)
	case "custom":
		// Custom validation logic can be extended here
		return nil
	}

	return nil
}

// validateRegexRule validates a regex rule
func (v *FieldValidator) validateRegexRule(fieldName string, value any, rule Rule) *Error {
	strValue, ok := value.(string)
	if !ok {
		return nil
	}

	pattern, patternOk := rule.Value.(string)
	if !patternOk {
		return nil
	}

	matched, err := regexp.MatchString(pattern, strValue)
	if err != nil {
		return &Error{
			Field:   fieldName,
			Message: fmt.Sprintf("Invalid regex pattern: %v", err),
			Rule:    rule.Type,
		}
	}

	if !matched {
		return &Error{
			Field:   fieldName,
			Message: rule.Message,
			Rule:    rule.Type,
		}
	}

	return nil
}

// toFloat64 converts a value to float64 for numeric validation
func (v *FieldValidator) toFloat64(value any) (float64, bool) {
	switch val := value.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}

	return 0, false
}
