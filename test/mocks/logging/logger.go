package logging

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// AnyValue is a placeholder for any field value
type AnyValue struct{}

// logCall represents a single logging call
type logCall struct {
	level   string
	message string
	fields  map[string]interface{}
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
	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		fieldMap[field.Key] = field.String
	}
	m.calls = append(m.calls, logCall{level: level, message: message, fields: fieldMap})
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

// Field creation methods
func (m *MockLogger) Int64(key string, value int64) logging.Field {
	return zap.Int64(key, value)
}

func (m *MockLogger) Int(key string, value int) logging.Field {
	return zap.Int(key, value)
}

func (m *MockLogger) Int32(key string, value int32) logging.Field {
	return zap.Int32(key, value)
}

func (m *MockLogger) Uint64(key string, value uint64) logging.Field {
	return zap.Uint64(key, value)
}

func (m *MockLogger) Uint(key string, value uint) logging.Field {
	return zap.Uint(key, value)
}

func (m *MockLogger) Uint32(key string, value uint32) logging.Field {
	return zap.Uint32(key, value)
}

func (m *MockLogger) String(key string, value string) logging.Field {
	return zap.String(key, value)
}

func (m *MockLogger) Bool(key string, value bool) logging.Field {
	return zap.Bool(key, value)
}

func (m *MockLogger) ErrorField(err error) logging.Field {
	return zap.Error(err)
}

func (m *MockLogger) Duration(key string, value time.Duration) logging.Field {
	return zap.Duration(key, value)
}

// Static field creation methods
func Int64(key string, value int64) logging.Field {
	return zap.Int64(key, value)
}

func Int(key string, value int) logging.Field {
	return zap.Int(key, value)
}

func Int32(key string, value int32) logging.Field {
	return zap.Int32(key, value)
}

func Uint64(key string, value uint64) logging.Field {
	return zap.Uint64(key, value)
}

func Uint(key string, value uint) logging.Field {
	return zap.Uint(key, value)
}

func Uint32(key string, value uint32) logging.Field {
	return zap.Uint32(key, value)
}

func String(key string, value string) logging.Field {
	return zap.String(key, value)
}

func Bool(key string, value bool) logging.Field {
	return zap.Bool(key, value)
}

func ErrorField(err error) logging.Field {
	return zap.Error(err)
}

func Duration(key string, value time.Duration) logging.Field {
	return zap.Duration(key, value)
}

// ExpectInfo adds an expectation for an info message
func (m *MockLogger) ExpectInfo(message string) *logCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := logCall{level: "info", message: message, fields: make(map[string]interface{})}
	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// ExpectError adds an expectation for an error message
func (m *MockLogger) ExpectError(message string) *logCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := logCall{level: "error", message: message, fields: make(map[string]interface{})}
	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// ExpectDebug adds an expectation for a debug message
func (m *MockLogger) ExpectDebug(message string) *logCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := logCall{level: "debug", message: message, fields: make(map[string]interface{})}
	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// ExpectWarn adds an expectation for a warning message
func (m *MockLogger) ExpectWarn(message string) *logCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := logCall{level: "warn", message: message, fields: make(map[string]interface{})}
	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// WithFields adds field expectations to a log call
func (c *logCall) WithFields(fields map[string]interface{}) *logCall {
	c.fields = fields
	return c
}

// Verify checks if all expected calls were made
func (m *MockLogger) Verify() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.expected) != len(m.calls) {
		return fmt.Errorf("expected %d calls but got %d", len(m.expected), len(m.calls))
	}

	for i, exp := range m.expected {
		got := m.calls[i]
		if exp.level != got.level {
			return fmt.Errorf("call %d: expected level %q but got %q", i, exp.level, got.level)
		}
		if !strings.Contains(got.message, exp.message) {
			return fmt.Errorf("call %d: expected message %q but got %q", i, exp.message, got.message)
		}

		// Check fields
		for key, expValue := range exp.fields {
			gotValue, ok := got.fields[key]
			if !ok {
				return fmt.Errorf("call %d: missing field %q", i, key)
			}
			if _, isAny := expValue.(AnyValue); !isAny && expValue != gotValue {
				return fmt.Errorf("call %d: field %q expected %v but got %v", i, key, expValue, gotValue)
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
