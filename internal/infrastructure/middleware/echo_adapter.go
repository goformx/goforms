// Package middleware provides infrastructure layer middleware adapters
// for integrating framework-agnostic middleware with specific HTTP frameworks.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware/core"
)

// EchoAdapter adapts our framework-agnostic middleware to Echo's middleware interface.
// This adapter follows the adapter pattern to bridge between our clean architecture
// and Echo's framework-specific implementation.
type EchoAdapter struct {
	middleware core.Middleware
}

// NewEchoAdapter creates a new Echo adapter for the given middleware.
func NewEchoAdapter(middleware core.Middleware) *EchoAdapter {
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
			resp := a.middleware.Process(
				c.Request().Context(),
				req,
				func(ctx context.Context, r core.Request) core.Response {
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

// determineStatusCode determines the appropriate HTTP status code based on the error
func (a *EchoAdapter) determineStatusCode(err error) int {
	httpError := &echo.HTTPError{}
	if errors.As(err, &httpError) {
		return httpError.Code
	}

	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "not found"):
		return http.StatusNotFound
	case strings.Contains(errMsg, "unauthorized"):
		return http.StatusUnauthorized
	case strings.Contains(errMsg, "forbidden"):
		return http.StatusForbidden
	case strings.Contains(errMsg, "bad request"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// convertEchoError converts an Echo error to our Response interface.
func (a *EchoAdapter) convertEchoError(err error, c echo.Context) core.Response {
	// Determine appropriate status code based on error type
	statusCode := a.determineStatusCode(err)

	// Create error response
	errorResp := core.NewErrorResponse(statusCode, err)

	// Set request ID if available
	if requestID := c.Get("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			errorResp.SetRequestID(id)
		}
	}

	return errorResp
}

// convertEchoResponse converts Echo's response to our Response interface.
func (a *EchoAdapter) convertEchoResponse(c echo.Context) core.Response {
	// Get response from Echo context
	response := c.Response()

	// Create our response
	resp := core.NewResponse(response.Status)

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
		resp.SetContentLength(contentLength)
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

// applyHeaders applies response headers to Echo context
func (a *EchoAdapter) applyHeaders(c echo.Context, resp core.Response) {
	// Apply headers
	for key, values := range resp.Headers() {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}
}

// applyCookies applies response cookies to Echo context
func (a *EchoAdapter) applyCookies(c echo.Context, resp core.Response) {
	// Apply cookies
	for _, cookie := range resp.Cookies() {
		c.SetCookie(cookie)
	}
}

// handleRedirect handles redirect responses
func (a *EchoAdapter) handleRedirect(c echo.Context, resp core.Response) error {
	if resp.IsRedirect() && resp.Location() != "" {
		return c.Redirect(resp.StatusCode(), resp.Location())
	}

	return nil
}

// handleError handles error responses
func (a *EchoAdapter) handleError(resp core.Response) error {
	if resp.IsError() {
		if resp.Error() != nil {
			return fmt.Errorf("response error: %w", resp.Error())
		}
		// Create HTTP error if no specific error is set
		return echo.NewHTTPError(resp.StatusCode(), http.StatusText(resp.StatusCode()))
	}

	return nil
}

// writeBody writes the response body to Echo's response writer
func (a *EchoAdapter) writeBody(c echo.Context, resp core.Response) error {
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

// applyResponse applies our Response interface to Echo's context.
func (a *EchoAdapter) applyResponse(c echo.Context, resp core.Response) error {
	// Set status code
	c.Response().Status = resp.StatusCode()

	// Apply headers and cookies
	a.applyHeaders(c, resp)
	a.applyCookies(c, resp)

	// Handle redirects
	if err := a.handleRedirect(c, resp); err != nil {
		return err
	}

	// Handle errors
	if err := a.handleError(resp); err != nil {
		return err
	}

	// Write response body
	return a.writeBody(c, resp)
}

// EchoChainAdapter adapts our middleware chain to Echo's middleware chain.
type EchoChainAdapter struct {
	chain core.Chain
}

// NewEchoChainAdapter creates a new Echo chain adapter.
func NewEchoChainAdapter(chain core.Chain) *EchoChainAdapter {
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
	registry core.Registry
}

// NewEchoRegistryAdapter creates a new Echo registry adapter.
func NewEchoRegistryAdapter(registry core.Registry) *EchoRegistryAdapter {
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

	if err := a.registry.Register(name, wrapper); err != nil {
		return fmt.Errorf("failed to register echo middleware: %w", err)
	}

	return nil
}

// EchoMiddlewareWrapper wraps Echo middleware to implement our Middleware interface.
type EchoMiddlewareWrapper struct {
	echoMiddleware echo.MiddlewareFunc
}

// Process implements our Middleware interface by converting to Echo middleware.
func (w *EchoMiddlewareWrapper) Process(ctx context.Context, req core.Request, next core.Handler) core.Response {
	// This is a simplified implementation
	// In a real implementation, you would need to create a mock Echo context
	// and handle the conversion properly

	// For now, we'll return a simple response indicating the middleware was processed
	resp := core.NewResponse(http.StatusOK)
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
	orchestrator core.Orchestrator
}

// NewEchoOrchestratorAdapter creates a new Echo orchestrator adapter.
func NewEchoOrchestratorAdapter(orchestrator core.Orchestrator) *EchoOrchestratorAdapter {
	return &EchoOrchestratorAdapter{
		orchestrator: orchestrator,
	}
}

// SetupEchoMiddleware sets up Echo middleware based on our orchestrator configuration.
func (a *EchoOrchestratorAdapter) SetupEchoMiddleware(e *echo.Echo, chainType core.ChainType) error {
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
func (a *EchoOrchestratorAdapter) RegisterEchoChain(name string, chainType core.ChainType) error {
	chain, err := a.orchestrator.CreateChain(chainType)
	if err != nil {
		return fmt.Errorf("failed to create chain for registration: %w", err)
	}

	if registerErr := a.orchestrator.RegisterChain(name, chain); registerErr != nil {
		return fmt.Errorf("failed to register echo chain: %w", registerErr)
	}

	return nil
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
