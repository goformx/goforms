package handlers

import (
	"net/http"

	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// SchemaHandler handles form schema-related HTTP requests
type SchemaHandler struct {
	formService form.Service
	logger      logging.Logger
	Base        *BaseHandler
}

// NewSchemaHandler creates a new schema handler
func NewSchemaHandler(
	formService form.Service,
	logger logging.Logger,
	base *BaseHandler,
) *SchemaHandler {
	return &SchemaHandler{
		formService: formService,
		logger:      logger,
		Base:        base,
	}
}

// Register sets up the schema routes
func (h *SchemaHandler) Register(e *echo.Echo) {
	schema := e.Group("/dashboard/forms/:id/schema")
	h.Base.SetupMiddleware(schema)

	schema.GET("", h.GetFormSchema)
	schema.PUT("", h.UpdateFormSchema)
}

// GetFormSchema handles getting a form's schema
func (h *SchemaHandler) GetFormSchema(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.Base.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, formObj.Schema)
}

// UpdateFormSchema handles updating a form's schema
func (h *SchemaHandler) UpdateFormSchema(c echo.Context) error {
	currentUser, err := h.Base.getAuthenticatedUser(c)
	if err != nil {
		return err
	}

	formObj, err := h.Base.getOwnedForm(c, currentUser)
	if err != nil {
		return err
	}

	var schema form.JSON
	if err := c.Bind(&schema); err != nil {
		return h.Base.handleError(err, http.StatusBadRequest, "Invalid schema data")
	}

	if schema == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Schema data is required")
	}

	if _, ok := schema["display"]; !ok {
		schema["display"] = "form"
	}
	if _, ok := schema["components"]; !ok {
		schema["components"] = []any{}
	}

	if _, ok := schema["components"].([]any); !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Components must be an array")
	}

	formObj.Schema = schema
	formObj.UserID = currentUser.ID
	formObj.Active = true

	if err := h.formService.UpdateForm(formObj); err != nil {
		return h.Base.handleError(err, http.StatusInternalServerError, "Failed to update form schema")
	}

	return c.JSON(http.StatusOK, formObj)
}
