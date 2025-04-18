package usermock

import (
	"sync"

	"github.com/jonesrussell/goforms/internal/domain/user"
)

// UserStore is a mock implementation of user.Store interface
type UserStore struct {
	users     map[uint]*user.User
	emailMap  map[string]uint
	mu        sync.RWMutex
	nextID    uint
	createErr error
	getErr    error
	updateErr error
	deleteErr error
	listErr   error
}

var _ user.Store = (*UserStore)(nil) // Ensure UserStore implements user.Store

// NewUserStore creates a new mock user store
func NewUserStore() *UserStore {
	return &UserStore{
		users:    make(map[uint]*user.User),
		emailMap: make(map[string]uint),
		nextID:   1,
	}
}

// SetError sets the error to be returned by the specified operation
func (m *UserStore) SetError(op string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch op {
	case "create":
		m.createErr = err
	case "get":
		m.getErr = err
	case "update":
		m.updateErr = err
	case "delete":
		m.deleteErr = err
	case "list":
		m.listErr = err
	}
}

// Create implements Store.Create
func (m *UserStore) Create(u *user.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createErr != nil {
		return m.createErr
	}

	if _, exists := m.emailMap[u.Email]; exists {
		return user.ErrEmailAlreadyExists
	}

	u.ID = m.nextID
	m.nextID++

	// Create deep copy to prevent external modifications
	userCopy := *u
	m.users[u.ID] = &userCopy
	m.emailMap[u.Email] = u.ID

	return nil
}

// GetByID implements Store.GetByID
func (m *UserStore) GetByID(id uint) (*user.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getErr != nil {
		return nil, m.getErr
	}

	u, exists := m.users[id]
	if !exists {
		return nil, user.ErrUserNotFound
	}

	// Return copy to prevent external modifications
	userCopy := *u
	return &userCopy, nil
}

// GetByEmail implements Store.GetByEmail
func (m *UserStore) GetByEmail(email string) (*user.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getErr != nil {
		return nil, m.getErr
	}

	id, exists := m.emailMap[email]
	if !exists {
		return nil, nil
	}

	u := m.users[id]
	// Return copy to prevent external modifications
	userCopy := *u
	return &userCopy, nil
}

// Update implements Store.Update
func (m *UserStore) Update(u *user.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateErr != nil {
		return m.updateErr
	}

	if _, exists := m.users[u.ID]; !exists {
		return user.ErrUserNotFound
	}

	// If email is being changed, update email map
	oldUser := m.users[u.ID]
	if oldUser.Email != u.Email {
		if _, exists := m.emailMap[u.Email]; exists {
			return user.ErrEmailAlreadyExists
		}
		delete(m.emailMap, oldUser.Email)
		m.emailMap[u.Email] = u.ID
	}

	// Create deep copy to prevent external modifications
	userCopy := *u
	m.users[u.ID] = &userCopy

	return nil
}

// Delete implements Store.Delete
func (m *UserStore) Delete(id uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleteErr != nil {
		return m.deleteErr
	}

	u, exists := m.users[id]
	if !exists {
		return user.ErrUserNotFound
	}

	delete(m.users, id)
	delete(m.emailMap, u.Email)

	return nil
}

// List implements Store.List
func (m *UserStore) List() ([]user.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.listErr != nil {
		return nil, m.listErr
	}

	users := make([]user.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, *u)
	}

	return users, nil
}
