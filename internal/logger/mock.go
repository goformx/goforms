package logger

import (
	"strings"
)

// MockLogCall represents a single log call
type MockLogCall struct {
	Message string
	Fields  []Field
}

// MockLogger is a mock implementation of the Logger interface for testing
type MockLogger struct {
	InfoCalls  []MockLogCall
	ErrorCalls []MockLogCall
	WarnCalls  []MockLogCall
	DebugCalls []MockLogCall
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		InfoCalls:  make([]MockLogCall, 0),
		ErrorCalls: make([]MockLogCall, 0),
		WarnCalls:  make([]MockLogCall, 0),
		DebugCalls: make([]MockLogCall, 0),
	}
}

// Info logs an info message
func (m *MockLogger) Info(msg string, fields ...Field) {
	m.InfoCalls = append(m.InfoCalls, MockLogCall{Message: msg, Fields: fields})
}

// Error logs an error message
func (m *MockLogger) Error(msg string, fields ...Field) {
	m.ErrorCalls = append(m.ErrorCalls, MockLogCall{Message: msg, Fields: fields})
}

// Warn logs a warning message
func (m *MockLogger) Warn(msg string, fields ...Field) {
	m.WarnCalls = append(m.WarnCalls, MockLogCall{Message: msg, Fields: fields})
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, fields ...Field) {
	m.DebugCalls = append(m.DebugCalls, MockLogCall{Message: msg, Fields: fields})
}

// HasInfoLog checks if a specific info message was logged
func (m *MockLogger) HasInfoLog(msg string) bool {
	for _, call := range m.InfoCalls {
		if strings.Contains(call.Message, msg) {
			return true
		}
	}
	return false
}
