package subscriptionmock

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

const (
	expectedReturnValues = 2
)

var (
	ErrNotFound = errors.New("subscription not found")
)

// Ensure MockStore implements Store interface
var _ subscription.Store = (*MockStore)(nil)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	mock.Mock
	t        *testing.T
	mu       sync.Mutex
	calls    []mockCall
	expected []mockCall
}

// mockCall represents a single method call
type mockCall struct {
	method string
	args   []any
	ret    []any
}

// NewMockStore creates a new instance of MockStore
func NewMockStore(t *testing.T) *MockStore {
	return &MockStore{t: t}
}

// recordCall records a method call
func (m *MockStore) recordCall(method string, args []any) []any {
	m.mu.Lock()
	defer m.mu.Unlock()

	call := mockCall{method: method, args: args}
	m.calls = append(m.calls, call)

	// Find matching expectation
	for _, exp := range m.expected {
		if exp.method == method && matchArgs(exp.args, args) {
			return exp.ret
		}
	}
	return nil
}

// matchArgs compares two argument slices
func matchArgs(exp, got []any) bool {
	if len(exp) != len(got) {
		return false
	}
	for i := range exp {
		// For context, just check if both are contexts
		if _, expIsCtx := exp[i].(context.Context); expIsCtx {
			_, gotIsCtx := got[i].(context.Context)
			return gotIsCtx
		}
		if exp[i] != got[i] {
			return false
		}
	}
	return true
}

// ExpectCreate sets up an expectation for Create method
func (m *MockStore) ExpectCreate(
	ctx context.Context,
	sub, ret *subscription.Subscription,
	err error,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "Create",
		args:   []any{ctx, sub},
		ret:    []any{ret, err},
	})
}

// ExpectList sets up an expectation for List method
func (m *MockStore) ExpectList(ctx context.Context, ret []subscription.Subscription, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "List",
		args:   []any{ctx},
		ret:    []any{ret, err},
	})
}

// ExpectGetByID sets up an expectation for GetByID method
func (m *MockStore) ExpectGetByID(ctx context.Context, id int64, ret *subscription.Subscription, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "GetByID",
		args:   []any{ctx, id},
		ret:    []any{ret, err},
	})
}

// ExpectGetByEmail sets up an expectation for GetByEmail method
func (m *MockStore) ExpectGetByEmail(ctx context.Context, email string, ret *subscription.Subscription, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "GetByEmail",
		args:   []any{ctx, email},
		ret:    []any{ret, err},
	})
}

// ExpectUpdateStatus sets up an expectation for UpdateStatus method
func (m *MockStore) ExpectUpdateStatus(ctx context.Context, id int64, status subscription.Status, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "UpdateStatus",
		args:   []any{ctx, id, status},
		ret:    []any{err},
	})
}

// ExpectDelete sets up an expectation for Delete method
func (m *MockStore) ExpectDelete(ctx context.Context, id int64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "Delete",
		args:   []any{ctx, id},
		ret:    []any{err},
	})
}

// Create mocks the Create method
func (m *MockStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	ret := m.recordCall("Create", []any{ctx, sub})
	if len(ret) == 0 || ret[0] == nil {
		return nil
	}
	if err, ok := ret[0].(error); ok {
		return err
	}
	return errors.New("invalid error type returned from mock")
}

// List mocks the List method
func (m *MockStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	subs, ok := args.Get(0).([]subscription.Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion for subscriptions")
	}
	return subs, args.Error(1)
}

// GetByID mocks the GetByID method
func (m *MockStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, ErrNotFound
	}
	sub, ok := args.Get(0).(*subscription.Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion for subscription")
	}
	return sub, args.Error(1)
}

// GetByEmail mocks the GetByEmail method
func (m *MockStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	ret := m.recordCall("GetByEmail", []any{ctx, email})
	if len(ret) < expectedReturnValues {
		return nil, errors.New("no return values from mock")
	}

	var sub *subscription.Subscription
	if ret[0] != nil {
		if s, ok := ret[0].(*subscription.Subscription); ok {
			sub = s
		} else {
			return nil, errors.New("invalid subscription type returned from mock")
		}
	}

	var err error
	if ret[1] != nil {
		if e, ok := ret[1].(error); ok {
			err = e
		} else {
			return nil, errors.New("invalid error type returned from mock")
		}
	}
	return sub, err
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	ret := m.recordCall("UpdateStatus", []any{ctx, id, status})
	if len(ret) == 0 || ret[0] == nil {
		return nil
	}
	if err, ok := ret[0].(error); ok {
		return err
	}
	return errors.New("invalid error type returned from mock")
}

// Delete mocks the Delete method
func (m *MockStore) Delete(ctx context.Context, id int64) error {
	ret := m.recordCall("Delete", []any{ctx, id})
	if len(ret) == 0 || ret[0] == nil {
		return nil
	}
	if err, ok := ret[0].(error); ok {
		return err
	}
	return errors.New("invalid error type returned from mock")
}

// Verify checks if all expectations were met
func (m *MockStore) Verify() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.calls) != len(m.expected) {
		return fmt.Errorf("expected %d calls but got %d", len(m.expected), len(m.calls))
	}

	for i, exp := range m.expected {
		got := m.calls[i]
		if exp.method != got.method {
			return fmt.Errorf("call %d: expected method %q but got %q", i, exp.method, got.method)
		}
		if !matchArgs(exp.args, got.args) {
			return fmt.Errorf("call %d: arguments do not match", i)
		}
	}

	return nil
}

// Reset clears all calls and expectations
func (m *MockStore) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = m.calls[:0]
	m.expected = m.expected[:0]
}

// Get implements subscription.Store
func (m *MockStore) Get(ctx context.Context, id int64) (*subscription.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	sub, ok := args.Get(0).(*subscription.Subscription)
	if !ok {
		return nil, errors.New("invalid type assertion for subscription")
	}
	return sub, args.Error(1)
}
