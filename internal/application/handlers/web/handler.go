package web

import "github.com/labstack/echo/v4"

// Handler defines the interface for web handlers
type Handler interface {
	// Register registers the handler's routes with the Echo instance
	Register(e *echo.Echo)
}
