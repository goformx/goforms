package validation

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

type ValidationHandler struct {
	base handlers.Base
}

func NewValidationHandler(logger logging.Logger) *ValidationHandler {
	return &ValidationHandler{
		base: handlers.Base{
			Logger: logger,
		},
	}
}

func (h *ValidationHandler) Register(e *echo.Echo) {
	h.base.RegisterRoute(e, "GET", "/validation/:schema", h.GetValidationRules)
}

func (h *ValidationHandler) GetValidationRules(c echo.Context) error {
	return c.JSON(http.StatusNotFound, map[string]string{"error": "validation schemas are not available"})
}
