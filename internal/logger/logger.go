// Package logger provides a unified logging interface using zap
package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
}

// Field represents a log field
type Field = zap.Field

type zapLogger struct {
	*zap.Logger
}

// GetLogger returns a new production logger instance
func GetLogger() Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	log, _ := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	return &zapLogger{log}
}

// NewTestLogger returns a logger suitable for testing
func NewTestLogger() Logger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	log, _ := config.Build()
	return &zapLogger{log}
}

// NewMockLogger returns a mock logger for testing
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (l *zapLogger) Info(msg string, fields ...Field)  { l.Logger.Info(msg, fields...) }
func (l *zapLogger) Error(msg string, fields ...Field) { l.Logger.Error(msg, fields...) }
func (l *zapLogger) Debug(msg string, fields ...Field) { l.Logger.Debug(msg, fields...) }
func (l *zapLogger) Warn(msg string, fields ...Field)  { l.Logger.Warn(msg, fields...) }

// Field constructors
func String(key string, value string) Field          { return zap.String(key, value) }
func Int(key string, value int) Field                { return zap.Int(key, value) }
func Error(err error) Field                          { return zap.Error(err) }
func Duration(key string, value time.Duration) Field { return zap.Duration(key, value) }
func Any(key string, value interface{}) Field        { return zap.Any(key, value) }

// UnderlyingZap returns the underlying zap logger for advanced usage
func UnderlyingZap(l Logger) *zap.Logger {
	if zl, ok := l.(*zapLogger); ok {
		return zl.Logger
	}
	return zap.NewNop()
}

// MockLogger implements Logger interface for testing
type MockLogger struct {
	InfoCalls []struct {
		Message string
		Fields  []Field
	}
	ErrorCalls []struct {
		Message string
		Fields  []Field
	}
	DebugCalls []struct {
		Message string
		Fields  []Field
	}
	WarnCalls []struct {
		Message string
		Fields  []Field
	}
}

func (m *MockLogger) Info(msg string, fields ...Field) {
	m.InfoCalls = append(m.InfoCalls, struct {
		Message string
		Fields  []Field
	}{msg, fields})
}

func (m *MockLogger) Error(msg string, fields ...Field) {
	m.ErrorCalls = append(m.ErrorCalls, struct {
		Message string
		Fields  []Field
	}{msg, fields})
}

func (m *MockLogger) Debug(msg string, fields ...Field) {
	m.DebugCalls = append(m.DebugCalls, struct {
		Message string
		Fields  []Field
	}{msg, fields})
}

func (m *MockLogger) Warn(msg string, fields ...Field) {
	m.WarnCalls = append(m.WarnCalls, struct {
		Message string
		Fields  []Field
	}{msg, fields})
}

// HasInfoLog checks if a specific info message was logged
func (m *MockLogger) HasInfoLog(msg string) bool {
	for _, call := range m.InfoCalls {
		if call.Message == msg {
			return true
		}
	}
	return false
}

// HasErrorLog checks if a specific error message was logged
func (m *MockLogger) HasErrorLog(msg string) bool {
	for _, call := range m.ErrorCalls {
		if call.Message == msg {
			return true
		}
	}
	return false
}
