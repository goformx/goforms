package contactmock

import (
	"context"
	"errors"

	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/contact"
)

// Ensure MockStore implements contact.Store interface
var _ contact.Store = (*MockStore)(nil)

// MockStore is a mock implementation of contact.Store
type MockStore struct {
	mock.Mock
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
	subs, ok := args.Get(0).([]contact.Submission)
	if !ok {
		return nil, errors.New("invalid return type for List")
	}
	return subs, args.Error(1)
}

// Get mocks the Get method
func (m *MockStore) Get(ctx context.Context, id int64) (*contact.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sub, ok := args.Get(0).(*contact.Submission)
	if !ok {
		return nil, errors.New("invalid return type for Get")
	}
	return sub, args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// NewMockStore creates a new mock store
func NewMockStore() *MockStore {
	return &MockStore{}
}
