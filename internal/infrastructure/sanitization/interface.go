// Package sanitization provides utilities for cleaning and validating user input
// to prevent XSS attacks, injection attacks, and other security vulnerabilities.
//
//go:generate mockgen -typed -source=interface.go -destination=../../../test/mocks/sanitization/mock_service.go -package=sanitization -mock_names=ServiceInterface=MockService
package sanitization

// ServiceInterface defines the interface for sanitization operations
type ServiceInterface interface {
	// Basic string sanitization methods
	String(input string) string
	Email(input string) string
	URL(input string) string
	HTML(input string) string
	Path(input string) string
	IPAddress(input string) string
	Domain(input string) (string, error)
	URI(input string) string
	Alpha(input string, spaces bool) string
	AlphaNumeric(input string, spaces bool) string
	Numeric(input string) string
	SingleLine(input string) string
	Scripts(input string) string
	XML(input string) string
	TrimAndSanitize(input string) string
	TrimAndSanitizeEmail(input string) string

	// Log-specific sanitization
	SanitizeForLogging(input string) string

	// Complex data structure sanitization
	SanitizeMap(data map[string]any)
	SanitizeSlice(data []any)
	SanitizeStruct(obj any)
	SanitizeFormData(data map[string]string, fieldTypes map[string]string) map[string]string
	SanitizeJSON(data any) any
	SanitizeWithOptions(input string, opts SanitizeOptions) string

	// Validation methods (removed: IsValidEmail, IsValidURL)
}

// Ensure Service implements ServiceInterface
var _ ServiceInterface = (*Service)(nil)
