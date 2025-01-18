package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response represents a standardized API response
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Success sends a successful response
func Success(c echo.Context, code int, data interface{}) error {
	return c.JSON(code, Response{
		Status: "success",
		Data:   data,
	})
}

// Error sends an error response
func Error(c echo.Context, code int, message string) error {
	return c.JSON(code, Response{
		Status:  "error",
		Message: message,
	})
}

// Created sends a 201 Created response
func Created(c echo.Context, data interface{}) error {
	return Success(c, http.StatusCreated, data)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, message)
}

// NotFound sends a 404 Not Found response
func NotFound(c echo.Context, message string) error {
	return Error(c, http.StatusNotFound, message)
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c echo.Context, message string) error {
	return Error(c, http.StatusInternalServerError, message)
}
