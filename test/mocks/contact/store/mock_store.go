package contactmock

import (
	"context"

	"github.com/jonesrussell/goforms/internal/domain/contact"
)

// MockStore is a mock implementation of contact.Store
type MockStore struct {
	CreateFunc       func(ctx context.Context, sub *contact.Submission) error
	ListFunc         func(ctx context.Context) ([]contact.Submission, error)
	GetFunc          func(ctx context.Context, id int64) (*contact.Submission, error)
	UpdateStatusFunc func(ctx context.Context, id int64, status contact.Status) error
}

// NewMockStore creates a new mock store
func NewMockStore() *MockStore {
	return &MockStore{
		CreateFunc: func(ctx context.Context, sub *contact.Submission) error {
			return nil
		},
		ListFunc: func(ctx context.Context) ([]contact.Submission, error) {
			return nil, nil
		},
		GetFunc: func(ctx context.Context, id int64) (*contact.Submission, error) {
			return nil, nil
		},
		UpdateStatusFunc: func(ctx context.Context, id int64, status contact.Status) error {
			return nil
		},
	}
}

// Create implements contact.Store
func (m *MockStore) Create(ctx context.Context, sub *contact.Submission) error {
	return m.CreateFunc(ctx, sub)
}

// List implements contact.Store
func (m *MockStore) List(ctx context.Context) ([]contact.Submission, error) {
	return m.ListFunc(ctx)
}

// Get implements contact.Store
func (m *MockStore) Get(ctx context.Context, id int64) (*contact.Submission, error) {
	return m.GetFunc(ctx, id)
}

// UpdateStatus implements contact.Store
func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	return m.UpdateStatusFunc(ctx, id, status)
}
