package providers

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// OpenAPIValidationProvider provides the OpenAPI validation middleware
func OpenAPIValidationProvider() fx.Option {
	return fx.Provide(
		func(logger logging.Logger, cfg *config.Config) (*middleware.OpenAPIValidationMiddleware, error) {
			// Convert centralized config to middleware config
			openAPIConfig := &middleware.Config{
				EnableRequestValidation:  cfg.API.OpenAPI.EnableRequestValidation,
				EnableResponseValidation: cfg.API.OpenAPI.EnableResponseValidation,
				LogValidationErrors:      cfg.API.OpenAPI.LogValidationErrors,
				BlockInvalidRequests:     cfg.API.OpenAPI.BlockInvalidRequests,
				BlockInvalidResponses:    cfg.API.OpenAPI.BlockInvalidResponses,
				SkipPaths:                cfg.API.OpenAPI.SkipPaths,
				SkipMethods:              cfg.API.OpenAPI.SkipMethods,
			}

			return middleware.NewOpenAPIValidationMiddleware(logger, openAPIConfig)
		},
	)
}
