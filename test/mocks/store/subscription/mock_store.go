package subscriptionmock

import (
	"context"
	"fmt"
	"sync"

	"github.com/jonesrussell/goforms/internal/domain/subscription"
)

// Ensure MockStore implements Store interface
var _ subscription.Store = (*MockStore)(nil)

// MockStore is a mock implementation of the Store interface
type MockStore struct {
	mu       sync.Mutex
	calls    []mockCall
	expected []mockCall
}

// mockCall represents a single method call
type mockCall struct {
	method string
	args   []interface{}
	ret    []interface{}
}

// NewMockStore creates a new instance of MockStore
func NewMockStore() *MockStore {
	return &MockStore{}
}

// recordCall records a method call
func (m *MockStore) recordCall(method string, args []interface{}) []interface{} {
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
func matchArgs(exp, got []interface{}) bool {
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
func (m *MockStore) ExpectCreate(ctx context.Context, sub *subscription.Subscription, ret error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "Create",
		args:   []interface{}{ctx, sub},
		ret:    []interface{}{ret},
	})
}

// ExpectList sets up an expectation for List method
func (m *MockStore) ExpectList(ctx context.Context, ret []subscription.Subscription, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "List",
		args:   []interface{}{ctx},
		ret:    []interface{}{ret, err},
	})
}

// ExpectGetByID sets up an expectation for GetByID method
func (m *MockStore) ExpectGetByID(ctx context.Context, id int64, ret *subscription.Subscription, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "GetByID",
		args:   []interface{}{ctx, id},
		ret:    []interface{}{ret, err},
	})
}

// ExpectGetByEmail sets up an expectation for GetByEmail method
func (m *MockStore) ExpectGetByEmail(ctx context.Context, email string, ret *subscription.Subscription, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "GetByEmail",
		args:   []interface{}{ctx, email},
		ret:    []interface{}{ret, err},
	})
}

// ExpectUpdateStatus sets up an expectation for UpdateStatus method
func (m *MockStore) ExpectUpdateStatus(ctx context.Context, id int64, status subscription.Status, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "UpdateStatus",
		args:   []interface{}{ctx, id, status},
		ret:    []interface{}{err},
	})
}

// ExpectDelete sets up an expectation for Delete method
func (m *MockStore) ExpectDelete(ctx context.Context, id int64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "Delete",
		args:   []interface{}{ctx, id},
		ret:    []interface{}{err},
	})
}

// Create mocks the Create method
func (m *MockStore) Create(ctx context.Context, sub *subscription.Subscription) error {
	ret := m.recordCall("Create", []interface{}{ctx, sub})
	if ret == nil || ret[0] == nil {
		return nil
	}
	return ret[0].(error)
}

// List mocks the List method
func (m *MockStore) List(ctx context.Context) ([]subscription.Subscription, error) {
	ret := m.recordCall("List", []interface{}{ctx})
	if ret == nil {
		return nil, nil
	}
	var subs []subscription.Subscription
	var err error
	if ret[0] != nil {
		subs = ret[0].([]subscription.Subscription)
	}
	if ret[1] != nil {
		err = ret[1].(error)
	}
	return subs, err
}

// GetByID mocks the GetByID method
func (m *MockStore) GetByID(ctx context.Context, id int64) (*subscription.Subscription, error) {
	ret := m.recordCall("GetByID", []interface{}{ctx, id})
	if ret == nil {
		return nil, nil
	}
	var sub *subscription.Subscription
	var err error
	if ret[0] != nil {
		sub = ret[0].(*subscription.Subscription)
	}
	if ret[1] != nil {
		err = ret[1].(error)
	}
	return sub, err
}

// GetByEmail mocks the GetByEmail method
func (m *MockStore) GetByEmail(ctx context.Context, email string) (*subscription.Subscription, error) {
	ret := m.recordCall("GetByEmail", []interface{}{ctx, email})
	if ret == nil {
		return nil, nil
	}
	var sub *subscription.Subscription
	var err error
	if ret[0] != nil {
		sub = ret[0].(*subscription.Subscription)
	}
	if ret[1] != nil {
		err = ret[1].(error)
	}
	return sub, err
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status subscription.Status) error {
	ret := m.recordCall("UpdateStatus", []interface{}{ctx, id, status})
	if ret == nil || ret[0] == nil {
		return nil
	}
	return ret[0].(error)
}

// Delete mocks the Delete method
func (m *MockStore) Delete(ctx context.Context, id int64) error {
	ret := m.recordCall("Delete", []interface{}{ctx, id})
	if ret == nil || ret[0] == nil {
		return nil
	}
	return ret[0].(error)
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
			return fmt.Errorf("call %d: expected method %s but got %s", i, exp.method, got.method)
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
