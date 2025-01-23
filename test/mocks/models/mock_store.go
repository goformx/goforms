package models

import (
	"github.com/jonesrussell/goforms/internal/handlers"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

// Verify interface compliance at compile time
var _ SubscriptionStore = (*MockSubscriptionStore)(nil)
var _ handlers.PingContexter = (*MockPingContexter)(nil)

// MockSubscriptionStore is a mock implementation of SubscriptionStore interface
type MockSubscriptionStore struct {
	mock.Mock
}

// NewMockSubscriptionStore creates a new mock subscription store
func NewMockSubscriptionStore() *MockSubscriptionStore {
	return &MockSubscriptionStore{}
}

// Create mocks the Create method of the SubscriptionStore interface.
// It records the subscription and returns the configured error.
func (m *MockSubscriptionStore) Create(subscription *Subscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

// List mocks the List method of the SubscriptionStore interface.
// It returns the configured list of subscriptions and error.
func (m *MockSubscriptionStore) List() ([]Subscription, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Subscription), args.Error(1)
}

// GetByID mocks the GetByID method of the SubscriptionStore interface.
// It returns the configured subscription and error based on the provided ID.
func (m *MockSubscriptionStore) GetByID(id int64) (*Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Subscription), args.Error(1)
}

// GetByEmail mocks the GetByEmail method of the SubscriptionStore interface.
// It returns the configured subscription and error based on the provided email.
func (m *MockSubscriptionStore) GetByEmail(email string) (*Subscription, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Subscription), args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method of the SubscriptionStore interface.
// It records the status update and returns the configured error.
func (m *MockSubscriptionStore) UpdateStatus(id int64, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

// Delete mocks the Delete method of the SubscriptionStore interface.
// It records the deletion and returns the configured error.
func (m *MockSubscriptionStore) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockPingContexter is a mock implementation of PingContexter interface
type MockPingContexter struct {
	mock.Mock
}

// NewMockPingContexter creates a new mock ping contexter
func NewMockPingContexter() *MockPingContexter {
	return &MockPingContexter{}
}

// PingContext mocks the PingContext method of the PingContexter interface.
// It records the context and returns the configured error.
func (m *MockPingContexter) PingContext(c echo.Context) error {
	args := m.Called(c)
	return args.Error(0)
}
