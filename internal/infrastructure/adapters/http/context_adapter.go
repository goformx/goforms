package http

// Set stores a value in the context - ensures it's available to both middleware and templates
func (e *EchoContextAdapter) Set(key string, value any) {
	e.Context.Set(key, value) // Store directly on Echo context for template access
}

// Get retrieves a value from the context - reads from Echo context
func (e *EchoContextAdapter) Get(key string) any {
	return e.Context.Get(key)
}
