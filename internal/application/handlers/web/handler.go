package web

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/session"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Handler defines the interface for web handlers.
// Each handler is responsible for a specific set of related routes.
// Handlers should:
// 1. Be focused on a single domain area (e.g., forms, auth, dashboard)
// 2. Use the HandlerDeps struct for common dependencies
// 3. Implement proper error handling and logging
// 4. Follow consistent route naming patterns
// 5. Use the context helpers for user data access
type Handler interface {
	// Register registers the handler's routes with the Echo instance.
	// This method should be called by the RegisterHandlers function in module.go.
	// The handler should not apply middleware directly - this is handled by RegisterHandlers.
	// Routes should be grouped logically and follow RESTful patterns where appropriate.
	Register(e *echo.Echo)

	// Start initializes the handler and any required resources.
	// This is called during application startup.
	Start(ctx context.Context) error

	// Stop cleans up any resources used by the handler.
	// This is called during application shutdown.
	Stop(ctx context.Context) error
}

// HandlerDeps contains dependencies for web handlers.
// This struct is embedded in each handler to provide access to these dependencies.
// All handlers should use these dependencies through this struct rather than
// accessing them directly or creating their own instances.
type HandlerDeps struct {
	// Logger provides logging capabilities for structured logging
	Logger logging.Logger
	// Config provides application configuration and settings
	Config *config.Config
	// SessionManager handles user sessions and authentication state
	SessionManager *session.Manager
	// MiddlewareManager manages middleware configuration and setup
	MiddlewareManager *middleware.Manager
	// UserService provides user-related operations and business logic
	UserService user.Service
	// FormService provides form-related operations and business logic
	FormService form.Service
}

// validateField checks if a field is nil and returns an error if it is.
// This is used internally by Validate to check each required dependency.
func (d *HandlerDeps) validateField(name string, value any) error {
	if value == nil {
		return errors.New(name + " is required")
	}

	return nil
}

// Validate checks if all required dependencies are present.
// This should be called when creating a new handler to ensure
// all required dependencies are properly initialized.
func (d *HandlerDeps) Validate() error {
	required := []struct {
		name  string
		value any
	}{
		{"UserService", d.UserService},
		{"FormService", d.FormService},
		{"Logger", d.Logger},
		{"Config", d.Config},
		{"SessionManager", d.SessionManager},
		{"MiddlewareManager", d.MiddlewareManager},
	}

	for _, r := range required {
		if err := d.validateField(r.name, r.value); err != nil {
			return err
		}
	}

	return nil
}

// HandlerParams contains parameters for creating a handler.
// This struct is used to pass dependencies to NewHandlerDeps
// in a type-safe and explicit way.
type HandlerParams struct {
	UserService       user.Service
	FormService       form.Service
	Logger            logging.Logger
	Config            *config.Config
	SessionManager    *session.Manager
	MiddlewareManager *middleware.Manager
}

// NewHandlerDeps creates a new HandlerDeps instance.
// This factory function ensures that all dependencies are properly
// initialized and validated before being used by a handler.
func NewHandlerDeps(params *HandlerParams) (*HandlerDeps, error) {
	deps := &HandlerDeps{
		UserService:       params.UserService,
		FormService:       params.FormService,
		Logger:            params.Logger,
		Config:            params.Config,
		SessionManager:    params.SessionManager,
		MiddlewareManager: params.MiddlewareManager,
	}

	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return deps, nil
}

// Start initializes the handler dependencies.
func (d *HandlerDeps) Start(_ context.Context) error {
	return nil // No initialization needed
}

// Stop cleans up any resources used by the handler dependencies.
func (d *HandlerDeps) Stop(_ context.Context) error {
	return nil // No cleanup needed
}
