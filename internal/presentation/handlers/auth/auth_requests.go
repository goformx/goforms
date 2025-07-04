package auth

// LoginRequest represents the expected fields for a login POST
// (expand as needed for validation, etc.)
type LoginRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

// SignupRequest represents the expected fields for a signup POST
type SignupRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
	Name     string `json:"name" form:"name"`
}

// AuthResponse is a generic response for auth endpoints
type AuthResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
