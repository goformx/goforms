package middleware

import (
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// OpenAPIIntegration provides integration helpers for the OpenAPI validation middleware
type OpenAPIIntegration struct {
	validationMiddleware *OpenAPIValidationMiddleware
	logger               logging.Logger
}

// NewOpenAPIIntegration creates a new OpenAPI integration helper
func NewOpenAPIIntegration(
	validationMiddleware *OpenAPIValidationMiddleware,
	logger logging.Logger,
) *OpenAPIIntegration {
	return &OpenAPIIntegration{
		validationMiddleware: validationMiddleware,
		logger:               logger,
	}
}

// AddToEcho adds the OpenAPI validation middleware to an Echo instance
// This can be called after the main middleware setup to add OpenAPI validation
func (oi *OpenAPIIntegration) AddToEcho(e *echo.Echo) {
	oi.logger.Info("adding OpenAPI validation middleware")

	// Add the OpenAPI validation middleware to the Echo instance
	// This will validate all requests and responses against the OpenAPI spec
	e.Use(oi.validationMiddleware.Middleware())

	oi.logger.Info("OpenAPI validation middleware added successfully")
}

// AddToGroup adds the OpenAPI validation middleware to a specific Echo group
// This is useful for applying validation only to certain API routes
func (oi *OpenAPIIntegration) AddToGroup(group *echo.Group) {
	oi.logger.Info("adding OpenAPI validation middleware to group")

	// Add the OpenAPI validation middleware to the specific group
	group.Use(oi.validationMiddleware.Middleware())

	oi.logger.Info("OpenAPI validation middleware added to group successfully")
}

// GetValidationMiddleware returns the underlying validation middleware
// This can be used for custom integration scenarios
func (oi *OpenAPIIntegration) GetValidationMiddleware() *OpenAPIValidationMiddleware {
	return oi.validationMiddleware
}

// UpdateConfig updates the validation middleware configuration
// This allows runtime configuration changes
func (oi *OpenAPIIntegration) UpdateConfig(config *Config) error {
	oi.logger.Info("updating OpenAPI validation middleware configuration",
		"enable_request_validation", config.EnableRequestValidation,
		"enable_response_validation", config.EnableResponseValidation,
		"block_invalid_requests", config.BlockInvalidRequests,
		"block_invalid_responses", config.BlockInvalidResponses,
	)

	// Create a new middleware with the updated config
	newMiddleware, err := NewOpenAPIValidationMiddleware(oi.logger, config)
	if err != nil {
		return err
	}

	oi.validationMiddleware = newMiddleware
	oi.logger.Info("OpenAPI validation middleware configuration updated successfully")

	return nil
}
