package validation

import (
	"sync"

	validator "github.com/go-playground/validator/v10"

	"github.com/jonesrussell/goforms/internal/domain/common/interfaces"
)

// validatorImpl implements the Validator interface
type validatorImpl struct {
	*validator.Validate
}

//nolint:gochecknoglobals // singleton pattern requires global instance and once
var (
	instance *validatorImpl
	once     sync.Once
)

// New returns a singleton instance of the validator
func New() interfaces.Validator {
	once.Do(func() {
		instance = &validatorImpl{
			Validate: validator.New(),
		}
	})
	return instance
}

// Struct implements validator.Struct
func (v *validatorImpl) Struct(s any) error {
	return v.Validate.Struct(s)
}

// Var implements validator.Var
func (v *validatorImpl) Var(field any, tag string) error {
	return v.Validate.Var(field, tag)
}

// RegisterValidation implements validator.RegisterValidation
func (v *validatorImpl) RegisterValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	return v.Validate.RegisterValidation(tag, fn)
}
