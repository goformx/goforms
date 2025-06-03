package user

import (
	"context"
	"strconv"
)

// Store defines the interface for user storage
type Store interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]User, error)
	GetByIDString(ctx context.Context, id string) (*User, error)
}

// StoreImpl implements the Store interface
type StoreImpl struct {
	// TODO: Add store implementation
}

// Create creates a new user
func (s *StoreImpl) Create(ctx context.Context, user *User) error {
	// TODO: Implement
	return nil
}

// GetByID retrieves a user by ID
func (s *StoreImpl) GetByID(ctx context.Context, id uint) (*User, error) {
	// TODO: Implement
	return nil, ErrUserNotFound
}

// GetByEmail retrieves a user by email
func (s *StoreImpl) GetByEmail(ctx context.Context, email string) (*User, error) {
	// TODO: Implement
	return nil, nil
}

// Update updates a user
func (s *StoreImpl) Update(ctx context.Context, user *User) error {
	// TODO: Implement
	return nil
}

// Delete deletes a user
func (s *StoreImpl) Delete(ctx context.Context, id uint) error {
	// TODO: Implement
	return nil
}

// List lists all users
func (s *StoreImpl) List(ctx context.Context) ([]User, error) {
	// TODO: Implement
	return nil, nil
}

// GetByIDString retrieves a user by ID string
func (s *StoreImpl) GetByIDString(ctx context.Context, id string) (*User, error) {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, uint(userID))
}
