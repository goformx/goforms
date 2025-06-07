package application

import (
	"github.com/goformx/goforms/internal/application/services/auth"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Module represents the application module
type Module struct {
	userService user.Service
	formService form.Service
	logger      logging.Logger
	services    ServiceContainer
}

// ServiceContainer holds all application services
type ServiceContainer struct {
	AuthService auth.Service
}

// NewModule creates a new application module
func NewModule(
	userService user.Service,
	formService form.Service,
	logger logging.Logger,
) *Module {
	m := &Module{
		userService: userService,
		formService: formService,
		logger:      logger,
	}
	m.initializeServices()
	return m
}

// initializeServices initializes all application services
func (m *Module) initializeServices() {
	m.services.AuthService = auth.NewService(m.userService, m.logger)
}

// GetServices returns the service container
func (m *Module) GetServices() ServiceContainer {
	return m.services
}
