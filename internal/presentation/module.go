// Package presentation provides the presentation layer components and their dependency injection setup.
package presentation

import (
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/presentation/inertia"
	"github.com/goformx/goforms/internal/presentation/view"
)

// Dependencies contains all presentation layer dependencies
type Dependencies struct {
	fx.In

	// Infrastructure
	Logger logging.Logger
	Config *config.Config
}

// Validate checks if all required dependencies are present
func (d *Dependencies) Validate() error {
	required := []struct {
		name  string
		value any
	}{
		{"Logger", d.Logger},
		{"Config", d.Config},
	}

	for _, r := range required {
		if r.value == nil {
			return errors.New(r.name + " is required")
		}
	}

	return nil
}

// NewRenderer creates a new view renderer
func NewRenderer(deps Dependencies) (view.Renderer, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return view.NewRenderer(deps.Logger), nil
}

// NewInertiaManager creates a new Inertia manager
func NewInertiaManager(deps Dependencies) (*inertia.Manager, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return inertia.NewManager(deps.Config, deps.Logger)
}

// Module provides all presentation layer dependencies
var Module = fx.Module("presentation",
	// View renderer (for backward compatibility during migration)
	fx.Provide(
		fx.Annotate(
			NewRenderer,
			fx.As(new(view.Renderer)),
		),
	),
	// Inertia manager for Vue SPA rendering
	fx.Provide(NewInertiaManager),
	fx.Provide(inertia.NewEchoHandler),
)
