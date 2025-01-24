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
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...logging.Field) {
	m.Called(msg, fields)
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
			return true
		}
	}
	return false
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
	return m.On("Info", msg, fields).Return()
}

// ExpectError sets up an expectation for an error message
func (m *MockLogger) ExpectError(msg string, fields ...logging.Field) *mock.Call {
	return m.On("Error", msg, fields).Return()
}

// ExpectDebug sets up an expectation for a debug message
func (m *MockLogger) ExpectDebug(msg string, fields ...logging.Field) *mock.Call {
	return m.On("Debug", msg, fields).Return()
}

// ExpectWarn sets up an expectation for a warning message
func (m *MockLogger) ExpectWarn(msg string, fields ...logging.Field) *mock.Call {
	return m.On("Warn", msg, fields).Return()
}
