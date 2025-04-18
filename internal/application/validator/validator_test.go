package validator_test

import (
	"testing"

	appvalidator "github.com/jonesrussell/goforms/internal/application/validator"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	v := appvalidator.NewValidator()
	assert.NotNil(t, v)
}

func TestValidate(t *testing.T) {
	v := appvalidator.NewValidator()

	t.Run("valid struct", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}

		test := &TestStruct{
			Name:  "Test",
			Email: "test@example.com",
		}

		err := v.Validate(test)
		assert.NoError(t, err)
	})

	t.Run("invalid struct", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required"`
			Email string `validate:"required,email"`
		}

		test := &TestStruct{
			Name:  "",
			Email: "invalid-email",
		}

		err := v.Validate(test)
		assert.Error(t, err)
	})
}
