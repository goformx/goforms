// Package presentation provides the presentation layer components and their dependency injection setup.
package presentation

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/presentation/view"
)

// Module provides all presentation layer dependencies
var Module = fx.Options(
	// View rendering
	fx.Provide(
		view.NewRenderer,
	),
)
