# Sanitization Service

## Overview

The sanitization service provides a standardized interface for sanitizing various types of input data using the `go-sanitize` library. This service centralizes all sanitization logic and provides a clean, testable interface for the entire application.

## Features

- **Comprehensive Coverage**: Supports all major input types (strings, emails, URLs, HTML, etc.)
- **Type Safety**: Uses interfaces for better testability and dependency injection
- **Consistent API**: Standardized methods across all sanitization operations
- **Advanced Features**: Support for complex data structures, custom options, and validation

## Usage

### Basic Sanitization

```go
// Inject the sanitization service
type MyHandler struct {
    Sanitizer sanitization.ServiceInterface
}

// Basic string sanitization
cleanText := sanitizer.String("<script>alert('xss')</script>Hello")
// Result: ">alert('xss')</Hello"

// Email sanitization
cleanEmail := sanitizer.Email("  user@example.com  ")
// Result: "user@example.com"

// URL sanitization
cleanURL := sanitizer.URL("https://example.com")
// Result: "https://example.com"
```

### Advanced Sanitization

```go
// Sanitize complex data structures
data := map[string]interface{}{
    "name": "<script>alert('xss')</script>John",
    "email": "john@example.com",
    "nested": map[string]interface{}{
        "comment": "<b>Bold</b> text",
    },
}
sanitizer.SanitizeMap(data)

// Sanitize form data with field types
formData := map[string]string{
    "name": "  John Doe  ",
    "email": "  user@example.com  ",
    "url": "  https://example.com  ",
}
fieldTypes := map[string]string{
    "name": "string",
    "email": "email",
    "url": "url",
}
cleanData := sanitizer.SanitizeFormData(formData, fieldTypes)

// Sanitize with custom options
opts := sanitization.SanitizeOptions{
    TrimWhitespace: true,
    RemoveHTML:     true,
    MaxLength:       100,
}
cleanText := sanitizer.SanitizeWithOptions(input, opts)
```

### Validation

```go
// Check if email is valid after sanitization
if sanitizer.IsValidEmail(input) {
    // Process valid email
}

// Check if URL is valid after sanitization
if sanitizer.IsValidURL(input) {
    // Process valid URL
}
```

## Available Methods

### Basic Sanitization
- `String(input string) string` - XSS protection
- `Email(input string) string` - Email sanitization
- `URL(input string) string` - URL sanitization
- `HTML(input string) string` - HTML tag removal
- `Path(input string) string` - File path sanitization
- `IPAddress(input string) string` - IP address sanitization
- `Domain(input string) (string, error)` - Domain sanitization
- `URI(input string) string` - URI sanitization

### Character Type Sanitization
- `Alpha(input string, spaces bool) string` - Alpha characters only
- `AlphaNumeric(input string, spaces bool) string` - Alphanumeric characters only
- `Numeric(input string) string` - Numeric characters only
- `SingleLine(input string) string` - Remove newlines and extra whitespace
- `Scripts(input string) string` - Remove script tags
- `XML(input string) string` - XML sanitization

### Convenience Methods
- `TrimAndSanitize(input string) string` - Trim whitespace and sanitize
- `TrimAndSanitizeEmail(input string) string` - Trim whitespace and sanitize email

### Complex Data Sanitization
- `SanitizeMap(data map[string]interface{})` - Sanitize map values recursively
- `SanitizeSlice(data []interface{})` - Sanitize slice values recursively
- `SanitizeStruct(obj interface{})` - Sanitize struct string fields
- `SanitizeFormData(data map[string]string, fieldTypes map[string]string) map[string]string` - Sanitize form data with field types
- `SanitizeJSON(data interface{}) interface{}` - Sanitize JSON data recursively

### Advanced Features
- `SanitizeWithOptions(input string, opts SanitizeOptions) string` - Sanitize with custom options
- `IsValidEmail(input string) bool` - Validate email after sanitization
- `IsValidURL(input string) bool` - Validate URL after sanitization

## Field Types for Form Data

When using `SanitizeFormData`, you can specify field types to apply appropriate sanitization:

- `"string"` - Default string sanitization (XSS protection)
- `"email"` - Email sanitization with trimming
- `"url"` - URL sanitization
- `"path"` - File path sanitization
- `"html"` - HTML tag removal
- `"alpha"` - Alpha characters only
- `"alphanumeric"` - Alphanumeric characters only
- `"numeric"` - Numeric characters only
- `"singleline"` - Remove newlines and extra whitespace

## SanitizeOptions

```go
type SanitizeOptions struct {
    TrimWhitespace bool     // Trim leading/trailing whitespace
    RemoveHTML     bool     // Remove HTML tags
    MaxLength      int      // Maximum length (0 = no limit)
    AllowedTags    []string // Allowed HTML tags (future use)
}
```

## Dependency Injection

The sanitization service is provided through dependency injection:

```go
// In your handler constructor
func NewMyHandler(sanitizer sanitization.ServiceInterface) *MyHandler {
    return &MyHandler{
        Sanitizer: sanitizer,
    }
}

// The service is automatically provided by the infrastructure module
```

## Testing

The service includes comprehensive tests and supports mocking for unit tests:

```go
// In your tests
func TestMyHandler(t *testing.T) {
    mockSanitizer := &MockSanitizationService{}
    handler := NewMyHandler(mockSanitizer)
    // Test implementation
}
```

## Best Practices

1. **Always sanitize user input** at the boundary (handlers/controllers)
2. **Use appropriate field types** for form data sanitization
3. **Validate after sanitization** when needed
4. **Log sanitization errors** for debugging
5. **Test sanitization logic** thoroughly
6. **Use dependency injection** for better testability

## Migration from Direct go-sanitize Usage

If you have existing code using `go-sanitize` directly:

```go
// Before
import "github.com/mrz1836/go-sanitize"
cleanText := sanitize.XSS(input)

// After
cleanText := sanitizer.String(input)
```

This provides better testability, consistency, and maintainability across the application. 