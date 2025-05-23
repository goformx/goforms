package view

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// TemplateRenderer handles template rendering
type TemplateRenderer struct {
	logger logging.Logger
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(logger logging.Logger) *TemplateRenderer {
	return &TemplateRenderer{
		logger: logger,
	}
}

// Render renders a template with the given data
func (r *TemplateRenderer) Render(c echo.Context, t templ.Component) error {
	if err := t.Render(c.Request().Context(), c.Response().Writer); err != nil {
		r.logger.Error("failed to render template",
			logging.ErrorField("error", err),
			logging.StringField("template", fmt.Sprintf("%T", t)),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render page")
	}
	return nil
}
