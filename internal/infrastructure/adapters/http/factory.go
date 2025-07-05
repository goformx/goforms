package http

import (
	"github.com/goformx/goforms/internal/infrastructure/view"
	"github.com/labstack/echo/v4"
)

// AdapterFactory provides a centralized way to create HTTP adapters
type AdapterFactory struct {
	echo     *echo.Echo
	renderer view.Renderer
}

// NewAdapterFactory creates a new adapter factory
func NewAdapterFactory(e *echo.Echo, renderer view.Renderer) *AdapterFactory {
	return &AdapterFactory{
		echo:     e,
		renderer: renderer,
	}
}

// CreateEchoAdapter creates a new EchoAdapter
func (f *AdapterFactory) CreateEchoAdapter() *EchoAdapter {
	return NewEchoAdapter(f.echo, f.renderer)
}

// CreateRequestAdapter creates a new RequestAdapter
func (f *AdapterFactory) CreateRequestAdapter() RequestAdapter {
	return NewEchoRequestAdapter()
}

// CreateResponseAdapter creates a new ResponseAdapter
func (f *AdapterFactory) CreateResponseAdapter() ResponseAdapter {
	return NewEchoResponseAdapter()
}

// CreateContextAdapter creates a new ContextAdapter from an Echo context
func (f *AdapterFactory) CreateContextAdapter(ctx echo.Context) *EchoContextAdapter {
	return NewEchoContextAdapter(ctx, f.renderer)
}
