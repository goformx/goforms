package contactmock

import (
	"context"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/stretchr/testify/mock"
)

// MockStore is a mock implementation of the contact store
type MockStore struct {
	mock.Mock
}

// NewMockStore creates a new mock store
func NewMockStore() *MockStore {
	return &MockStore{}
}

// Create creates a new contact submission
func (m *MockStore) Create(ctx context.Context, submission *contact.Submission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

// Get gets a contact submission by ID
func (m *MockStore) Get(ctx context.Context, id int64) (*contact.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	submission, ok := args.Get(0).(*contact.Submission)
	if !ok {
		return nil, args.Error(1)
	}
	return submission, args.Error(1)
}

// List lists all contact submissions
func (m *MockStore) List(ctx context.Context) ([]contact.Submission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	submissions, ok := args.Get(0).([]contact.Submission)
	if !ok {
		return nil, args.Error(1)
	}
	return submissions, args.Error(1)
}

// UpdateStatus updates a submission's status
func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
