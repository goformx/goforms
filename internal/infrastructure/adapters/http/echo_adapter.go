package http

import (
	"context"
	"fmt"
	"strings"

	"github.com/a-h/templ"
	core "github.com/goformx/goforms/internal/application/middleware/core"
	"github.com/goformx/goforms/internal/infrastructure/view"
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	"github.com/labstack/echo/v4"
)

// EchoContextAdapter implements Context for Echo
type EchoContextAdapter struct {
	echo.Context
	renderer view.Renderer
}

// NewEchoContextAdapter creates a new Echo context adapter
func NewEchoContextAdapter(ctx echo.Context, renderer view.Renderer) *EchoContextAdapter {
	return &EchoContextAdapter{Context: ctx, renderer: renderer}
}

// Method returns the HTTP method
func (e *EchoContextAdapter) Method() string {
	return e.Request().Method
}

// Path returns the request path
func (e *EchoContextAdapter) Path() string {
	return e.Request().URL.Path
}

// Param returns a path parameter by name
func (e *EchoContextAdapter) Param(name string) string {
	return e.Context.Param(name)
}

// QueryParam returns a query parameter by name
func (e *EchoContextAdapter) QueryParam(name string) string {
	return e.Context.QueryParam(name)
}

// FormValue returns a form value by name
func (e *EchoContextAdapter) FormValue(name string) string {
	return e.Context.FormValue(name)
}

// Headers returns all request headers
func (e *EchoContextAdapter) Headers() map[string]string {
	headers := make(map[string]string)

	for key, values := range e.Request().Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return headers
}

// Body returns the request body as bytes
func (e *EchoContextAdapter) Body() []byte {
	// This would need to be implemented based on how you want to handle the body
	// For now, return empty slice
	return []byte{}
}

// JSON sends a JSON response
func (e *EchoContextAdapter) JSON(statusCode int, data any) error {
	return e.Context.JSON(statusCode, data)
}

// JSONBlob sends a JSON blob response
func (e *EchoContextAdapter) JSONBlob(statusCode int, data []byte) error {
	if err := e.Context.JSONBlob(statusCode, data); err != nil {
		return fmt.Errorf("failed to write JSON blob response: %w", err)
	}

	return nil
}

// String sends a string response
func (e *EchoContextAdapter) String(statusCode int, data string) error {
	if err := e.Context.String(statusCode, data); err != nil {
		return fmt.Errorf("failed to write string response: %w", err)
	}

	return nil
}

// Redirect redirects the request
func (e *EchoContextAdapter) Redirect(statusCode int, url string) error {
	return e.Context.Redirect(statusCode, url)
}

// NoContent sends a no content response
func (e *EchoContextAdapter) NoContent(statusCode int) error {
	if err := e.Context.NoContent(statusCode); err != nil {
		return fmt.Errorf("failed to write no content response: %w", err)
	}

	return nil
}

// RequestContext returns the underlying context.Context
func (e *EchoContextAdapter) RequestContext() context.Context {
	return e.Request().Context()
}

// GetUnderlyingContext returns the underlying Echo context for bridge methods
func (e *EchoContextAdapter) GetUnderlyingContext() any {
	return e.Context
}

// RenderComponent renders a component
func (e *EchoContextAdapter) RenderComponent(component any) error {
	// Type assert to templ.Component
	templComponent, ok := component.(templ.Component)
	if !ok {
		return fmt.Errorf("component is not a templ.Component: %T", component)
	}

	// Use the renderer service to render the templ component
	if err := e.renderer.Render(e.Context, templComponent); err != nil {
		return fmt.Errorf("failed to render component: %w", err)
	}

	return nil
}

// EchoAdapter registers handlers with an echo.Echo instance.
type EchoAdapter struct {
	e        *echo.Echo
	renderer view.Renderer
	// Pre-defined method map to reduce cyclomatic complexity
	methodMap map[string]func(string, echo.HandlerFunc) *echo.Route
	// Middleware orchestrator for applying middleware chains
	middlewareOrchestrator interface {
		BuildChainForPath(path string) (core.Chain, error)
		ConvertChainToEcho(chain core.Chain) []echo.MiddlewareFunc
	}
}

// NewEchoAdapter creates a new EchoAdapter for the given echo.Echo instance.
func NewEchoAdapter(e *echo.Echo, renderer view.Renderer) *EchoAdapter {
	adapter := &EchoAdapter{
		e:        e,
		renderer: renderer,
	}

	// Initialize method map once - using wrapper functions to match signature
	adapter.methodMap = map[string]func(string, echo.HandlerFunc) *echo.Route{
		"GET":     func(path string, h echo.HandlerFunc) *echo.Route { return e.GET(path, h) },
		"POST":    func(path string, h echo.HandlerFunc) *echo.Route { return e.POST(path, h) },
		"PUT":     func(path string, h echo.HandlerFunc) *echo.Route { return e.PUT(path, h) },
		"DELETE":  func(path string, h echo.HandlerFunc) *echo.Route { return e.DELETE(path, h) },
		"PATCH":   func(path string, h echo.HandlerFunc) *echo.Route { return e.PATCH(path, h) },
		"OPTIONS": func(path string, h echo.HandlerFunc) *echo.Route { return e.OPTIONS(path, h) },
		"HEAD":    func(path string, h echo.HandlerFunc) *echo.Route { return e.HEAD(path, h) },
	}

	return adapter
}

// SetMiddlewareOrchestrator sets the middleware orchestrator for applying middleware chains
func (a *EchoAdapter) SetMiddlewareOrchestrator(orchestrator interface {
	BuildChainForPath(path string) (core.Chain, error)
	ConvertChainToEcho(chain core.Chain) []echo.MiddlewareFunc
}) {
	a.middlewareOrchestrator = orchestrator
}

// RegisterHandler registers all routes from the given handler with Echo.
func (a *EchoAdapter) RegisterHandler(handler any) error {
	// Type assert to the Handler interface
	h, ok := handler.(httpiface.Handler)
	if !ok {
		return fmt.Errorf("handler does not implement httpiface.Handler interface")
	}

	// Register all routes from the handler
	for _, route := range h.Routes() {
		if err := a.registerRoute(route); err != nil {
			return err
		}
	}

	return nil
}

// registerRoute registers a single route with Echo
func (a *EchoAdapter) registerRoute(route httpiface.Route) error {
	echoHandler := func(c echo.Context) error {
		ctx := NewEchoContextAdapter(c, a.renderer)

		return route.Handler(ctx)
	}

	// Refactored to reduce nesting
	if a.middlewareOrchestrator == nil {
		// Log when middleware orchestrator is not available
		fmt.Printf("DEBUG: No middleware orchestrator for route %s %s\n", route.Method, route.Path)
		registerFunc, exists := a.methodMap[strings.ToUpper(route.Method)]
		if !exists {
			return fmt.Errorf("unsupported HTTP method: %s", route.Method)
		}

		registerFunc(route.Path, echoHandler)

		return nil
	}

	// Log when middleware orchestrator is being used
	fmt.Printf("DEBUG: Building middleware chain for route %s %s\n", route.Method, route.Path)

	chain, err := a.middlewareOrchestrator.BuildChainForPath(route.Path)
	if err != nil {
		fmt.Printf("DEBUG: Failed to build chain for %s %s: %v\n", route.Method, route.Path, err)
	} else {
		fmt.Printf("DEBUG: Successfully built chain for %s %s\n", route.Method, route.Path)
		echoMiddleware := a.middlewareOrchestrator.ConvertChainToEcho(chain)
		fmt.Printf("DEBUG: Converted %d middleware for %s %s\n", len(echoMiddleware), route.Method, route.Path)
		for i := len(echoMiddleware) - 1; i >= 0; i-- {
			echoHandler = echoMiddleware[i](echoHandler)
		}
	}

	registerFunc, exists := a.methodMap[strings.ToUpper(route.Method)]
	if !exists {
		return fmt.Errorf("unsupported HTTP method: %s", route.Method)
	}

	registerFunc(route.Path, echoHandler)

	return nil
}
