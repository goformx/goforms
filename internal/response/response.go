package response

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success returns a successful response with data
func Success(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, Response{
		Success: true,
		Data:    data,
	})
}

// Error returns an error response
func Error(c echo.Context, status int, message string) error {
	return c.JSON(status, Response{
		Success: false,
		Error:   message,
	})
}

// ParseInt64Param parses an int64 parameter from the URL path
func ParseInt64Param(c echo.Context, param string) (int64, error) {
	val := c.Param(param)
	if val == "" {
		return 0, echo.NewHTTPError(400, "missing "+param+" parameter")
	}

	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(400, "invalid "+param+" parameter")
	}

	return id, nil
}
