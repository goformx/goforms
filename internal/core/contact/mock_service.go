package contact

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// Ensure MockService implements Service interface
var _ Service = (*MockService)(nil)

// MockService is a mock implementation of the Service interface
type MockService struct {
	mock.Mock
}

// NewMockService creates a new instance of MockService
func NewMockService() *MockService {
	return &MockService{}
}

// CreateSubmission mocks the CreateSubmission method
func (m *MockService) CreateSubmission(ctx context.Context, submission *Submission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

// ListSubmissions mocks the ListSubmissions method
func (m *MockService) ListSubmissions(ctx context.Context) ([]Submission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Submission), args.Error(1)
}

// GetSubmission mocks the GetSubmission method
func (m *MockService) GetSubmission(ctx context.Context, id int64) (*Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Submission), args.Error(1)
}

// UpdateSubmissionStatus mocks the UpdateSubmissionStatus method
func (m *MockService) UpdateSubmissionStatus(ctx context.Context, id int64, status Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
