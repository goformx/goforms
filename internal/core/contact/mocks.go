package contact

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// Verify interface compliance at compile time
var _ Store = (*MockStore)(nil)

// MockStore is a mock implementation of Store interface
type MockStore struct {
	mock.Mock
}

// NewMockStore creates a new mock store
func NewMockStore() *MockStore {
	return &MockStore{}
}

// Create mocks the Create method of the Store interface.
// It records the submission and returns the configured error.
func (m *MockStore) Create(ctx context.Context, submission *Submission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

// List mocks the List method of the Store interface.
// It returns the configured list of submissions and error.
func (m *MockStore) List(ctx context.Context) ([]Submission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Submission), args.Error(1)
}

// GetByID mocks the GetByID method of the Store interface.
// It returns the configured submission and error based on the provided ID.
func (m *MockStore) GetByID(ctx context.Context, id int64) (*Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Submission), args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method of the Store interface.
// It records the status update and returns the configured error.
func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
