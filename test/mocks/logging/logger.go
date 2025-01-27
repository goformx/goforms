package logging

import (
	"fmt"
	"sync"

	forbidden_zap "go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// logCall represents a single logging call
type logCall struct {
	level   string
	message string
	fields  []logging.Field
}

// MockLogger is a mock implementation of logging.Logger
type MockLogger struct {
	mu       sync.Mutex
	calls    []logCall
	expected []logCall
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) recordCall(level, message string, fields ...logging.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, logCall{level: level, message: message, fields: fields})
}

func (m *MockLogger) Info(message string, fields ...logging.Field) {
	m.recordCall("info", message, fields...)
}

func (m *MockLogger) Error(message string, fields ...logging.Field) {
	m.recordCall("error", message, fields...)
}

func (m *MockLogger) Debug(message string, fields ...logging.Field) {
	m.recordCall("debug", message, fields...)
}

func (m *MockLogger) Warn(message string, fields ...logging.Field) {
	m.recordCall("warn", message, fields...)
}

func (m *MockLogger) Int64(key string, value int64) logging.Field {
	return forbidden_zap.Int64(key, value)
}

func (m *MockLogger) Int(key string, value int) logging.Field {
	return forbidden_zap.Int(key, value)
}

func (m *MockLogger) Int32(key string, value int32) logging.Field {
	return forbidden_zap.Int32(key, value)
}

func (m *MockLogger) Uint64(key string, value uint64) logging.Field {
	return forbidden_zap.Uint64(key, value)
}

func (m *MockLogger) Uint(key string, value uint) logging.Field {
	return forbidden_zap.Uint(key, value)
}

func (m *MockLogger) Uint32(key string, value uint32) logging.Field {
	return forbidden_zap.Uint32(key, value)
}

// ExpectInfo adds an expectation for an info message
func (m *MockLogger) ExpectInfo(message string, fields ...logging.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, logCall{level: "info", message: message, fields: fields})
}

// ExpectError adds an expectation for an error message
func (m *MockLogger) ExpectError(message string, fields ...logging.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, logCall{level: "error", message: message, fields: fields})
}

// ExpectDebug adds an expectation for a debug message
func (m *MockLogger) ExpectDebug(message string, fields ...logging.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, logCall{level: "debug", message: message, fields: fields})
}

// ExpectWarn adds an expectation for a warning message
func (m *MockLogger) ExpectWarn(message string, fields ...logging.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expected = append(m.expected, logCall{level: "warn", message: message, fields: fields})
}

// matchField compares two logging fields
func matchField(got, exp logging.Field) bool {
	if exp.Key != got.Key {
		return false
	}
	// For error fields, just check if both are errors
	if _, gotIsErr := got.Interface.(error); gotIsErr {
		_, expIsErr := exp.Interface.(error)
		return expIsErr
	}
	return exp.Interface == got.Interface
}

// Verify checks if all expected calls were made in order
func (m *MockLogger) Verify() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.calls) != len(m.expected) {
		return fmt.Errorf("expected %d calls but got %d", len(m.expected), len(m.calls))
	}

	for i, exp := range m.expected {
		got := m.calls[i]
		if exp.level != got.level {
			return fmt.Errorf("call %d: expected level %s but got %s", i, exp.level, got.level)
		}
		if exp.message != got.message {
			return fmt.Errorf("call %d: expected message %q but got %q", i, exp.message, got.message)
		}
		if len(exp.fields) != len(got.fields) {
			return fmt.Errorf("call %d: expected %d fields but got %d", i, len(exp.fields), len(got.fields))
		}
		for j, expField := range exp.fields {
			gotField := got.fields[j]
			if !matchField(gotField, expField) {
				return fmt.Errorf("call %d field %d: fields do not match", i, j)
			}
		}
	}
	return nil
}

// Reset clears all calls and expectations
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = m.calls[:0]
	m.expected = m.expected[:0]
}
