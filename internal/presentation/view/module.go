package view

import "go.uber.org/fx"

// Module provides the view rendering module for the application
var Module = fx.Options(
	fx.Provide(
		NewRenderer,
	),
)
