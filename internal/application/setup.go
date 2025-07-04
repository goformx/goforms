package application

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// SetupOpenAPIValidation demonstrates how to integrate OpenAPI validation middleware
// This function can be called from your main setup to add validation to your API
func SetupOpenAPIValidation(
	e *echo.Echo,
	validationMiddleware *middleware.OpenAPIValidationMiddleware,
	logger logging.Logger,
) error {
	logger.Info("setting up OpenAPI validation middleware")

	// Create integration helper for easier setup
	integration := middleware.NewOpenAPIIntegration(validationMiddleware, logger)

	// Option 1: Add validation to all API routes
	// integration.AddToEcho(e)

	// Option 2: Add validation only to specific API groups (recommended)
	apiGroup := e.Group("/api/v1")
	integration.AddToGroup(apiGroup)

	logger.Info("OpenAPI validation middleware setup completed")

	return nil
}

// SetupOpenAPIValidationWithConfig demonstrates how to setup validation with custom configuration
func SetupOpenAPIValidationWithConfig(
	e *echo.Echo,
	validationMiddleware *middleware.OpenAPIValidationMiddleware,
	logger logging.Logger,
	config *middleware.Config,
) error {
	logger.Info("setting up OpenAPI validation middleware with custom config")

	// Update the middleware configuration
	integration := middleware.NewOpenAPIIntegration(validationMiddleware, logger)
	if err := integration.UpdateConfig(config); err != nil {
		return fmt.Errorf("update OpenAPI validation config: %w", err)
	}

	// Add to API routes
	apiGroup := e.Group("/api/v1")
	integration.AddToGroup(apiGroup)

	logger.Info("OpenAPI validation middleware setup completed with custom config")

	return nil
}
