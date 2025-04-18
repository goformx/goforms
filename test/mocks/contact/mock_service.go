package contactmock

import (
	"context"
	"fmt"
	"sync"

	"github.com/jonesrussell/goforms/internal/domain/contact"
)

// Ensure MockService implements contact.Service interface
var _ contact.Service = (*MockService)(nil)

// MockService is a mock implementation of the contact service
type MockService struct {
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

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{}
}

// recordCall records a method call
func (m *MockService) recordCall(method string, args []any) []any {
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

// ExpectSubmit sets up an expectation for Submit method
func (m *MockService) ExpectSubmit(ctx context.Context, sub *contact.Submission, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "Submit",
		args:   []any{ctx, sub},
		ret:    []any{err},
	})
}

// ExpectListSubmissions sets up an expectation for ListSubmissions method
func (m *MockService) ExpectListSubmissions(ctx context.Context, ret []contact.Submission, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "ListSubmissions",
		args:   []any{ctx},
		ret:    []any{ret, err},
	})
}

// ExpectGetSubmission sets up an expectation for GetSubmission method
func (m *MockService) ExpectGetSubmission(ctx context.Context, id int64, ret *contact.Submission, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "GetSubmission",
		args:   []any{ctx, id},
		ret:    []any{ret, err},
	})
}

// ExpectUpdateSubmissionStatus sets up an expectation for UpdateSubmissionStatus method
func (m *MockService) ExpectUpdateSubmissionStatus(ctx context.Context, id int64, status contact.Status, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, mockCall{
		method: "UpdateSubmissionStatus",
		args:   []any{ctx, id, status},
		ret:    []any{err},
	})
}

// Submit mocks the Submit method
func (m *MockService) Submit(ctx context.Context, sub *contact.Submission) error {
	ret := m.recordCall("Submit", []any{ctx, sub})
	if len(ret) == 0 || ret[0] == nil {
		return nil
	}
	if err, ok := ret[0].(error); ok {
		return err
	}
	return fmt.Errorf("invalid error type returned from mock")
}

// ListSubmissions mocks the ListSubmissions method
func (m *MockService) ListSubmissions(ctx context.Context) ([]contact.Submission, error) {
	ret := m.recordCall("ListSubmissions", []any{ctx})
	if len(ret) < 2 {
		return nil, nil
	}

	var subs []contact.Submission
	if ret[0] != nil {
		if s, ok := ret[0].([]contact.Submission); ok {
			subs = s
		} else {
			return nil, fmt.Errorf("invalid submissions type returned from mock")
		}
	}

	var err error
	if ret[1] != nil {
		if e, ok := ret[1].(error); ok {
			err = e
		} else {
			return nil, fmt.Errorf("invalid error type returned from mock")
		}
	}
	return subs, err
}

// GetSubmission mocks the GetSubmission method
func (m *MockService) GetSubmission(ctx context.Context, id int64) (*contact.Submission, error) {
	ret := m.recordCall("GetSubmission", []any{ctx, id})
	if len(ret) < 2 {
		return nil, nil
	}

	var sub *contact.Submission
	if ret[0] != nil {
		if s, ok := ret[0].(*contact.Submission); ok {
			sub = s
		} else {
			return nil, fmt.Errorf("invalid submission type returned from mock")
		}
	}

	var err error
	if ret[1] != nil {
		if e, ok := ret[1].(error); ok {
			err = e
		} else {
			return nil, fmt.Errorf("invalid error type returned from mock")
		}
	}
	return sub, err
}

// UpdateSubmissionStatus mocks the UpdateSubmissionStatus method
func (m *MockService) UpdateSubmissionStatus(ctx context.Context, id int64, status contact.Status) error {
	ret := m.recordCall("UpdateSubmissionStatus", []any{ctx, id, status})
	if len(ret) == 0 || ret[0] == nil {
		return nil
	}
	if err, ok := ret[0].(error); ok {
		return err
	}
	return fmt.Errorf("invalid error type returned from mock")
}

// Verify checks if all expected calls were made
func (m *MockService) Verify() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.expected) != len(m.calls) {
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
func (m *MockService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = nil
	m.expected = nil
}
