package contact

import (
	"context"

	"github.com/jonesrussell/goforms/internal/core/contact"
	"github.com/stretchr/testify/mock"
)

// Ensure MockStore implements Store interface
var _ contact.Store = (*MockStore)(nil)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	mock.Mock
}

// NewMockStore creates a new instance of MockStore
func NewMockStore() *MockStore {
	return &MockStore{}
}

// Create mocks the Create method
func (m *MockStore) Create(ctx context.Context, sub *contact.Submission) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// List mocks the List method
func (m *MockStore) List(ctx context.Context) ([]contact.Submission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]contact.Submission), args.Error(1)
}

// GetByID mocks the GetByID method
func (m *MockStore) GetByID(ctx context.Context, id int64) (*contact.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Submission), args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
