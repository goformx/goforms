// Package application provides application creation and lifecycle management.
package application

import (
	"embed"

	"go.uber.org/fx"

	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/providers"
	"github.com/goformx/goforms/internal/domain"
	"github.com/goformx/goforms/internal/infrastructure"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/server"
	"github.com/goformx/goforms/internal/presentation"
)

// NewApplication creates a new fx application with the provided modules and options
func NewApplication(distFS embed.FS, modules ...fx.Option) *fx.App {
	baseModules := []fx.Option{
		// Provide embedded filesystem
		fx.Provide(func() embed.FS {
			return distFS
		}),
		// Include all application modules in dependency order
		config.Module,         // Config must come first as other modules depend on it
		infrastructure.Module, // Infrastructure (database, logging, etc.)
		domain.Module,         // Domain services and repositories
		Module,                // Application layer services (correct reference)
		appmiddleware.Module,  // Middleware orchestration
		presentation.Module,   // Presentation layer (handlers, templates)
		// Include OpenAPI validation provider
		providers.OpenAPIValidationProvider(),
		// Invoke setup functions
		fx.Invoke(setupApplication),
		fx.Invoke(setupLifecycle),
	}

	// Combine base modules with additional modules
	baseModules = append(baseModules, modules...)

	return fx.New(baseModules...)
}

// setupApplication initializes the application using the ApplicationSetup service.
func setupApplication(params appParams) error {
	return params.SetupService.Setup()
}

// setupLifecycle configures the application lifecycle using the LifecycleManager.
func setupLifecycle(params appParams) {
	lifecycleManager := NewLifecycleManager(LifecycleParams{
		Lifecycle: params.Lifecycle,
		Logger:    params.Logger,
		Server:    params.Server,
		Config:    params.Config,
	})
	lifecycleManager.SetupLifecycle()
}

// appParams defines the dependency injection parameters for the application.
type appParams struct {
	fx.In
	SetupService *ApplicationSetup      // Application setup service
	Lifecycle    fx.Lifecycle           // Manages application lifecycle hooks
	Logger       logging.Logger         // Application logger
	Server       server.ServerInterface // Server interface instead of concrete type
	Config       config.ConfigInterface // Config interface instead of concrete type
}
