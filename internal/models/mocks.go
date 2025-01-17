package models

import (
	"context"
)

// MockSubscriptionStore implements SubscriptionStore
type MockSubscriptionStore struct{}

func (m *MockSubscriptionStore) CreateSubscription(ctx context.Context, sub *Subscription) error {
	return nil
}

func (m *MockSubscriptionStore) GetSubscription(ctx context.Context, email string) (*Subscription, error) {
	return &Subscription{Email: email}, nil
}

// MockContactStore implements ContactStore
type MockContactStore struct{}

func (m *MockContactStore) CreateContact(ctx context.Context, submission *ContactSubmission) error {
	return nil
}

func (m *MockContactStore) GetContacts(ctx context.Context) ([]ContactSubmission, error) {
	return []ContactSubmission{}, nil
}

// MockPingContexter implements handlers.PingContexter
type MockPingContexter struct{}

func (m *MockPingContexter) PingContext(ctx context.Context) error {
	return nil
}
