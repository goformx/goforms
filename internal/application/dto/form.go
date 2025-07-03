package dto

// FormCreateRequest represents the data needed to create a form
// Moved from handlers/web/form_interfaces.go
// Only contains struct definition, no handler-specific logic
type FormCreateRequest struct {
	Title string `json:"title"`
}

// FormUpdateRequest represents the data needed to update a form
// Moved from handlers/web/form_interfaces.go
type FormUpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CorsOrigins string `json:"cors_origins"`
}
