package view

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

type Renderer struct {
	logger logging.Logger
}

func NewRenderer(logger logging.Logger) *Renderer {
	return &Renderer{
		logger: logger,
	}
}

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

//nolint:gochecknoglobals // This is an intentional global following fx module pattern
var Module = fx.Options(
	fx.Provide(
		NewRenderer,
	),
)
