package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jonesrussell/goforms/internal/domain/common/errors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validator provides validation functionality
type Validator struct {
	rules map[string]ValidationRule
}

// ValidationRule defines a validation rule function
type ValidationRule func(value any, params ...string) error

// New creates a new validator
func New() *Validator {
	v := &Validator{
		rules: make(map[string]ValidationRule),
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
func (v *Validator) RegisterRule(name string, rule ValidationRule) {
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
	v.registerStringRules()
	v.registerNumericRules()
	v.registerTimeRules()
	v.registerArrayRules()
	v.registerMapRules()
}

// registerStringRules registers string validation rules
func (v *Validator) registerStringRules() {
	v.RegisterRule("required", v.validateRequired)
	v.RegisterRule("email", v.validateEmail)
	v.RegisterRule("min", v.validateStringMin)
	v.RegisterRule("max", v.validateStringMax)
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
	if !emailRegex.MatchString(str) {
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

// registerNumericRules registers numeric validation rules
func (v *Validator) registerNumericRules() {
	v.RegisterRule("min", v.validateNumericMin)
	v.RegisterRule("max", v.validateNumericMax)
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

// registerArrayRules registers array validation rules
func (v *Validator) registerArrayRules() {
	v.RegisterRule("min", v.validateArrayMin)
	v.RegisterRule("max", v.validateArrayMax)
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

// registerMapRules registers map validation rules
func (v *Validator) registerMapRules() {
	v.RegisterRule("min", v.validateMapMin)
	v.RegisterRule("max", v.validateMapMax)
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
