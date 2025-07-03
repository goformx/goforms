// Package view provides types and utilities for rendering page data and templates.
package view

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Renderer defines the interface for rendering views
type Renderer interface {
	// Render renders a templ component to the response writer
	Render(c echo.Context, t templ.Component) error
}

// renderer handles rendering of views using templates
type renderer struct {
	logger    logging.Logger
	templates *template.Template
}

// NewRenderer creates a new view renderer with the given logger
func NewRenderer(logger logging.Logger) Renderer {
	return &renderer{
		logger:    logger,
		templates: template.New(""),
	}
}

// Render renders a templ component to the response writer
func (r *renderer) Render(c echo.Context, t templ.Component) error {
	if c == nil {
		r.logger.Error("failed to render template", "error", "nil context", "template", nil)

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render page")
	}

	if t == nil {
		r.logger.Error("failed to render template", "error", "nil component", "template", nil)

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render page")
	}

	// Set Content-Type header for HTML responses
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := t.Render(c.Request().Context(), c.Response().Writer); err != nil {
		r.logger.Error("failed to render template", "error", err, "template", fmt.Sprintf("%T", t))

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render page")
	}

	return nil
}
