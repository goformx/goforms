package validation

import (
	"sync"

	validator "github.com/go-playground/validator/v10"
	"github.com/jonesrussell/goforms/internal/domain/common/interfaces"
)

// validatorImpl implements the Validator interface
type validatorImpl struct {
	validate *validator.Validate
}

var (
	instance *validatorImpl
	once     sync.Once
)

// New returns a singleton instance of the validator
func New() interfaces.Validator {
	once.Do(func() {
		instance = &validatorImpl{
			validate: validator.New(),
		}
	})
	return instance
}

// Struct implements validator.Struct
func (v *validatorImpl) Struct(s interface{}) error {
	return v.validate.Struct(s)
}

// Var implements validator.Var
func (v *validatorImpl) Var(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// RegisterValidation implements validator.RegisterValidation
func (v *validatorImpl) RegisterValidation(tag string, fn func(fl validator.FieldLevel) bool) error {
	return v.validate.RegisterValidation(tag, fn)
}
