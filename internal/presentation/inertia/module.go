// Package inertia provides FX module for Gonertia integration.
package inertia

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Module provides Gonertia/Inertia.js dependencies for FX.
var Module = fx.Module("inertia",
	fx.Provide(
		NewManager,
		NewEchoHandler,
	),
)

// Params contains the dependencies needed to create the Inertia manager.
type Params struct {
	fx.In

	Config *config.Config
	Logger logging.Logger
}

// Result contains the Inertia manager that will be provided to the application.
type Result struct {
	fx.Out

	Manager     *Manager
	EchoHandler *EchoHandler
}

// ProvideInertia creates the Inertia manager with its dependencies.
func ProvideInertia(p Params) (Result, error) {
	manager, err := NewManager(p.Config, p.Logger)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Manager:     manager,
		EchoHandler: NewEchoHandler(manager),
	}, nil
}
