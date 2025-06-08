// Package logging provides a unified logging interface
package logging

import (
	"time"
)

// Field represents a field in a log entry
type Field = LogField

// Logger defines the interface for application logging
type Logger interface {
	// Basic logging methods
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)

	// Context methods
	With(fields ...any) Logger
	WithComponent(component string) Logger
	WithOperation(operation string) Logger
	WithRequestID(requestID string) Logger
	WithUserID(userID string) Logger
	WithError(err error) Logger
	WithFields(fields map[string]any) Logger
}

// LogField represents a field in a log entry
type LogField struct {
	Key   string
	Value any
}

// String creates a string field
func String(key, value string) LogField {
	return LogField{Key: key, Value: value}
}

// Int creates an integer field
func Int(key string, value int) LogField {
	return LogField{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) LogField {
	return LogField{Key: key, Value: value}
}

// Uint creates an unsigned integer field
func Uint(key string, value uint) LogField {
	return LogField{Key: key, Value: value}
}

// Uint64 creates an uint64 field
func Uint64(key string, value uint64) LogField {
	return LogField{Key: key, Value: value}
}

// Float32 creates a float32 field
func Float32(key string, value float32) LogField {
	return LogField{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) LogField {
	return LogField{Key: key, Value: value}
}

// Bool creates a boolean field
func Bool(key string, value bool) LogField {
	return LogField{Key: key, Value: value}
}

// Time creates a time field
func Time(key string, value time.Time) LogField {
	return LogField{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) LogField {
	return LogField{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) LogField {
	return LogField{Key: "error", Value: err}
}

// ErrorField creates an error field with a custom key
func ErrorField(key string, err error) LogField {
	return LogField{Key: key, Value: err}
}
