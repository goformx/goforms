package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/infrastructure/openapi"
)

// OpenAPIHandler handles OpenAPI documentation requests
type OpenAPIHandler struct {
	*BaseHandler
	spec *openapi3.T
}

// NewOpenAPIHandler creates a new OpenAPIHandler
func NewOpenAPIHandler(base *BaseHandler) *OpenAPIHandler {
	// Parse the OpenAPI specification
	loader := openapi3.NewLoader()

	spec, err := loader.LoadFromData([]byte(openapi.OpenAPISpec))
	if err != nil {
		// Log error but don't fail - we'll serve the raw spec as fallback
		base.Logger.Error("failed to parse OpenAPI spec", "error", err)
	}

	return &OpenAPIHandler{
		BaseHandler: base,
		spec:        spec,
	}
}

// RegisterRoutes registers OpenAPI documentation routes
func (h *OpenAPIHandler) RegisterRoutes(e *echo.Echo) {
	// API documentation routes
	api := e.Group(constants.PathAPIv1)

	// Health check endpoint
	api.GET("/health", h.serveHealthCheck)

	// Serve OpenAPI specification
	api.GET("/openapi.yaml", h.serveOpenAPISpec)
	api.GET("/openapi.json", h.serveOpenAPIJSON)

	// Serve documentation UI
	api.GET("/docs", h.serveDocsUI)
	api.GET("/docs/*", h.serveDocsUI)

	// OpenAPI validation endpoint
	api.POST("/validate", h.validateOpenAPISpec)
}

// Register registers the OpenAPIHandler with the Echo instance
func (h *OpenAPIHandler) Register(_ *echo.Echo) {
	// Routes are registered by RegisterHandlers function
	// This method is required to satisfy the Handler interface
}

// serveOpenAPISpec serves the OpenAPI YAML specification
func (h *OpenAPIHandler) serveOpenAPISpec(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/x-yaml")

	if err := c.String(http.StatusOK, openapi.OpenAPISpec); err != nil {
		return fmt.Errorf("serve openapi spec: %w", err)
	}

	return nil
}

// serveOpenAPIJSON serves the OpenAPI JSON specification
func (h *OpenAPIHandler) serveOpenAPIJSON(c echo.Context) error {
	if h.spec == nil {
		// Fallback to raw spec if parsing failed
		c.Response().Header().Set("Content-Type", "application/json")

		if err := c.String(http.StatusOK, openapi.OpenAPISpec); err != nil {
			return fmt.Errorf("serve openapi spec fallback: %w", err)
		}

		return nil
	}

	// Convert the parsed spec to JSON
	jsonData, err := json.MarshalIndent(h.spec, "", "  ")
	if err != nil {
		h.Logger.Error("failed to marshal OpenAPI spec to JSON", "error", err)
		// Fallback to raw spec
		c.Response().Header().Set("Content-Type", "application/json")

		if fallbackErr := c.String(http.StatusOK, openapi.OpenAPISpec); fallbackErr != nil {
			return fmt.Errorf("serve openapi spec fallback: %w", fallbackErr)
		}

		return nil
	}

	c.Response().Header().Set("Content-Type", "application/json")

	if jsonErr := c.JSONBlob(http.StatusOK, jsonData); jsonErr != nil {
		return fmt.Errorf("serve openapi json: %w", jsonErr)
	}

	return nil
}

// serveHealthCheck serves the health check endpoint
func (h *OpenAPIHandler) serveHealthCheck(c echo.Context) error {
	healthData := map[string]any{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
		"version":   "1.0.0",
	}

	return c.JSON(http.StatusOK, healthData)
}

// validateOpenAPISpec validates an OpenAPI specification
func (h *OpenAPIHandler) validateOpenAPISpec(c echo.Context) error {
	var request struct {
		Spec string `json:"spec"`
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request format",
		})
	}

	// Parse the provided specification
	loader := openapi3.NewLoader()

	spec, parseErr := loader.LoadFromData([]byte(request.Spec))
	if parseErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":   "Failed to parse OpenAPI specification",
			"details": parseErr.Error(),
		})
	}

	// Validate the specification
	if validateErr := spec.Validate(c.Request().Context()); validateErr != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error":   "OpenAPI specification validation failed",
			"details": validateErr.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"valid":   true,
		"message": "OpenAPI specification is valid",
	})
}

// serveDocsUI serves the OpenAPI documentation UI
func (h *OpenAPIHandler) serveDocsUI(c echo.Context) error {
	// Create a simple HTML page that embeds Swagger UI
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoFormX API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/v1/openapi.yaml',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                onComplete: function() {
                    console.log('Swagger UI loaded successfully');
                }
            });
        };
    </script>
</body>
</html>`

	c.Response().Header().Set("Content-Type", "text/html")

	if err := c.String(http.StatusOK, html); err != nil {
		return fmt.Errorf("serve docs ui: %w", err)
	}

	return nil
}

// Start initializes the OpenAPI handler
func (h *OpenAPIHandler) Start(ctx context.Context) error {
	// Validate the OpenAPI specification if it was parsed successfully
	if h.spec != nil {
		if err := h.spec.Validate(ctx); err != nil {
			h.Logger.Warn("OpenAPI specification validation failed", "error", err)
		} else {
			h.Logger.Info("OpenAPI specification validated successfully")
		}
	}

	return nil
}

// Stop cleans up any resources used by the OpenAPI handler
func (h *OpenAPIHandler) Stop(_ context.Context) error {
	return nil
}
