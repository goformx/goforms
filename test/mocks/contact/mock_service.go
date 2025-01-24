package contactmock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/contact"
)

// Ensure MockService implements contact.Service interface
var _ contact.Service = (*MockService)(nil)

// MockService is a mock implementation of the contact service
type MockService struct {
	mock.Mock
}

// Submit mocks the Submit method
func (m *MockService) Submit(ctx context.Context, sub *contact.Submission) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

// ListSubmissions mocks the ListSubmissions method
func (m *MockService) ListSubmissions(ctx context.Context) ([]contact.Submission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]contact.Submission), args.Error(1)
}

// GetSubmission mocks the GetSubmission method
func (m *MockService) GetSubmission(ctx context.Context, id int64) (*contact.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Submission), args.Error(1)
}

// UpdateSubmissionStatus mocks the UpdateSubmissionStatus method
func (m *MockService) UpdateSubmissionStatus(ctx context.Context, id int64, status contact.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{}
}
