package providers

import (
	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	openapipkg "github.com/goformx/goforms/internal/infrastructure/openapi"
)

// OpenAPIValidationProvider provides the OpenAPI validation middleware
func OpenAPIValidationProvider() fx.Option {
	return fx.Provide(
		func(logger logging.Logger, cfg *config.Config) (*openapipkg.OpenAPIValidationMiddleware, error) {
			// Convert centralized config to middleware config
			openAPIConfig := &openapipkg.Config{
				EnableRequestValidation:  cfg.API.OpenAPI.EnableRequestValidation,
				EnableResponseValidation: cfg.API.OpenAPI.EnableResponseValidation,
				LogValidationErrors:      cfg.API.OpenAPI.LogValidationErrors,
				BlockInvalidRequests:     cfg.API.OpenAPI.BlockInvalidRequests,
				BlockInvalidResponses:    cfg.API.OpenAPI.BlockInvalidResponses,
				SkipPaths:                cfg.API.OpenAPI.SkipPaths,
				SkipMethods:              cfg.API.OpenAPI.SkipMethods,
			}

			return openapipkg.NewOpenAPIValidationMiddleware(logger, openAPIConfig)
		},
	)
}
