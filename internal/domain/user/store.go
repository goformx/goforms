package user

import (
	"context"
)

// Repository defines the interface for user storage
// (renamed from Store for consistency)
type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]User, error)
	GetByIDString(ctx context.Context, id string) (*User, error)
}
