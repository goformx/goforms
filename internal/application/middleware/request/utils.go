package request

import (
	"encoding/json"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// Utils provides common request processing utilities
type Utils struct {
	Sanitizer sanitization.ServiceInterface
}

// NewUtils creates a new request utils instance
func NewUtils(sanitizer sanitization.ServiceInterface) *Utils {
	return &Utils{
		Sanitizer: sanitizer,
	}
}

// ContentType represents the type of content in a request
type ContentType string

const (
	ContentTypeJSON      ContentType = "application/json"
	ContentTypeForm      ContentType = "application/x-www-form-urlencoded"
	ContentTypeMultipart ContentType = "multipart/form-data"
)

// ParseRequestData parses request data based on content type
func (ru *Utils) ParseRequestData(c echo.Context, target any) error {
	contentType := c.Request().Header.Get("Content-Type")

	switch {
	case contentType == string(ContentTypeJSON):
		return json.NewDecoder(c.Request().Body).Decode(target)
	default:
		// Handle form data
		return c.Bind(target)
	}
}

// SanitizeString sanitizes a string input using XSS protection
func (ru *Utils) SanitizeString(input string) string {
	return ru.Sanitizer.String(input)
}

// SanitizeEmail sanitizes an email input
func (ru *Utils) SanitizeEmail(input string) string {
	return ru.Sanitizer.Email(input)
}

// IsAJAXRequest checks if the request is an AJAX request
func (ru *Utils) IsAJAXRequest(c echo.Context) bool {
	return c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// GetContentType determines the content type of the request
func (ru *Utils) GetContentType(c echo.Context) ContentType {
	contentType := c.Request().Header.Get("Content-Type")
	switch contentType {
	case string(ContentTypeJSON):
		return ContentTypeJSON
	case string(ContentTypeMultipart):
		return ContentTypeMultipart
	default:
		return ContentTypeForm
	}
}

// AuthFormData represents common form data structure for login/signup
type AuthFormData struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

// ParseAuthData parses authentication form data
func (ru *Utils) ParseAuthData(c echo.Context) (*AuthFormData, error) {
	var data AuthFormData

	if err := ru.ParseRequestData(c, &data); err != nil {
		return nil, err
	}

	// Sanitize inputs
	data.Email = ru.SanitizeEmail(data.Email)

	return &data, nil
}
