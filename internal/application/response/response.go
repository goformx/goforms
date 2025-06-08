package response

import (
	"net/http"

	"github.com/goformx/goforms/internal/presentation/templates/pages"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
)

// APIResponse represents a standardized API response structure
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Success sends a successful response with the given data
func Success(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse sends an error response with a custom status code
func ErrorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
	})
}

// WebErrorResponse renders an error page for web handlers
func WebErrorResponse(c echo.Context, renderer *view.Renderer, statusCode int, message string) error {
	data := shared.BuildPageData(nil, "Error")
	data.Error = message
	return renderer.Render(c, pages.Error(data))
}
