package response

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Response represents a standardized API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// getLogger retrieves the logger from the context
func getLogger(c echo.Context) logging.Logger {
	logger := c.Get("logger")
	if logger == nil {
		// Fallback to echo's logger if our logger is not set
		return logging.NewTestLogger()
	}
	return logger.(logging.Logger)
}

// Success returns a successful response with data
func Success(c echo.Context, data interface{}) error {
	logger := getLogger(c)
	logger.Debug("sending success response",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
		logging.Int("status", http.StatusOK),
		logging.Any("data", data),
	)

	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created returns a 201 response with data
func Created(c echo.Context, data interface{}) error {
	logger := getLogger(c)
	logger.Debug("sending created response",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
		logging.Int("status", http.StatusCreated),
		logging.Any("data", data),
	)

	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// BadRequest returns a 400 response with error message
func BadRequest(c echo.Context, message string) error {
	logger := getLogger(c)
	logger.Debug("sending bad request response",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
		logging.Int("status", http.StatusBadRequest),
		logging.String("error", message),
	)

	return c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   message,
	})
}

// NotFound returns a 404 response with error message
func NotFound(c echo.Context, message string) error {
	logger := getLogger(c)
	logger.Debug("sending not found response",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
		logging.Int("status", http.StatusNotFound),
		logging.String("error", message),
	)

	return c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error:   message,
	})
}

// InternalError returns a 500 response with error message
func InternalError(c echo.Context, message string) error {
	logger := getLogger(c)
	logger.Error("sending internal error response",
		logging.String("path", c.Path()),
		logging.String("method", c.Request().Method),
		logging.Int("status", http.StatusInternalServerError),
		logging.String("error", message),
	)

	return c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   message,
	})
}
