package repositories

import "github.com/goformx/goforms/internal/domain/entities"

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByEmail(email string) (*entities.User, error)
	FindByID(id uint) (*entities.User, error)
	Create(user *entities.User) error
	Update(user *entities.User) error
	Delete(id uint) error
}
