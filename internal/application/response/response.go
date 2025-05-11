package response

import (
	"net/http"

	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Response represents the standard API response format
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	logger  logging.Logger
}

// MapResponse represents a map response
type MapResponse struct {
	Data map[string]any `json:"data"`
}

// NewMapResponse creates a new map response
func NewMapResponse(data map[string]any) *MapResponse {
	return &MapResponse{
		Data: data,
	}
}

// getLogger retrieves the logger from the context
func getLogger(c echo.Context) logging.Logger {
	logger := c.Get("logger")
	if logger == nil {
		// Fallback to echo's logger if our logger is not set
		logger, err := logging.NewTestLogger()
		if err != nil {
			// If we can't create a test logger, return a no-op logger
			return logging.NewNoopLogger()
		}
		return logger
	}

	log, ok := logger.(logging.Logger)
	if !ok {
		logger, err := logging.NewTestLogger()
		if err != nil {
			// If we can't create a test logger, return a no-op logger
			return logging.NewNoopLogger()
		}
		return logger
	}
	return log
}

// Success sends a successful response with data
func Success(c echo.Context, data any) error {
	logger := getLogger(c)
	logger.Debug("sending success response",
		logging.StringField("path", c.Path()),
		logging.StringField("method", c.Request().Method),
		logging.IntField("status", http.StatusOK),
		logging.AnyField("data", data),
	)

	// Use cmp package for better comparison
	if cmp.Equal(data, nil) {
		data = struct{}{}
	}

	if err := c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   data,
	}); err != nil {
		logger.Error("failed to send success response", logging.ErrorField("error", err))
		return err
	}
	return nil
}

// Created sends a 201 response with data
func Created(c echo.Context, data any) error {
	logger := getLogger(c)
	logger.Debug("sending created response",
		logging.StringField("path", c.Path()),
		logging.StringField("method", c.Request().Method),
		logging.IntField("status", http.StatusCreated),
		logging.AnyField("data", data),
	)

	if err := c.JSON(http.StatusCreated, Response{
		Status: "success",
		Data:   data,
	}); err != nil {
		logger.Error("failed to send created response", logging.ErrorField("error", err))
		return err
	}
	return nil
}

// BadRequest sends a 400 response with error message
func BadRequest(c echo.Context, message string) error {
	logger := getLogger(c)
	logger.Debug("sending bad request response",
		logging.StringField("path", c.Path()),
		logging.StringField("method", c.Request().Method),
		logging.IntField("status", http.StatusBadRequest),
		logging.StringField("error", message),
	)

	if err := c.JSON(http.StatusBadRequest, Response{
		Status:  "error",
		Message: message,
	}); err != nil {
		logger.Error("failed to send bad request response", logging.ErrorField("error", err))
		return err
	}
	return nil
}

// NotFound sends a 404 response with error message
func NotFound(c echo.Context, message string) error {
	logger := getLogger(c)
	logger.Debug("sending not found response",
		logging.StringField("path", c.Path()),
		logging.StringField("method", c.Request().Method),
		logging.IntField("status", http.StatusNotFound),
		logging.StringField("error", message),
	)

	if err := c.JSON(http.StatusNotFound, Response{
		Status:  "error",
		Message: message,
	}); err != nil {
		logger.Error("failed to send not found response", logging.ErrorField("error", err))
		return err
	}
	return nil
}

// InternalError sends a 500 response with error message
func InternalError(c echo.Context, message string) error {
	logger := getLogger(c)
	logger.Error("sending internal error response",
		logging.StringField("path", c.Path()),
		logging.StringField("method", c.Request().Method),
		logging.IntField("status", http.StatusInternalServerError),
		logging.StringField("error", message),
	)

	if err := c.JSON(http.StatusInternalServerError, Response{
		Status:  "error",
		Message: message,
	}); err != nil {
		logger.Error("failed to send internal error response", logging.ErrorField("error", err))
		return err
	}
	return nil
}

// Unauthorized sends a 401 response with error message
func Unauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, Response{
		Status:  "error",
		Message: message,
	})
}

// Forbidden sends a 403 response with error message
func Forbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, Response{
		Status:  "error",
		Message: message,
	})
}

// SetLogger sets the logger for the response
func (r *Response) SetLogger(logger logging.Logger) error {
	r.logger = logger
	return nil
}

// Equal compares two responses for equality
func (r *Response) Equal(other *Response) bool {
	if r == nil || other == nil {
		return r == other
	}
	return r.Status == other.Status && r.Message == other.Message && r.Data == other.Data
}

// JSON sends a JSON response with the given status code and data
func JSON(c echo.Context, status int, data any) error {
	return c.JSON(status, data)
}

// JSONError sends a JSON error response
func JSONError(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{
		"error": message,
	})
}

// JSONSuccess sends a JSON success response
func JSONSuccess(c echo.Context, message string) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": message,
	})
}

// JSONCreated sends a JSON created response
func JSONCreated(c echo.Context, message string) error {
	return c.JSON(http.StatusCreated, map[string]string{
		"message": message,
	})
}

// JSONNotFound sends a JSON not found response
func JSONNotFound(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, map[string]string{
		"error": message,
	})
}

// JSONUnauthorized sends a JSON unauthorized response
func JSONUnauthorized(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{
		"error": message,
	})
}

// JSONForbidden sends a JSON forbidden response
func JSONForbidden(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, map[string]string{
		"error": message,
	})
}
