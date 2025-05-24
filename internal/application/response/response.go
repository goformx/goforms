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

// SuccessWithHeaders sends a successful response with custom headers
func SuccessWithHeaders(c echo.Context, data any, headers map[string]string) error {
	for key, value := range headers {
		c.Response().Header().Set(key, value)
	}
	return Success(c, data)
}

// ErrorResponse sends an error response with a custom status code
func ErrorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
	})
}

// ErrorResponseWithHeaders sends an error response with custom headers
func ErrorResponseWithHeaders(c echo.Context, statusCode int, message string, headers map[string]string) error {
	for key, value := range headers {
		c.Response().Header().Set(key, value)
	}
	return ErrorResponse(c, statusCode, message)
}

// InternalError sends an internal server error response with the given message
func InternalError(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusInternalServerError, message)
}

// BadRequest sends a bad request error response
func BadRequest(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusBadRequest, message)
}

// NotFound sends a not found error response
func NotFound(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusNotFound, message)
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusUnauthorized, message)
}

// Forbidden sends a forbidden error response
func Forbidden(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusForbidden, message)
}
