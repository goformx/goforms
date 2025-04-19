package contact

import (
	"context"
	"errors"

	"github.com/jonesrussell/goforms/internal/domain/contact"
)

type MockStore struct {
	expectations []expectation
}

type expectation struct {
	method   string
	args     []interface{}
	returns  []interface{}
	executed bool
}

func NewMockStore() *MockStore {
	return &MockStore{}
}

func (m *MockStore) Create(ctx context.Context, sub *contact.Submission) error {
	if len(m.expectations) == 0 {
		return errors.New("no expectations set")
	}

	e := m.expectations[0]
	if e.method != "Create" {
		return errors.New("unexpected method call")
	}

	m.expectations = m.expectations[1:]
	e.executed = true

	if len(e.returns) > 0 {
		if err, ok := e.returns[0].(error); ok {
			return err
		}
	}

	return nil
}

func (m *MockStore) List(ctx context.Context) ([]contact.Submission, error) {
	if len(m.expectations) == 0 {
		return nil, errors.New("no expectations set")
	}

	e := m.expectations[0]
	if e.method != "List" {
		return nil, errors.New("unexpected method call")
	}

	m.expectations = m.expectations[1:]
	e.executed = true

	if len(e.returns) > 0 {
		if submissions, ok := e.returns[0].([]contact.Submission); ok {
			return submissions, nil
		}
		return nil, errors.New("invalid return type")
	}

	return nil, nil
}

func (m *MockStore) Get(ctx context.Context, id int64) (*contact.Submission, error) {
	if len(m.expectations) == 0 {
		return nil, errors.New("no expectations set")
	}

	e := m.expectations[0]
	if e.method != "Get" {
		return nil, errors.New("unexpected method call")
	}

	m.expectations = m.expectations[1:]
	e.executed = true

	if len(e.returns) > 0 {
		if submission, ok := e.returns[0].(*contact.Submission); ok {
			return submission, nil
		}
		return nil, errors.New("invalid return type")
	}

	return nil, nil
}

func (m *MockStore) UpdateStatus(ctx context.Context, id int64, status contact.Status) error {
	if len(m.expectations) == 0 {
		return errors.New("no expectations set")
	}

	e := m.expectations[0]
	if e.method != "UpdateStatus" {
		return errors.New("unexpected method call")
	}

	m.expectations = m.expectations[1:]
	e.executed = true

	if len(e.returns) > 0 {
		if err, ok := e.returns[0].(error); ok {
			return err
		}
	}

	return nil
}

func (m *MockStore) Expect(method string, args ...interface{}) *MockStore {
	m.expectations = append(m.expectations, expectation{
		method:   method,
		args:     args,
		executed: false,
	})
	return m
}

func (m *MockStore) Return(returns ...interface{}) *MockStore {
	if len(m.expectations) == 0 {
		return m
	}
	m.expectations[len(m.expectations)-1].returns = returns
	return m
}

func (m *MockStore) ExpectationsWereMet() error {
	for _, e := range m.expectations {
		if !e.executed {
			return errors.New("not all expectations were met")
		}
	}
	return nil
}
