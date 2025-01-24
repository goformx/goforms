package logging

import (
	"github.com/stretchr/testify/mock"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// MockLogger is a mock implementation of logging.Logger
type MockLogger struct {
	mock.Mock
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Info(msg string, fields ...logging.Field) {
	args := []interface{}{msg}
	if len(fields) > 0 {
		args = append(args, fields)
	}
	m.Called(args...)
}

func (m *MockLogger) Error(msg string, fields ...logging.Field) {
	args := []interface{}{msg}
	if len(fields) > 0 {
		args = append(args, fields)
	}
	m.Called(args...)
}

func (m *MockLogger) Debug(msg string, fields ...logging.Field) {
	args := []interface{}{msg}
	if len(fields) > 0 {
		args = append(args, fields)
	}
	m.Called(args...)
}

func (m *MockLogger) Warn(msg string, fields ...logging.Field) {
	args := []interface{}{msg}
	if len(fields) > 0 {
		args = append(args, fields)
	}
	m.Called(args...)
}

// Helper methods for common verifications

// VerifyInfo verifies that an info message was logged
func (m *MockLogger) VerifyInfo(msg string) bool {
	for _, call := range m.Calls {
		if call.Method == "Info" && call.Arguments[0].(string) == msg {
			return true
		}
	}
	return false
}

// VerifyError verifies that an error message was logged
func (m *MockLogger) VerifyError(msg string) bool {
	for _, call := range m.Calls {
		if call.Method == "Error" && call.Arguments[0].(string) == msg {
			return false
		}
	}
	return true
}

// VerifyDebug verifies that a debug message was logged
func (m *MockLogger) VerifyDebug(msg string) bool {
	for _, call := range m.Calls {
		if call.Method == "Debug" && call.Arguments[0].(string) == msg {
			return true
		}
	}
	return false
}

// VerifyWarn verifies that a warning message was logged
func (m *MockLogger) VerifyWarn(msg string) bool {
	for _, call := range m.Calls {
		if call.Method == "Warn" && call.Arguments[0].(string) == msg {
			return true
		}
	}
	return false
}

// VerifyNoErrors verifies that no errors were logged
func (m *MockLogger) VerifyNoErrors() bool {
	for _, call := range m.Calls {
		if call.Method == "Error" {
			return false
		}
	}
	return true
}

// ExpectInfo sets up an expectation for an info message
func (m *MockLogger) ExpectInfo(msg string, fields ...logging.Field) *mock.Call {
	args := []interface{}{msg}
	if len(fields) > 0 {
		args = append(args, mock.MatchedBy(func(actual []logging.Field) bool {
			if len(actual) != len(fields) {
				return false
			}
			for i, field := range fields {
				if field.String == "mock.Anything" {
					continue
				}
				if actual[i].Key != field.Key || actual[i].String != field.String {
					return false
				}
			}
			return true
		}))
	}
	return m.On("Info", args...).Return()
}

// ExpectError sets up an expectation for an error message
func (m *MockLogger) ExpectError(msg string, fields ...logging.Field) *mock.Call {
	args := []interface{}{msg}
	if len(fields) > 0 {
		args = append(args, mock.MatchedBy(func(actual []logging.Field) bool {
			if len(actual) != len(fields) {
				return false
			}
			for i, field := range fields {
				if field.String == "mock.Anything" {
					continue
				}
				if actual[i].Key != field.Key || actual[i].String != field.String {
					return false
				}
			}
			return true
		}))
	}
	return m.On("Error", args...).Return()
}

// ExpectDebug sets up an expectation for a debug message
func (m *MockLogger) ExpectDebug(msg string, fields ...logging.Field) *mock.Call {
	args := []interface{}{msg}
	if len(fields) > 0 {
		args = append(args, mock.MatchedBy(func(actual []logging.Field) bool {
			if len(actual) != len(fields) {
				return false
			}
			for i, field := range fields {
				if field.String == "mock.Anything" {
					continue
				}
				if actual[i].Key != field.Key || actual[i].String != field.String {
					return false
				}
			}
			return true
		}))
	}
	return m.On("Debug", args...).Return()
}

// ExpectWarn sets up an expectation for a warning message
func (m *MockLogger) ExpectWarn(msg string, fields ...logging.Field) *mock.Call {
	args := []interface{}{msg}
	if len(fields) > 0 {
		args = append(args, mock.MatchedBy(func(actual []logging.Field) bool {
			if len(actual) != len(fields) {
				return false
			}
			for i, field := range fields {
				if field.String == "mock.Anything" {
					continue
				}
				if actual[i].Key != field.Key || actual[i].String != field.String {
					return false
				}
			}
			return true
		}))
	}
	return m.On("Warn", args...).Return()
}
