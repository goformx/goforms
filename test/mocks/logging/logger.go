package logging

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/stretchr/testify/mock"
)

// AnyValue is a placeholder for any field value
type AnyValue struct{}

// LogCall represents a single logging call (exported for test chaining)
type LogCall struct {
	level   string
	message string
	fields  map[string]any
}

// MockLogger is a mock implementation of the Logger interface
type MockLogger struct {
	mock.Mock
	mu       sync.Mutex
	expected []LogCall
	calls    []LogCall
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) recordCall(level, message string, fields ...logging.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fieldMap := make(map[string]any)
	for _, field := range fields {
		if field.Key == "error" {
			fieldMap[field.Key] = field.Interface
		} else {
			fieldMap[field.Key] = field.String
		}
	}
	m.calls = append(m.calls, LogCall{level: level, message: message, fields: fieldMap})
}

// Info logs an info level message
func (m *MockLogger) Info(msg string, fields ...logging.Field) {
	m.recordCall("info", msg, fields...)
}

// Error logs an error level message
func (m *MockLogger) Error(message string, fields ...logging.Field) {
	m.recordCall("error", message, fields...)
}

// Debug logs a debug level message
func (m *MockLogger) Debug(msg string, fields ...logging.Field) {
	m.recordCall("debug", msg, fields...)
}

// Warn logs a warning level message
func (m *MockLogger) Warn(msg string, fields ...logging.Field) {
	m.recordCall("warn", msg, fields...)
}

// Field creation methods
// Int64 returns a field with an int64 value.
func (m *MockLogger) Int64(key string, value int64) logging.Field {
	return zap.Int64(key, value)
}

// Int creates a field with an int value
func (m *MockLogger) Int(key string, value int) logging.Field {
	return zap.Int(key, value)
}

// Int32 creates a field with an int32 value
func (m *MockLogger) Int32(key string, value int32) logging.Field {
	return zap.Int32(key, value)
}

// Uint64 creates a field with a uint64 value
func (m *MockLogger) Uint64(key string, value uint64) logging.Field {
	return zap.Uint64(key, value)
}

// Uint creates a field with a uint value
func (m *MockLogger) Uint(key string, value uint) logging.Field {
	return zap.Uint(key, value)
}

// Uint32 creates a field with a uint32 value
func (m *MockLogger) Uint32(key string, value uint32) logging.Field {
	return zap.Uint32(key, value)
}

// String creates a string field
func (m *MockLogger) String(key, value string) logging.Field {
	return zap.String(key, value)
}

// Bool creates a field with a boolean value
func (m *MockLogger) Bool(key string, value bool) logging.Field {
	return zap.Bool(key, value)
}

// ErrorField creates a field with an error value
func (m *MockLogger) ErrorField(err error) logging.Field {
	return zap.Error(err)
}

// Duration creates a field with a duration value
func (m *MockLogger) Duration(key string, value time.Duration) logging.Field {
	return zap.Duration(key, value)
}

// Int64 returns a field with an int64 value.
func Int64(key string, value int64) logging.Field {
	return zap.Int64(key, value)
}

// Int creates a field with an int value
func Int(key string, value int) logging.Field {
	return zap.Int(key, value)
}

// Int32 creates a field with an int32 value
func Int32(key string, value int32) logging.Field {
	return zap.Int32(key, value)
}

// Uint64 creates a field with a uint64 value
func Uint64(key string, value uint64) logging.Field {
	return zap.Uint64(key, value)
}

// Uint creates a field with a uint value
func Uint(key string, value uint) logging.Field {
	return zap.Uint(key, value)
}

// Uint32 creates a field with a uint32 value
func Uint32(key string, value uint32) logging.Field {
	return zap.Uint32(key, value)
}

// String creates a string field
func String(key, value string) logging.Field {
	return zap.String(key, value)
}

// Bool creates a field with a boolean value
func Bool(key string, value bool) logging.Field {
	return zap.Bool(key, value)
}

// ErrorField creates a field with an error value
func ErrorField(err error) logging.Field {
	return zap.Error(err)
}

// Duration creates a field with a duration value
func Duration(key string, value time.Duration) logging.Field {
	return zap.Duration(key, value)
}

// ExpectInfo adds an expectation for an info message and returns a pointer for chaining in tests
func (m *MockLogger) ExpectInfo(message string) *LogCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := LogCall{level: "info", message: message, fields: make(map[string]any)}
	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// ExpectError adds an expectation for an error message and returns a pointer for chaining in tests
func (m *MockLogger) ExpectError(message string) *LogCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := LogCall{level: "error", message: message, fields: make(map[string]any)}
	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// ExpectDebug adds an expectation for a debug message and returns a pointer for chaining in tests
func (m *MockLogger) ExpectDebug(message string) *LogCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := LogCall{level: "debug", message: message, fields: make(map[string]any)}
	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// ExpectWarn adds an expectation for a warning message and returns a pointer for chaining in tests
func (m *MockLogger) ExpectWarn(message string) *LogCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	call := LogCall{level: "warn", message: message, fields: make(map[string]any)}
	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// WithFields adds field expectations to a log call
func (c *LogCall) WithFields(fields map[string]any) *LogCall {
	c.fields = fields
	return c
}

// Verify checks if all expected calls were made
func (m *MockLogger) Verify() error {
	m.AssertExpectations(mock.TestingT(nil))
	return nil
}

// Reset clears all calls and expectations
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = m.calls[:0]
	m.expected = m.expected[:0]
}
