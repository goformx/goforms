package user

import "github.com/goformx/goforms/internal/domain/entities"

type Signup struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type Login struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type LoginResponse struct {
	User  *entities.User
	Token *TokenPair
}
