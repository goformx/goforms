package validation

// Validator defines the interface for validation operations
type Validator interface {
	// Struct validates a struct based on validation tags
	Struct(interface{}) error
	// Var validates a single variable using a tag
	Var(interface{}, string) error
	// RegisterValidation adds a custom validation with the given tag
	RegisterValidation(string, func(interface{}) bool) error
}
