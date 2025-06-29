// Package sanitization provides utilities for cleaning and validating user input
// to prevent XSS attacks, injection attacks, and other security vulnerabilities.
package sanitization

import (
	"fmt"
	"html"
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
	domain, err := sanitize.Domain(input, false, false)
	if err != nil {
		return "", fmt.Errorf("sanitize domain: %w", err)
	}

	return domain, nil
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
	trimmed := strings.TrimSpace(input)

	return s.HTML(trimmed)
}

// TrimAndSanitizeEmail trims whitespace and sanitizes an email
func (s *Service) TrimAndSanitizeEmail(input string) string {
	return s.Email(strings.TrimSpace(input))
}

// SanitizeForLogging sanitizes a string specifically for safe logging
// This method prevents log injection attacks by removing newlines, null bytes,
// HTML tags, and HTML escaping the content
func (s *Service) SanitizeForLogging(input string) string {
	if input == "" {
		return ""
	}

	// First, remove any newline characters that could be used for log injection
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.ReplaceAll(input, "\r\n", " ")

	// Remove any null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Use the sanitization service to clean the string (removes HTML tags, etc.)
	input = s.SingleLine(input)

	// HTML escape the string to prevent HTML injection if logs are displayed in HTML
	input = html.EscapeString(input)

	// Trim any extra whitespace that might have been introduced
	input = strings.TrimSpace(input)

	return input
}

// SanitizeMap sanitizes a map of string keys to any values
func (s *Service) SanitizeMap(data map[string]any) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = s.TrimAndSanitize(v)
		case map[string]any:
			s.SanitizeMap(v)
		case []any:
			s.SanitizeSlice(v)
		}
	}
}

// SanitizeSlice sanitizes a slice of any values
func (s *Service) SanitizeSlice(data []any) {
	for i, value := range data {
		switch v := value.(type) {
		case string:
			data[i] = s.TrimAndSanitize(v)
		case map[string]any:
			s.SanitizeMap(v)
		case []any:
			s.SanitizeSlice(v)
		}
	}
}

// sanitizeStructField handles sanitization of individual struct fields
// Skip other types (bool, int, float, etc.)
func (s *Service) sanitizeStructField(field reflect.Value) {
	switch field.Kind() {
	case reflect.String:
		if field.CanSet() {
			field.SetString(s.String(field.String()))
		}
	case reflect.Struct:
		s.SanitizeStruct(field.Interface())
	case reflect.Slice:
		s.sanitizeSliceField(field)
	case reflect.Map:
		s.sanitizeMapField(field)
	case reflect.Ptr:
		if !field.IsNil() {
			s.sanitizeStructField(field.Elem())
		}
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func,
		reflect.Interface, reflect.UnsafePointer:
	}
}

// sanitizeSliceField handles sanitization of slice fields
func (s *Service) sanitizeSliceField(field reflect.Value) {
	if field.Type().Elem().Kind() == reflect.String {
		for i := range field.Len() {
			if field.Index(i).CanSet() {
				field.Index(i).SetString(s.String(field.Index(i).String()))
			}
		}
	}
}

// sanitizeMapField handles sanitization of map fields
func (s *Service) sanitizeMapField(field reflect.Value) {
	if field.Type().Key().Kind() == reflect.String {
		iter := field.MapRange()
		for iter.Next() {
			key := iter.Key()
			value := iter.Value()

			if value.Kind() == reflect.String {
				field.SetMapIndex(key, reflect.ValueOf(s.String(value.String())))
			}
		}
	}
}

// SanitizeStruct sanitizes all string fields in a struct
func (s *Service) SanitizeStruct(obj any) {
	if obj == nil {
		return
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := range v.NumField() {
		field := v.Field(i)
		s.sanitizeStructField(field)
	}
}

// SanitizeFormData sanitizes form data based on field types
func (s *Service) SanitizeFormData(data, fieldTypes map[string]string) map[string]string {
	result := make(map[string]string)

	for key, value := range data {
		fieldType, exists := fieldTypes[key]
		if !exists {
			fieldType = "string" // default to string
		}

		switch fieldType {
		case "email":
			result[key] = s.TrimAndSanitizeEmail(value)
		case "url":
			result[key] = s.URL(strings.TrimSpace(value))
		case "path":
			result[key] = s.Path(strings.TrimSpace(value))
		case "html":
			result[key] = s.HTML(strings.TrimSpace(value))
		case "numeric":
			result[key] = s.Numeric(strings.TrimSpace(value))
		default:
			result[key] = s.TrimAndSanitize(value)
		}
	}

	return result
}

// SanitizeJSON sanitizes JSON data recursively
func (s *Service) SanitizeJSON(data any) any {
	switch v := data.(type) {
	case string:
		return s.TrimAndSanitize(v)
	case map[string]any:
		sanitized := make(map[string]any)
		for key, value := range v {
			sanitized[key] = s.SanitizeJSON(value)
		}

		return sanitized
	case []any:
		sanitized := make([]any, len(v))
		for i, value := range v {
			sanitized[i] = s.SanitizeJSON(value)
		}

		return sanitized
	default:
		return data
	}
}

// SanitizeOptions provides advanced sanitization with options
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
