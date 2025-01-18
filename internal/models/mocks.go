package models

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

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
