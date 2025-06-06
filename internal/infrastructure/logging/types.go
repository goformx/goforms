// Package logging provides a unified logging interface
package logging

import (
	"time"

	"github.com/goformx/goforms/internal/infrastructure/common"
)

// Field represents a field in a log entry
type Field = LogField

// Logger defines the interface for application logging
type Logger interface {
	common.Logger

	// Additional logging methods
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

// Error creates an error field
func Error(err error) LogField {
	return LogField{Key: "error", Value: err}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) LogField {
	return LogField{Key: key, Value: value.String()}
}

// StringField creates a string field
func StringField(key, value string) LogField {
	return LogField{Key: key, Value: value}
}

// IntField creates an integer field
func IntField(key string, value int) LogField {
	return LogField{Key: key, Value: value}
}

// ErrorField creates an error field
func ErrorField(key string, err error) LogField {
	return LogField{Key: key, Value: err}
}

// DurationField creates a duration field
func DurationField(key string, value time.Duration) LogField {
	return LogField{Key: key, Value: value.String()}
}

// BoolField creates a boolean field
func BoolField(key string, value bool) LogField {
	return LogField{Key: key, Value: value}
}

// UintField creates an unsigned integer field
func UintField(key string, value uint) LogField {
	return LogField{Key: key, Value: value}
}

// Int64Field creates an int64 field
func Int64Field(key string, value int64) LogField {
	return LogField{Key: key, Value: value}
}
