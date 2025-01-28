package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=0,lte=130"`
}

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	assert.NotNil(t, v)
	assert.IsType(t, &CustomValidator{}, v)
}

func TestValidate(t *testing.T) {
	v := NewValidator()

	t.Run("valid struct", func(t *testing.T) {
		test := TestStruct{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   25,
		}
		err := v.Validate(test)
		assert.NoError(t, err)
	})

	t.Run("invalid email", func(t *testing.T) {
		test := TestStruct{
			Name:  "Test User",
			Email: "invalid-email",
			Age:   25,
		}
		err := v.Validate(test)
		assert.Error(t, err)
		validationErr, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Contains(t, validationErr[0].Tag(), "email")
	})

	t.Run("missing required field", func(t *testing.T) {
		test := TestStruct{
			Email: "test@example.com",
			Age:   25,
		}
		err := v.Validate(test)
		assert.Error(t, err)
		validationErr, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Contains(t, validationErr[0].Tag(), "required")
	})

	t.Run("age out of range", func(t *testing.T) {
		test := TestStruct{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   150,
		}
		err := v.Validate(test)
		assert.Error(t, err)
		validationErr, ok := err.(validator.ValidationErrors)
		assert.True(t, ok)
		assert.Contains(t, validationErr[0].Tag(), "lte")
	})

	t.Run("non-struct value", func(t *testing.T) {
		err := v.Validate("not a struct")
		assert.Error(t, err)
	})
}
