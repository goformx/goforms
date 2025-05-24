package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Success sends a successful response with the given data
func Success(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, data)
}

// InternalError sends an internal server error response with the given message
func InternalError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, map[string]string{
		"error": message,
	})
}
