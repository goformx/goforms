package validation

import (
	"go.uber.org/fx"
)

// Module provides validation dependencies
var Module = fx.Options(
	fx.Provide(
		// Core validator
		New,

		// Schema generator
		NewSchemaGenerator,

		// Form validator
		fx.Annotate(
			NewFormValidator,
		),
	),
)
