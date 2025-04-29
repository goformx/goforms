package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/application/validation"
	"github.com/labstack/echo/v4"
)

type ValidationHandler struct{}

func NewValidationHandler() *ValidationHandler {
	return &ValidationHandler{}
}

func (h *ValidationHandler) GetValidationRules(c echo.Context) error {
	schemaName := c.Param("schema")
	
	var schema validation.ValidationSchema
	switch schemaName {
	case "signup":
		schema = validation.GetSignupSchema()
	case "login":
		schema = validation.GetLoginSchema()
	default:
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Validation schema not found"})
	}

	return c.JSON(http.StatusOK, schema)
} 