package validation

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/domain/common/interfaces"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Module provides validation dependencies
var Module = fx.Options(
	fx.Provide(
		// Core validator
		func() (interfaces.Validator, error) {
			return New()
		},

		// Schema generator
		func() *SchemaGenerator {
			return NewSchemaGenerator()
		},

		// Form validator
		fx.Annotate(
			func(logger logging.Logger) *FormValidator {
				return NewFormValidator(logger)
			},
		),
	),
)
