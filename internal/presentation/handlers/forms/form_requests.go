package forms

// CreateFormRequest represents the expected fields for creating a form
type CreateFormRequest struct {
	Title       string `json:"title" form:"title"`
	Description string `json:"description" form:"description"`
}

// UpdateFormRequest represents the expected fields for updating a form
type UpdateFormRequest struct {
	Title       string `json:"title" form:"title"`
	Description string `json:"description" form:"description"`
}

// FormResponse is a generic response for form endpoints
type FormResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
