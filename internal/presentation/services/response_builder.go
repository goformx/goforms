package services

import (
	"net/http"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/labstack/echo/v4"
)

// ResponseBuilder handles consistent response building
type ResponseBuilder struct {
	logger logging.Logger
}

// NewResponseBuilder creates a new response builder
func NewResponseBuilder(logger logging.Logger) *ResponseBuilder {
	return &ResponseBuilder{
		logger: logger,
	}
}

// BuildJSONResponse builds a JSON response with the given data and status code
func (b *ResponseBuilder) BuildJSONResponse(c echo.Context, data any, status int) error {
	if status == 0 {
		status = http.StatusOK
	}

	b.logger.Debug("building JSON response",
		logging.IntField("status", status),
		logging.StringField("operation", "response_building"),
	)

	return c.JSON(status, data)
}

// BuildErrorResponse builds an error response with the given error, status code, and message
func (b *ResponseBuilder) BuildErrorResponse(c echo.Context, err error, status int, message string) error {
	if status == 0 {
		status = http.StatusInternalServerError
	}

	b.logger.Error("building error response",
		logging.Error(err),
		logging.IntField("status", status),
		logging.StringField("message", message),
		logging.StringField("operation", "error_response"),
	)

	return echo.NewHTTPError(status, message)
}

// BuildRedirectResponse builds a redirect response to the given path with the specified status code
func (b *ResponseBuilder) BuildRedirectResponse(c echo.Context, path string, status int) error {
	if status == 0 {
		status = http.StatusSeeOther
	}

	b.logger.Debug("building redirect response",
		logging.StringField("path", path),
		logging.IntField("status", status),
		logging.StringField("operation", "redirect_response"),
	)

	return c.Redirect(status, path)
}

// BuildHTMLResponse builds an HTML response with the given template and data
func (b *ResponseBuilder) BuildHTMLResponse(c echo.Context, template string, data any, status int) error {
	if status == 0 {
		status = http.StatusOK
	}

	b.logger.Debug("building HTML response",
		logging.StringField("template", template),
		logging.IntField("status", status),
		logging.StringField("operation", "html_response"),
	)

	return c.Render(status, template, data)
}

// BuildNoContentResponse builds a response with no content and the specified status code
func (b *ResponseBuilder) BuildNoContentResponse(c echo.Context, status int) error {
	if status == 0 {
		status = http.StatusNoContent
	}

	b.logger.Debug("building no content response",
		logging.IntField("status", status),
		logging.StringField("operation", "no_content_response"),
	)

	return c.NoContent(status)
}

// BuildValidationErrorResponse builds a validation error response
func (b *ResponseBuilder) BuildValidationErrorResponse(c echo.Context, err error) error {
	b.logger.Error("building validation error response",
		logging.Error(err),
		logging.StringField("operation", "validation_error"),
	)

	return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
}

// BuildNotFoundResponse builds a not found error response
func (b *ResponseBuilder) BuildNotFoundResponse(c echo.Context, message string) error {
	if message == "" {
		message = "Resource not found"
	}

	b.logger.Error("building not found response",
		logging.StringField("message", message),
		logging.StringField("operation", "not_found_error"),
	)

	return echo.NewHTTPError(http.StatusNotFound, message)
}

// BuildForbiddenResponse builds a forbidden error response
func (b *ResponseBuilder) BuildForbiddenResponse(c echo.Context, message string) error {
	if message == "" {
		message = "Access denied"
	}

	b.logger.Error("building forbidden response",
		logging.StringField("message", message),
		logging.StringField("operation", "forbidden_error"),
	)

	return echo.NewHTTPError(http.StatusForbidden, message)
}
