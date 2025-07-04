package api

// SubmitFormRequest represents the expected fields for submitting a form
// (expand as needed for validation, etc.)
type SubmitFormRequest struct {
	Data map[string]any `json:"data"`
}

// UpdateSchemaRequest represents the expected fields for updating a form schema
type UpdateSchemaRequest struct {
	Schema map[string]any `json:"schema"`
}

// APIResponse is a generic response for API endpoints
type APIResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
