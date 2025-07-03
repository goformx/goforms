// Package chain provides chain execution logic for middleware orchestration.
package chain

import (
	"context"

	"github.com/goformx/goforms/internal/application/middleware/core"
)

// chain implements the core.Chain interface.
type chain struct {
	middlewares []core.Middleware
}

// NewChainImpl creates a new chain from a list of middleware.
func NewChainImpl(middlewares []core.Middleware) *chain {
	return &chain{middlewares: middlewares}
}

// Process executes the middleware chain for a given request.
func (c *chain) Process(ctx context.Context, req core.Request) core.Response {
	return c.execute(ctx, req, 0)
}

// execute runs the middleware at the given index, recursively.
func (c *chain) execute(ctx context.Context, req core.Request, idx int) core.Response {
	if idx >= len(c.middlewares) {
		// End of chain: return a default response
		return nil
	}

	mw := c.middlewares[idx]

	return mw.Process(ctx, req, func(nextCtx context.Context, nextReq core.Request) core.Response {
		return c.execute(nextCtx, nextReq, idx+1)
	})
}

// Add appends middleware to the end of the chain.
func (c *chain) Add(mws ...core.Middleware) core.Chain {
	c.middlewares = append(c.middlewares, mws...)

	return c
}

// Insert adds middleware at a specific position in the chain.
func (c *chain) Insert(position int, mws ...core.Middleware) core.Chain {
	if position < 0 || position > len(c.middlewares) {
		position = len(c.middlewares)
	}

	c.middlewares = append(c.middlewares[:position], append(mws, c.middlewares[position:]...)...)

	return c
}

// Remove removes middleware by name from the chain.
func (c *chain) Remove(name string) bool {
	for i, mw := range c.middlewares {
		if mw.Name() == name {
			c.middlewares = append(c.middlewares[:i], c.middlewares[i+1:]...)

			return true
		}
	}

	return false
}

// Get returns middleware by name from the chain.
func (c *chain) Get(name string) core.Middleware {
	for _, mw := range c.middlewares {
		if mw.Name() == name {
			return mw
		}
	}

	return nil
}

// List returns all middleware in the chain in execution order.
func (c *chain) List() []core.Middleware {
	return append([]core.Middleware(nil), c.middlewares...)
}

// Clear removes all middleware from the chain.
func (c *chain) Clear() core.Chain {
	c.middlewares = nil

	return c
}

// Length returns the number of middleware in the chain.
func (c *chain) Length() int {
	return len(c.middlewares)
}
