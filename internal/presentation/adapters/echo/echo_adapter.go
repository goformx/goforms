package echo

import (
	httpiface "github.com/goformx/goforms/internal/presentation/interfaces/http"
	echo "github.com/labstack/echo/v4"
)

// EchoAdapter registers httpiface.Handlers with an echo.Echo instance.
type EchoAdapter struct {
	e *echo.Echo
}

// NewEchoAdapter creates a new EchoAdapter for the given echo.Echo instance.
func NewEchoAdapter(e *echo.Echo) *EchoAdapter {
	return &EchoAdapter{e: e}
}

// RegisterHandler registers all routes from the given handler with Echo.
func (a *EchoAdapter) RegisterHandler(handler httpiface.Handler) {
	for _, route := range handler.Routes() {
		handlerMethod := route.Handler // httpiface.HandlerMethod

		a.e.Add(route.Method, route.Path, func(c echo.Context) error {
			ctx := NewEchoContextAdapter(c)

			return handlerMethod(ctx)
		})
	}
}
