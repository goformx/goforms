package models

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

// MockContactStore is a mock implementation of ContactStore
type MockContactStore struct {
	mock.Mock
}

// NewMockContactStore creates a new mock contact store
func NewMockContactStore() *MockContactStore {
	return &MockContactStore{}
}

// CreateContact mocks the CreateContact method
func (m *MockContactStore) CreateContact(ctx context.Context, submission *ContactSubmission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

// GetContacts mocks the GetContacts method
func (m *MockContactStore) GetContacts(ctx context.Context) ([]ContactSubmission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ContactSubmission), args.Error(1)
}

// MockSubscriptionStore is a mock implementation of SubscriptionStore
type MockSubscriptionStore struct {
	mock.Mock
}

// NewMockSubscriptionStore creates a new mock subscription store
func NewMockSubscriptionStore() *MockSubscriptionStore {
	return &MockSubscriptionStore{}
}

// Create mocks the Create method
func (m *MockSubscriptionStore) Create(subscription *Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

// MockPingContexter is a mock implementation of PingContexter
type MockPingContexter struct {
	mock.Mock
}

// NewMockPingContexter creates a new mock ping contexter
func NewMockPingContexter() *MockPingContexter {
	return &MockPingContexter{}
}

// PingContext mocks the PingContext method
func (m *MockPingContexter) PingContext(c echo.Context) error {
	args := m.Called(c)
	return args.Error(0)
}
