package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jonesrussell/goforms/internal/domain/common/errors"
)

// Validator provides validation functionality
type Validator struct {
	rules map[string]ValidationRule
}

// ValidationRule defines a validation rule function
type ValidationRule func(value any, ruleValue string) error

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
	for i := 0; i < typ.NumField(); i++ {
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

func (v *Validator) registerDefaultRules() {
	// Required rule
	v.RegisterRule("required", func(value any, _ string) error {
		if value == nil {
			return errors.ErrRequiredField
		}

		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.IsZero() {
			return errors.ErrRequiredField
		}

		return nil
	})

	// Email rule
	v.RegisterRule("email", func(value any, _ string) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return errors.ErrInvalidFormat.WithContext("type", reflect.TypeOf(value).String())
		}

		if !strings.Contains(str, "@") || !strings.Contains(str, ".") {
			return errors.ErrInvalidFormat.WithContext("value", str)
		}

		return nil
	})

	// Min length rule
	v.RegisterRule("min", func(value any, ruleValue string) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return errors.ErrInvalidFormat.WithContext("type", reflect.TypeOf(value).String())
		}

		minLen, err := strconv.Atoi(ruleValue)
		if err != nil {
			return errors.ErrInvalidValue.WithContext("min_length", ruleValue)
		}

		if len(str) < minLen {
			return errors.ErrInvalidValue.WithContext("min_length", ruleValue)
		}

		return nil
	})
}
