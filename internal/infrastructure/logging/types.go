// Package logging provides a unified logging interface
package logging

import (
	"time"

	"github.com/goformx/goforms/internal/infrastructure/common"
)

// Logger extends the common.Logger interface with additional functionality
type Logger interface {
	common.Logger
	Fatal(msg string, fields ...any)
	With(fields ...any) Logger
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

// Any creates a field for any type
func Any(key string, value any) LogField {
	return LogField{Key: key, Value: value}
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
