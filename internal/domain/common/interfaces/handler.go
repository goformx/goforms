package interfaces

import "github.com/labstack/echo/v4"

// WebHandler represents a web handler that can register routes with an Echo instance.
// This is different from the middleware Handler which is a function type.
type WebHandler interface {
	Register(*echo.Echo)
}
