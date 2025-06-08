package view

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// Renderer handles rendering of views using templates
type Renderer struct {
	logger    logging.Logger
	templates *template.Template
}

// NewRenderer creates a new view renderer with the given logger
func NewRenderer(logger logging.Logger) *Renderer {
	return &Renderer{
		logger:    logger,
		templates: template.New(""),
	}
}

// Render renders a templ component to the response writer
func (r *Renderer) Render(c echo.Context, t templ.Component) error {
	if err := t.Render(c.Request().Context(), c.Response().Writer); err != nil {
		r.logger.Error("failed to render template",
			logging.Error(err),
			logging.String("template", fmt.Sprintf("%T", t)),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render page")
	}
	return nil
}
