package sanitization

// ServiceInterface defines the interface for sanitization operations
type ServiceInterface interface {
	// Basic sanitization methods
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

	// Convenience methods
	TrimAndSanitize(input string) string
	TrimAndSanitizeEmail(input string) string

	// Complex data sanitization
	SanitizeMap(data map[string]interface{})
	SanitizeSlice(data []interface{})
	SanitizeStruct(obj interface{})
	SanitizeFormData(data map[string]string, fieldTypes map[string]string) map[string]string
	SanitizeJSON(data interface{}) interface{}

	// Advanced sanitization
	SanitizeWithOptions(input string, opts SanitizeOptions) string

	// Validation methods
	IsValidEmail(input string) bool
	IsValidURL(input string) bool
}

// Ensure Service implements ServiceInterface
var _ ServiceInterface = (*Service)(nil)
