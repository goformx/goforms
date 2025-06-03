package handlers

import (
	"net/http"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// SchemaHandler handles form schema-related HTTP requests
type SchemaHandler struct {
	formService form.Service
	userService user.Service
	config      *config.Config
	logger      logging.Logger
	Base        *BaseHandler
}

// NewSchemaHandler creates a new schema handler
func NewSchemaHandler(
	formService form.Service,
	userService user.Service,
	cfg *config.Config,
	logger logging.Logger,
	base *BaseHandler,
) *SchemaHandler {
	return &SchemaHandler{
		formService: formService,
		userService: userService,
		config:      cfg,
		logger:      logger,
		Base:        base,
	}
}

// Register sets up the schema routes
func (h *SchemaHandler) Register(e *echo.Echo) {
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(h.userService, h.logger, h.config).Middleware()

	// Dashboard routes
	schema := e.Group("/dashboard/forms/:id/schema")
	schema.Use(authMiddleware)
	schema.GET("", h.GetFormSchema)
	schema.PUT("", h.UpdateFormSchema)

	// API routes for frontend XHR - also need authentication
	apiSchema := e.Group("/api/v1/forms/:id/schema")
	apiSchema.Use(authMiddleware)
	apiSchema.GET("", h.GetFormSchema)
	apiSchema.PUT("", h.UpdateFormSchema)
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
	if bindErr := c.Bind(&schema); bindErr != nil {
		return h.Base.handleError(bindErr, http.StatusBadRequest, "Invalid schema data")
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

	if updateErr := h.formService.UpdateForm(formObj); updateErr != nil {
		return h.Base.handleError(updateErr, http.StatusInternalServerError, "Failed to update form schema")
	}

	return c.JSON(http.StatusOK, formObj)
}
