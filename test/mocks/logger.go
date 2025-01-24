package mocks

import (
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
)

// Logger implements logging.Logger interface for testing
type Logger struct {
	InfoCalls []struct {
		Message string
		Fields  []logging.Field
	}
	ErrorCalls []struct {
		Message string
		Fields  []logging.Field
	}
	DebugCalls []struct {
		Message string
		Fields  []logging.Field
	}
	WarnCalls []struct {
		Message string
		Fields  []logging.Field
	}
}

// NewLogger creates a new mock logger for testing
func NewLogger() *Logger {
	return &Logger{}
}

func (m *Logger) Info(msg string, fields ...logging.Field) {
	m.InfoCalls = append(m.InfoCalls, struct {
		Message string
		Fields  []logging.Field
	}{msg, fields})
}

func (m *Logger) Error(msg string, fields ...logging.Field) {
	m.ErrorCalls = append(m.ErrorCalls, struct {
		Message string
		Fields  []logging.Field
	}{msg, fields})
}

func (m *Logger) Debug(msg string, fields ...logging.Field) {
	m.DebugCalls = append(m.DebugCalls, struct {
		Message string
		Fields  []logging.Field
	}{msg, fields})
}

func (m *Logger) Warn(msg string, fields ...logging.Field) {
	m.WarnCalls = append(m.WarnCalls, struct {
		Message string
		Fields  []logging.Field
	}{msg, fields})
}

// HasInfoLog checks if a specific info message was logged
func (m *Logger) HasInfoLog(msg string) bool {
	for _, call := range m.InfoCalls {
		if call.Message == msg {
			return true
		}
	}
	return false
}

// HasErrorLog checks if a specific error message was logged
func (m *Logger) HasErrorLog(msg string) bool {
	for _, call := range m.ErrorCalls {
		if call.Message == msg {
			return true
		}
	}
	return false
}

// HasDebugLog checks if a specific debug message was logged
func (m *Logger) HasDebugLog(msg string) bool {
	for _, call := range m.DebugCalls {
		if call.Message == msg {
			return true
		}
	}
	return false
}

// HasWarnLog checks if a specific warning message was logged
func (m *Logger) HasWarnLog(msg string) bool {
	for _, call := range m.WarnCalls {
		if call.Message == msg {
			return true
		}
	}
	return false
}
