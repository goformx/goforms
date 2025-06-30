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

// Build constructs a middleware chain for the given chain type.
func (b *ChainBuilder) Build(chainType core.ChainType) (core.Chain, error) {
	var names []string

	switch chainType {
	case core.ChainTypeDefault:
		names = []string{"logging", "security", "session", "csrf", "auth", "access"}
	case core.ChainTypeAPI:
		names = []string{"logging", "security", "auth", "access"}
	case core.ChainTypeWeb:
		names = []string{"logging", "security", "session", "csrf", "auth", "access"}
	case core.ChainTypeAuth:
		names = []string{"logging", "security", "auth"}
	case core.ChainTypeAdmin:
		names = []string{"logging", "security", "auth", "access", "admin"}
	case core.ChainTypePublic:
		names = []string{"logging", "security"}
	case core.ChainTypeStatic:
		names = []string{"logging"}
	default:
		return nil, fmt.Errorf("unknown chain type: %v", chainType)
	}

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

	return NewChainImpl(middlewares), nil
}
