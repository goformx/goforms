// Package middleware provides infrastructure layer middleware adapters
// for integrating framework-agnostic middleware with specific HTTP frameworks.
package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	appmiddleware "github.com/goformx/goforms/internal/application/middleware"
)

// EchoAdapter adapts our framework-agnostic middleware to Echo's middleware interface.
// This adapter follows the adapter pattern to bridge between our clean architecture
// and Echo's framework-specific implementation.
type EchoAdapter struct {
	middleware appmiddleware.Middleware
}

// NewEchoAdapter creates a new Echo adapter for the given middleware.
func NewEchoAdapter(middleware appmiddleware.Middleware) *EchoAdapter {
	return &EchoAdapter{
		middleware: middleware,
	}
}

// ToEchoMiddleware converts our middleware to Echo's middleware function.
// This method handles the conversion between our Request/Response interfaces
// and Echo's echo.Context, ensuring proper error handling and context management.
func (a *EchoAdapter) ToEchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Convert Echo context to our Request interface
			req := NewEchoRequest(c)

			// Process through our middleware
			resp := a.middleware.Process(c.Request().Context(), req, func(ctx context.Context, r appmiddleware.Request) appmiddleware.Response {
				// Call next handler in Echo chain
				if err := next(c); err != nil {
					// Convert Echo error to our Response interface
					return a.convertEchoError(err, c)
				}

				// Convert Echo response to our Response interface
				return a.convertEchoResponse(c)
			})

			// Apply our response to Echo context
			return a.applyResponse(c, resp)
		}
	}
}

// convertEchoError converts an Echo error to our Response interface.
func (a *EchoAdapter) convertEchoError(err error, c echo.Context) appmiddleware.Response {
	// Determine appropriate status code based on error type
	statusCode := http.StatusInternalServerError

	if httpError, ok := err.(*echo.HTTPError); ok {
		statusCode = httpError.Code
	} else if strings.Contains(err.Error(), "not found") {
		statusCode = http.StatusNotFound
	} else if strings.Contains(err.Error(), "unauthorized") {
		statusCode = http.StatusUnauthorized
	} else if strings.Contains(err.Error(), "forbidden") {
		statusCode = http.StatusForbidden
	} else if strings.Contains(err.Error(), "bad request") {
		statusCode = http.StatusBadRequest
	}

	// Create error response
	errorResp := appmiddleware.NewErrorResponse(statusCode, err)

	// Set request ID if available
	if requestID := c.Get("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			errorResp.SetRequestID(id)
		}
	}

	return errorResp
}

// convertEchoResponse converts Echo's response to our Response interface.
func (a *EchoAdapter) convertEchoResponse(c echo.Context) appmiddleware.Response {
	// Get response from Echo context
	response := c.Response()

	// Create our response
	resp := appmiddleware.NewResponse(response.Status)

	// Copy headers
	for key, values := range response.Header() {
		for _, value := range values {
			resp.AddHeader(key, value)
		}
	}

	// Set content type
	if contentType := response.Header().Get("Content-Type"); contentType != "" {
		resp.SetContentType(contentType)
	}

	// Set content length
	if contentLength := response.Size; contentLength > 0 {
		resp.SetContentLength(int64(contentLength))
	}

	// Set request ID if available
	if requestID := c.Get("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			resp.SetRequestID(id)
		}
	}

	// Note: Echo doesn't provide direct access to the response body
	// in this context, so we can't copy it. The body will be written
	// by Echo's response writer.

	return resp
}

// applyResponse applies our Response interface to Echo's context.
func (a *EchoAdapter) applyResponse(c echo.Context, resp appmiddleware.Response) error {
	// Set status code
	c.Response().Status = resp.StatusCode()

	// Apply headers
	for key, values := range resp.Headers() {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	// Apply cookies
	for _, cookie := range resp.Cookies() {
		c.SetCookie(cookie)
	}

	// Handle redirects
	if resp.IsRedirect() && resp.Location() != "" {
		return c.Redirect(resp.StatusCode(), resp.Location())
	}

	// Handle errors
	if resp.IsError() {
		if resp.Error() != nil {
			return resp.Error()
		}
		// Create HTTP error if no specific error is set
		return echo.NewHTTPError(resp.StatusCode(), http.StatusText(resp.StatusCode()))
	}

	// Write response body if available
	if resp.Body() != nil {
		// Copy body to Echo's response writer
		if _, err := io.Copy(c.Response().Writer, resp.Body()); err != nil {
			return fmt.Errorf("failed to write response body: %w", err)
		}
	} else if resp.BodyBytes() != nil {
		// Write bytes directly
		if _, err := c.Response().Writer.Write(resp.BodyBytes()); err != nil {
			return fmt.Errorf("failed to write response body: %w", err)
		}
	}

	return nil
}

// EchoChainAdapter adapts our middleware chain to Echo's middleware chain.
type EchoChainAdapter struct {
	chain appmiddleware.Chain
}

// NewEchoChainAdapter creates a new Echo chain adapter.
func NewEchoChainAdapter(chain appmiddleware.Chain) *EchoChainAdapter {
	return &EchoChainAdapter{
		chain: chain,
	}
}

// ToEchoMiddleware converts our middleware chain to Echo's middleware function.
func (a *EchoChainAdapter) ToEchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Convert Echo context to our Request interface
			req := NewEchoRequest(c)

			// Process through our middleware chain
			resp := a.chain.Process(c.Request().Context(), req)

			// Apply our response to Echo context
			adapter := &EchoAdapter{}
			return adapter.applyResponse(c, resp)
		}
	}
}

// EchoRegistryAdapter adapts our middleware registry to Echo's middleware system.
type EchoRegistryAdapter struct {
	registry appmiddleware.Registry
}

// NewEchoRegistryAdapter creates a new Echo registry adapter.
func NewEchoRegistryAdapter(registry appmiddleware.Registry) *EchoRegistryAdapter {
	return &EchoRegistryAdapter{
		registry: registry,
	}
}

// GetEchoMiddleware retrieves middleware by name and converts it to Echo middleware.
func (a *EchoRegistryAdapter) GetEchoMiddleware(name string) (echo.MiddlewareFunc, bool) {
	middleware, exists := a.registry.Get(name)
	if !exists {
		return nil, false
	}

	adapter := NewEchoAdapter(middleware)
	return adapter.ToEchoMiddleware(), true
}

// RegisterEchoMiddleware registers Echo middleware with our registry.
func (a *EchoRegistryAdapter) RegisterEchoMiddleware(name string, echoMiddleware echo.MiddlewareFunc) error {
	// Create a wrapper that converts our interfaces to Echo middleware
	wrapper := &EchoMiddlewareWrapper{
		echoMiddleware: echoMiddleware,
	}

	return a.registry.Register(name, wrapper)
}

// EchoMiddlewareWrapper wraps Echo middleware to implement our Middleware interface.
type EchoMiddlewareWrapper struct {
	echoMiddleware echo.MiddlewareFunc
}

// Process implements our Middleware interface by converting to Echo middleware.
func (w *EchoMiddlewareWrapper) Process(ctx context.Context, req appmiddleware.Request, next appmiddleware.Handler) appmiddleware.Response {
	// This is a simplified implementation
	// In a real implementation, you would need to create a mock Echo context
	// and handle the conversion properly

	// For now, we'll return a simple response indicating the middleware was processed
	resp := appmiddleware.NewResponse(http.StatusOK)
	resp.SetContentType("text/plain")
	resp.SetBodyBytes([]byte("Echo middleware processed"))

	// Call the next handler
	return next(ctx, req)
}

// Name returns the name of this middleware wrapper.
func (w *EchoMiddlewareWrapper) Name() string {
	return "echo-middleware-wrapper"
}

// Priority returns the priority of this middleware wrapper.
func (w *EchoMiddlewareWrapper) Priority() int {
	return 0 // Default priority
}

// EchoOrchestratorAdapter adapts our orchestrator to Echo's middleware system.
type EchoOrchestratorAdapter struct {
	orchestrator appmiddleware.Orchestrator
}

// NewEchoOrchestratorAdapter creates a new Echo orchestrator adapter.
func NewEchoOrchestratorAdapter(orchestrator appmiddleware.Orchestrator) *EchoOrchestratorAdapter {
	return &EchoOrchestratorAdapter{
		orchestrator: orchestrator,
	}
}

// SetupEchoMiddleware sets up Echo middleware based on our orchestrator configuration.
func (a *EchoOrchestratorAdapter) SetupEchoMiddleware(e *echo.Echo, chainType appmiddleware.ChainType) error {
	// Create middleware chain
	chain, err := a.orchestrator.CreateChain(chainType)
	if err != nil {
		return fmt.Errorf("failed to create middleware chain: %w", err)
	}

	// Convert to Echo middleware
	adapter := NewEchoChainAdapter(chain)
	echoMiddleware := adapter.ToEchoMiddleware()

	// Apply to Echo
	e.Use(echoMiddleware)

	return nil
}

// RegisterEchoChain registers a named chain for use with Echo.
func (a *EchoOrchestratorAdapter) RegisterEchoChain(name string, chainType appmiddleware.ChainType) error {
	chain, err := a.orchestrator.CreateChain(chainType)
	if err != nil {
		return fmt.Errorf("failed to create chain for registration: %w", err)
	}

	return a.orchestrator.RegisterChain(name, chain)
}

// GetEchoChain retrieves a named chain and converts it to Echo middleware.
func (a *EchoOrchestratorAdapter) GetEchoChain(name string) (echo.MiddlewareFunc, bool) {
	chain, exists := a.orchestrator.GetChain(name)
	if !exists {
		return nil, false
	}

	adapter := NewEchoChainAdapter(chain)
	return adapter.ToEchoMiddleware(), true
}
