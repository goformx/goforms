package handlers

import (
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/pages"
)

type PageHandler struct{}

func NewPageHandler() *PageHandler {
	return &PageHandler{}
}

func (h *PageHandler) HomePage(c echo.Context) error {
	return pages.Home().Render(c.Request().Context(), c.Response().Writer)
}

func (h *PageHandler) ContactPage(c echo.Context) error {
	return pages.Contact().Render(c.Request().Context(), c.Response().Writer)
}
