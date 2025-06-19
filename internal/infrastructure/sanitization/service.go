package sanitization

import (
	"reflect"
	"strings"

	"github.com/mrz1836/go-sanitize"
)

// Service provides sanitization functionality for various input types
type Service struct{}

// NewService creates a new sanitization service
func NewService() *Service {
	return &Service{}
}

// String sanitizes a string input using XSS protection
func (s *Service) String(input string) string {
	return sanitize.XSS(input)
}

// Email sanitizes an email address
func (s *Service) Email(input string) string {
	return sanitize.Email(input, false)
}

// URL sanitizes a URL
func (s *Service) URL(input string) string {
	return sanitize.URL(input)
}

// HTML sanitizes HTML content
func (s *Service) HTML(input string) string {
	return sanitize.HTML(input)
}

// Path sanitizes a file path
func (s *Service) Path(input string) string {
	return sanitize.PathName(input)
}

// IPAddress sanitizes an IP address
func (s *Service) IPAddress(input string) string {
	return sanitize.IPAddress(input)
}

// Domain sanitizes a domain name
func (s *Service) Domain(input string) (string, error) {
	return sanitize.Domain(input, false, false)
}

// URI sanitizes a URI
func (s *Service) URI(input string) string {
	return sanitize.URI(input)
}

// Alpha sanitizes to alpha characters only
func (s *Service) Alpha(input string, spaces bool) string {
	return sanitize.Alpha(input, spaces)
}

// AlphaNumeric sanitizes to alphanumeric characters only
func (s *Service) AlphaNumeric(input string, spaces bool) string {
	return sanitize.AlphaNumeric(input, spaces)
}

// Numeric sanitizes to numeric characters only
func (s *Service) Numeric(input string) string {
	return sanitize.Numeric(input)
}

// SingleLine removes newlines and extra whitespace
func (s *Service) SingleLine(input string) string {
	return sanitize.SingleLine(input)
}

// Scripts removes script tags
func (s *Service) Scripts(input string) string {
	return sanitize.Scripts(input)
}

// XML sanitizes XML content
func (s *Service) XML(input string) string {
	return sanitize.XML(input)
}

// TrimAndSanitize trims whitespace and sanitizes a string
func (s *Service) TrimAndSanitize(input string) string {
	return s.String(strings.TrimSpace(input))
}

// TrimAndSanitizeEmail trims whitespace and sanitizes an email
func (s *Service) TrimAndSanitizeEmail(input string) string {
	return s.Email(strings.TrimSpace(input))
}

// SanitizeMap sanitizes all string values in a map
func (s *Service) SanitizeMap(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = s.String(v)
		case map[string]interface{}:
			s.SanitizeMap(v)
		case []interface{}:
			s.SanitizeSlice(v)
		}
	}
}

// SanitizeSlice sanitizes all string values in a slice
func (s *Service) SanitizeSlice(data []interface{}) {
	for i, value := range data {
		switch v := value.(type) {
		case string:
			data[i] = s.String(v)
		case map[string]interface{}:
			s.SanitizeMap(v)
		case []interface{}:
			s.SanitizeSlice(v)
		}
	}
}

// SanitizeStruct sanitizes all string fields in a struct
func (s *Service) SanitizeStruct(obj interface{}) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			if field.CanSet() {
				field.SetString(s.String(field.String()))
			}
		case reflect.Struct:
			if field.CanAddr() {
				s.SanitizeStruct(field.Addr().Interface())
			}
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				for j := 0; j < field.Len(); j++ {
					if field.Index(j).CanSet() {
						field.Index(j).SetString(s.String(field.Index(j).String()))
					}
				}
			}
		}
	}
}

// SanitizeFormData sanitizes form data with specific field types
func (s *Service) SanitizeFormData(data map[string]string, fieldTypes map[string]string) map[string]string {
	sanitized := make(map[string]string)

	for key, value := range data {
		fieldType, exists := fieldTypes[key]
		if !exists {
			fieldType = "string" // default to string sanitization
		}

		switch strings.ToLower(fieldType) {
		case "email":
			sanitized[key] = s.TrimAndSanitizeEmail(value)
		case "url":
			sanitized[key] = s.URL(value)
		case "path":
			sanitized[key] = s.Path(value)
		case "html":
			sanitized[key] = s.HTML(value)
		case "alpha":
			sanitized[key] = s.Alpha(value, false)
		case "alphanumeric":
			sanitized[key] = s.AlphaNumeric(value, false)
		case "numeric":
			sanitized[key] = s.Numeric(value)
		case "singleline":
			sanitized[key] = s.SingleLine(value)
		default:
			sanitized[key] = s.TrimAndSanitize(value)
		}
	}

	return sanitized
}

// SanitizeJSON sanitizes JSON data recursively
func (s *Service) SanitizeJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		return s.String(v)
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, value := range v {
			sanitized[key] = s.SanitizeJSON(value)
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, value := range v {
			sanitized[i] = s.SanitizeJSON(value)
		}
		return sanitized
	default:
		return v
	}
}

// SanitizeWithOptions provides advanced sanitization with options
type SanitizeOptions struct {
	TrimWhitespace bool
	RemoveHTML     bool
	MaxLength      int
	AllowedTags    []string
}

// SanitizeWithOptions sanitizes a string with custom options
func (s *Service) SanitizeWithOptions(input string, opts SanitizeOptions) string {
	if opts.TrimWhitespace {
		input = strings.TrimSpace(input)
	}

	if opts.RemoveHTML {
		input = s.HTML(input)
	} else {
		input = s.String(input)
	}

	if opts.MaxLength > 0 && len(input) > opts.MaxLength {
		input = input[:opts.MaxLength]
	}

	return input
}

// IsValidEmail checks if an email is valid after sanitization
func (s *Service) IsValidEmail(input string) bool {
	sanitized := s.Email(input)
	return sanitized != "" && strings.Contains(sanitized, "@")
}

// IsValidURL checks if a URL is valid after sanitization
func (s *Service) IsValidURL(input string) bool {
	sanitized := s.URL(input)
	return sanitized != "" && (strings.HasPrefix(sanitized, "http://") || strings.HasPrefix(sanitized, "https://"))
}
