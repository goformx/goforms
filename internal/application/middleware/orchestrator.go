// Package middleware provides the orchestrator for managing middleware chains and orchestration.
package middleware

import (
	"fmt"
	"sync"

	"github.com/goformx/goforms/internal/application/middleware/chain"
	"github.com/goformx/goforms/internal/application/middleware/core"
)

// orchestrator implements the Orchestrator interface.
type orchestrator struct {
	registry   core.Registry
	chains     map[string]core.Chain
	chainsLock sync.RWMutex
	logger     core.Logger // Optional: for logging orchestration events
}

// NewOrchestrator creates a new middleware orchestrator with the given registry and logger.
func NewOrchestrator(registry core.Registry, logger core.Logger) Orchestrator {
	return &orchestrator{
		registry: registry,
		chains:   make(map[string]core.Chain),
		logger:   logger,
	}
}

// CreateChain creates a new middleware chain for the specified type.
func (o *orchestrator) CreateChain(chainType core.ChainType) (core.Chain, error) {
	// Use the chain builder to construct the chain
	chainObj, err := chain.NewChainBuilder(o.registry, o.logger).Build(chainType)
	if err != nil {
		o.logger.Error("failed to build chain: %v", err)
		return nil, err
	}
	return chainObj, nil
}

// GetChain retrieves a pre-configured chain by name.
func (o *orchestrator) GetChain(name string) (core.Chain, bool) {
	o.chainsLock.RLock()
	defer o.chainsLock.RUnlock()
	chain, ok := o.chains[name]
	return chain, ok
}

// RegisterChain registers a named chain for later retrieval.
func (o *orchestrator) RegisterChain(name string, chain core.Chain) error {
	o.chainsLock.Lock()
	defer o.chainsLock.Unlock()
	if _, exists := o.chains[name]; exists {
		return fmt.Errorf("chain with name %q already exists", name)
	}
	o.chains[name] = chain
	return nil
}

// ListChains returns all registered chain names.
func (o *orchestrator) ListChains() []string {
	o.chainsLock.RLock()
	defer o.chainsLock.RUnlock()
	names := make([]string, 0, len(o.chains))
	for name := range o.chains {
		names = append(names, name)
	}
	return names
}

// RemoveChain removes a chain by name.
func (o *orchestrator) RemoveChain(name string) bool {
	o.chainsLock.Lock()
	defer o.chainsLock.Unlock()
	if _, exists := o.chains[name]; exists {
		delete(o.chains, name)
		return true
	}
	return false
}
