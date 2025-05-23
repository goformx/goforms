package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"net/mail"

	"github.com/goformx/goforms/internal/domain/common/errors"
)

// Rule represents a validation rule
type Rule func(value any, params ...string) error

// Validator provides validation functionality
type Validator struct {
	rules map[string]Rule
}

// New creates a new validator
func New() *Validator {
	v := &Validator{
		rules: make(map[string]Rule),
	}
	v.registerDefaultRules()
	return v
}

// ValidateStruct validates a struct
func (v *Validator) ValidateStruct(s any) error {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors.ErrValidation.WithContext("type", val.Kind().String())
	}

	typ := val.Type()
	for i := range typ.NumField() {
		field := typ.Field(i)
		value := val.Field(i)

		if tag := field.Tag.Get("validate"); tag != "" {
			if err := v.validateField(value.Interface(), tag); err != nil {
				return errors.Wrap(err, errors.ErrCodeValidation, fmt.Sprintf("field %s validation failed", field.Name))
			}
		}
	}

	return nil
}

// ValidateField validates a field
func (v *Validator) ValidateField(field any, tag string) error {
	return v.validateField(field, tag)
}

// RegisterRule registers a new validation rule
func (v *Validator) RegisterRule(name string, rule Rule) {
	v.rules[name] = rule
}

func (v *Validator) validateField(value any, tag string) error {
	rules := strings.Split(tag, ",")
	for _, rule := range rules {
		parts := strings.Split(rule, "=")
		ruleName := parts[0]
		var ruleValue string
		if len(parts) > 1 {
			ruleValue = parts[1]
		}

		if validator, exists := v.rules[ruleName]; exists {
			if err := validator(value, ruleValue); err != nil {
				return err
			}
		}
	}
	return nil
}

// registerDefaultRules registers the default validation rules
func (v *Validator) registerDefaultRules() {
	v.RegisterRule("min", v.dispatchMin)
	v.RegisterRule("max", v.dispatchMax)
	v.RegisterRule("required", v.validateRequired)
	v.RegisterRule("email", v.validateEmail)
	v.registerTimeRules()
}

// dispatchMin dispatches min validation based on value type
func (v *Validator) dispatchMin(value any, params ...string) error {
	switch val := value.(type) {
	case string:
		return v.validateStringMin(val, params...)
	case int, float64:
		return v.validateNumericMin(val, params...)
	case []any:
		return v.validateArrayMin(val, params...)
	case map[string]any:
		return v.validateMapMin(val, params...)
	default:
		return nil
	}
}

// dispatchMax dispatches max validation based on value type
func (v *Validator) dispatchMax(value any, params ...string) error {
	switch val := value.(type) {
	case string:
		return v.validateStringMax(val, params...)
	case int, float64:
		return v.validateNumericMax(val, params...)
	case []any:
		return v.validateArrayMax(val, params...)
	case map[string]any:
		return v.validateMapMax(val, params...)
	default:
		return nil
	}
}

// validateRequired validates that a field is not empty
func (v *Validator) validateRequired(value any, params ...string) error {
	if value == nil || value == "" {
		return errors.New(errors.ErrCodeValidation, "field is required")
	}
	return nil
}

// validateEmail validates email format
func (v *Validator) validateEmail(value any, params ...string) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New(errors.ErrCodeValidation, "value must be a string")
	}
	if _, err := mail.ParseAddress(str); err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid email format")
	}
	return nil
}

// validateStringMin validates minimum string length
func (v *Validator) validateStringMin(value any, params ...string) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New(errors.ErrCodeValidation, "value must be a string")
	}
	minLength, err := strconv.Atoi(params[0])
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid min parameter")
	}
	if len(str) < minLength {
		return errors.New(errors.ErrCodeValidation, fmt.Sprintf("string length must be at least %d", minLength))
	}
	return nil
}

// validateStringMax validates maximum string length
func (v *Validator) validateStringMax(value any, params ...string) error {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return errors.New(errors.ErrCodeValidation, "value must be a string")
	}
	maxLength, err := strconv.Atoi(params[0])
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid max parameter")
	}
	if len(str) > maxLength {
		return errors.New(errors.ErrCodeValidation, fmt.Sprintf("string length must be at most %d", maxLength))
	}
	return nil
}

// validateNumericMin validates minimum numeric value
func (v *Validator) validateNumericMin(value any, params ...string) error {
	if value == nil {
		return nil
	}
	minValue, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid min parameter")
	}
	switch val := value.(type) {
	case int:
		if float64(val) < minValue {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("value must be at least %f", minValue))
		}
	case float64:
		if val < minValue {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("value must be at least %f", minValue))
		}
	default:
		return errors.New(errors.ErrCodeValidation, "value must be numeric")
	}
	return nil
}

// validateNumericMax validates maximum numeric value
func (v *Validator) validateNumericMax(value any, params ...string) error {
	if value == nil {
		return nil
	}
	maxValue, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid max parameter")
	}
	switch val := value.(type) {
	case int:
		if float64(val) > maxValue {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("value must be at most %f", maxValue))
		}
	case float64:
		if val > maxValue {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("value must be at most %f", maxValue))
		}
	default:
		return errors.New(errors.ErrCodeValidation, "value must be numeric")
	}
	return nil
}

// registerTimeRules registers time validation rules
func (v *Validator) registerTimeRules() {
	v.RegisterRule("after", func(value any, params ...string) error {
		if value == nil {
			return nil
		}
		t, ok := value.(time.Time)
		if !ok {
			return errors.New(errors.ErrCodeValidation, "value must be a time")
		}
		after, err := time.Parse(time.RFC3339, params[0])
		if err != nil {
			return errors.New(errors.ErrCodeValidation, "invalid after parameter")
		}
		if !t.After(after) {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("time must be after %s", after))
		}
		return nil
	})

	v.RegisterRule("before", func(value any, params ...string) error {
		if value == nil {
			return nil
		}
		t, ok := value.(time.Time)
		if !ok {
			return errors.New(errors.ErrCodeValidation, "value must be a time")
		}
		before, err := time.Parse(time.RFC3339, params[0])
		if err != nil {
			return errors.New(errors.ErrCodeValidation, "invalid before parameter")
		}
		if !t.Before(before) {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("time must be before %s", before))
		}
		return nil
	})
}

// validateArrayMin validates minimum array length
func (v *Validator) validateArrayMin(value any, params ...string) error {
	if value == nil {
		return nil
	}
	minLength, err := strconv.Atoi(params[0])
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid min parameter")
	}
	switch val := value.(type) {
	case []any:
		if len(val) < minLength {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("array length must be at least %d", minLength))
		}
	default:
		return errors.New(errors.ErrCodeValidation, "value must be an array")
	}
	return nil
}

// validateArrayMax validates maximum array length
func (v *Validator) validateArrayMax(value any, params ...string) error {
	if value == nil {
		return nil
	}
	maxLength, err := strconv.Atoi(params[0])
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid max parameter")
	}
	switch val := value.(type) {
	case []any:
		if len(val) > maxLength {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("array length must be at most %d", maxLength))
		}
	default:
		return errors.New(errors.ErrCodeValidation, "value must be an array")
	}
	return nil
}

// validateMapMin validates minimum map length
func (v *Validator) validateMapMin(value any, params ...string) error {
	if value == nil {
		return nil
	}
	minLength, err := strconv.Atoi(params[0])
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid min parameter")
	}
	switch val := value.(type) {
	case map[string]any:
		if len(val) < minLength {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("map length must be at least %d", minLength))
		}
	default:
		return errors.New(errors.ErrCodeValidation, "value must be a map")
	}
	return nil
}

// validateMapMax validates maximum map length
func (v *Validator) validateMapMax(value any, params ...string) error {
	if value == nil {
		return nil
	}
	maxLength, err := strconv.Atoi(params[0])
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "invalid max parameter")
	}
	switch val := value.(type) {
	case map[string]any:
		if len(val) > maxLength {
			return errors.New(errors.ErrCodeValidation, fmt.Sprintf("map length must be at most %d", maxLength))
		}
	default:
		return errors.New(errors.ErrCodeValidation, "value must be a map")
	}
	return nil
}
