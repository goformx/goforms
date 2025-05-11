package handlers

import (
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// Base provides common handler functionality
type Base struct {
	Logger logging.Logger
}

// RegisterRoute is a helper method to register routes with middleware
func (b *Base) RegisterRoute(
	e *echo.Echo,
	method, path string,
	handler echo.HandlerFunc,
	middleware ...echo.MiddlewareFunc,
) {
	switch method {
	case "GET":
		e.GET(path, handler, middleware...)
	case "POST":
		e.POST(path, handler, middleware...)
	case "PUT":
		e.PUT(path, handler, middleware...)
	case "DELETE":
		e.DELETE(path, handler, middleware...)
	}
	b.Logger.Debug("registered route",
		logging.StringField("method", method),
		logging.StringField("path", path),
	)
}
