package user

import "github.com/goformx/goforms/internal/domain/entities"

type Signup struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type LoginResponse struct {
	User  *entities.User
	Token *TokenPair
}
