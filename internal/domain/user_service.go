package domain

import (
	"context"

	"github.com/goformx/goforms/internal/domain/user"
)

// UserService defines the interface for user operations
type UserService interface {
	SignUp(ctx context.Context, signup *user.Signup) (*user.User, error)
	Login(ctx context.Context, login *user.Login) (*user.LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetUserByID(ctx context.Context, id uint) (*user.User, error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	UpdateUser(ctx context.Context, user *user.User) error
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context) ([]user.User, error)
	GetByID(ctx context.Context, id string) (*user.User, error)
	ValidateToken(ctx context.Context, token string) error
	GetUserIDFromToken(ctx context.Context, token string) (uint, error)
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	Authenticate(ctx context.Context, email, password string) (*user.User, error)
}
