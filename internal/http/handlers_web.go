package http

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/presentation/templates/pages"
)

func (h *Handlers) handleHome(c echo.Context) error {
	return pages.Home().Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handlers) handleContact(c echo.Context) error {
	return pages.Contact().Render(c.Request().Context(), c.Response().Writer)
}
