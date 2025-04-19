package contactmock

import (
	"context"
	"errors"
	"sync"

	"github.com/jonesrussell/goforms/internal/domain/contact"
	"github.com/stretchr/testify/mock"
)

var (
	ErrNoReturnValues = errors.New("no return values from mock")
	ErrNotFound       = errors.New("contact not found")
)

// Ensure MockService implements contact.Service interface
var _ contact.Service = (*MockService)(nil)

// MockService is a mock implementation of the contact service
type MockService struct {
	mock.Mock
	mu sync.Mutex
}

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{}
}

// Submit submits a contact form
func (m *MockService) Submit(ctx context.Context, submission *contact.Submission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

// ListSubmissions lists all contact submissions
func (m *MockService) ListSubmissions(ctx context.Context) ([]contact.Submission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]contact.Submission), args.Error(1)
}

// GetSubmission gets a contact submission by ID
func (m *MockService) GetSubmission(ctx context.Context, id int64) (*contact.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Submission), args.Error(1)
}

// UpdateSubmissionStatus updates a submission's status
func (m *MockService) UpdateSubmissionStatus(ctx context.Context, id int64, status contact.Status) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// Verify checks if all expected calls were made
func (m *MockService) Verify() error {
	m.Mock.AssertExpectations(mock.TestingT(nil))
	return nil
}

// Reset resets the mock
func (m *MockService) Reset() {
	m.Mock = mock.Mock{}
}

func (m *MockService) GetByID(ctx context.Context, id string) (*contact.Submission, error) {
	ret := m.Called(ctx, id)
	if len(ret) == 0 {
		return nil, ErrNoReturnValues
	}
	return ret[0].(*contact.Submission), ret.Error(1)
}

func (m *MockService) List(ctx context.Context) ([]*contact.Submission, error) {
	ret := m.Called(ctx)
	if len(ret) == 0 {
		return nil, ErrNoReturnValues
	}
	return ret[0].([]*contact.Submission), ret.Error(1)
}

func (m *MockService) Create(ctx context.Context, submission *contact.Submission) error {
	ret := m.Called(ctx, submission)
	if len(ret) == 0 {
		return ErrNoReturnValues
	}
	return ret.Error(0)
}
