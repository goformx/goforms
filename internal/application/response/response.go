package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response represents a standardized API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success returns a successful response with data
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created returns a 201 response with data
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// BadRequest returns a 400 response with error message
func BadRequest(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   message,
	})
}

// NotFound returns a 404 response with error message
func NotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error:   message,
	})
}

// InternalError returns a 500 response with error message
func InternalError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   message,
	})
}
