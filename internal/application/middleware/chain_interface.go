package middleware

import (
	"context"

	"github.com/goformx/goforms/internal/application/middleware/chain"
	"github.com/goformx/goforms/internal/application/middleware/core"
)

// Chain represents a sequence of middleware that can be executed in order.
// The chain manages the execution flow and ensures proper middleware ordering.
type Chain interface {
	// Process executes the middleware chain for a given request.
	// Returns the final response after all middleware have been processed.
	Process(ctx context.Context, req Request) Response

	// Add appends middleware to the end of the chain.
	// Returns the chain for method chaining.
	Add(middleware ...Middleware) Chain

	// Insert adds middleware at a specific position in the chain.
	// Position 0 inserts at the beginning, position -1 inserts at the end.
	Insert(position int, middleware ...Middleware) Chain

	// Remove removes middleware by name from the chain.
	// Returns true if middleware was found and removed.
	Remove(name string) bool

	// Get returns middleware by name from the chain.
	// Returns nil if middleware is not found.
	Get(name string) Middleware

	// List returns all middleware in the chain in execution order.
	List() []Middleware

	// Clear removes all middleware from the chain.
	Clear() Chain

	// Length returns the number of middleware in the chain.
	Length() int
}

// NewChain creates a new chain from a list of middleware.
func NewChain(middlewares []core.Middleware) core.Chain {
	return chain.NewChainImpl(middlewares)
}
