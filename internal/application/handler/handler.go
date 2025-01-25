package handler

import "github.com/labstack/echo/v4"

// Handler defines the interface for HTTP handlers
type Handler interface {
	Register(e *echo.Echo)
}
