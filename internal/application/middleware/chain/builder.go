// Package chain provides chain building logic for middleware orchestration.
package chain

import (
	"fmt"

	"github.com/goformx/goforms/internal/application/middleware/core"
)

// ChainBuilder builds middleware chains based on type and registry.
type ChainBuilder struct {
	registry core.Registry
	logger   core.Logger // Optional
}

// NewChainBuilder creates a new chain builder.
func NewChainBuilder(registry core.Registry, logger core.Logger) *ChainBuilder {
	return &ChainBuilder{
		registry: registry,
		logger:   logger,
	}
}

// chainConfig defines the middleware configuration for each chain type
type chainConfig struct {
	middlewareNames []string
	description     string
}

// getChainConfig returns the configuration for a given chain type
func (b *ChainBuilder) getChainConfig(chainType core.ChainType) (chainConfig, error) {
	configs := map[core.ChainType]chainConfig{
		core.ChainTypeDefault: {
			middlewareNames: []string{"logging", "security", "session", "csrf", "auth", "access"},
			description:     "Default middleware chain for most requests",
		},
		core.ChainTypeAPI: {
			middlewareNames: []string{"logging", "security", "auth", "access"},
			description:     "Middleware chain for API requests",
		},
		core.ChainTypeWeb: {
			middlewareNames: []string{"logging", "security", "session", "csrf", "auth", "access"},
			description:     "Middleware chain for web page requests",
		},
		core.ChainTypeAuth: {
			middlewareNames: []string{"logging", "security", "auth"},
			description:     "Middleware chain for authentication endpoints",
		},
		core.ChainTypeAdmin: {
			middlewareNames: []string{"logging", "security", "auth", "access", "admin"},
			description:     "Middleware chain for admin-only endpoints",
		},
		core.ChainTypePublic: {
			middlewareNames: []string{"logging", "security"},
			description:     "Middleware chain for public endpoints",
		},
		core.ChainTypeStatic: {
			middlewareNames: []string{"logging"},
			description:     "Middleware chain for static asset requests",
		},
	}

	config, exists := configs[chainType]
	if !exists {
		return chainConfig{}, fmt.Errorf("unknown chain type: %v", chainType)
	}

	return config, nil
}

// resolveMiddleware resolves middleware names to actual middleware instances
func (b *ChainBuilder) resolveMiddleware(names []string) []core.Middleware {
	middlewares := make([]core.Middleware, 0, len(names))

	for _, name := range names {
		mw, ok := b.registry.Get(name)
		if !ok {
			if b.logger != nil {
				b.logger.Warn("middleware %q not found in registry", name)
			}

			continue
		}

		middlewares = append(middlewares, mw)
	}

	return middlewares
}

// Build constructs a middleware chain for the given chain type.
func (b *ChainBuilder) Build(chainType core.ChainType) (core.Chain, error) {
	config, err := b.getChainConfig(chainType)
	if err != nil {
		return nil, err
	}

	middlewares := b.resolveMiddleware(config.middlewareNames)

	if b.logger != nil {
		b.logger.Info("built chain",
			"chain_type", chainType.String(),
			"middleware_count", len(middlewares),
			"description", config.description)
	}

	return NewChainImpl(middlewares), nil
}
