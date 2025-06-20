package response

import (
	"net/http"

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
