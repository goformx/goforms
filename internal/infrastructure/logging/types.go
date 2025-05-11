// Package logging provides a unified logging interface using zap
package logging

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// LogField represents a structured logging field
type LogField struct {
	Key    string
	Type   FieldType
	String string
	Int    int
	Error  error
}

// FieldType represents the type of a logging field
type FieldType int

const (
	StringType FieldType = iota
	IntType
	ErrorType
	DurationType
	BoolType
	AnyType
	UintType
)

// Field represents a logging field
type Field = zap.Field

// String creates a string field
func String(key, value string) Field { return zap.String(key, value) }

// Int creates an integer field
func Int(key string, value int) Field { return zap.Int(key, value) }

// Int64 creates a 64-bit integer field
func Int64(key string, value int64) Field { return zap.Int64(key, value) }

// Uint creates an unsigned integer field
func Uint(key string, value uint) Field { return zap.Uint(key, value) }

// Bool creates a boolean field
func Bool(key string, value bool) Field { return zap.Bool(key, value) }

// Error creates an error field
func Error(err error) Field { return zap.Error(err) }

// Duration creates a duration field
func Duration(key string, value time.Duration) Field { return zap.Duration(key, value) }

// Any creates a field with any value
func Any(key string, value any) Field { return zap.Any(key, value) }

// StringField creates a string field
func StringField(key string, value string) LogField {
	return LogField{
		Key:    key,
		Type:   StringType,
		String: value,
	}
}

// IntField creates an int field
func IntField(key string, value int) LogField {
	return LogField{
		Key:  key,
		Type: IntType,
		Int:  value,
	}
}

// ErrorField creates an error field
func ErrorField(key string, err error) LogField {
	return LogField{
		Key:   key,
		Type:  ErrorType,
		Error: err,
	}
}

// DurationField creates a duration field
func DurationField(key string, value time.Duration) LogField {
	return LogField{
		Key:    key,
		Type:   DurationType,
		String: value.String(),
	}
}

// BoolField creates a boolean field
func BoolField(key string, value bool) LogField {
	return LogField{
		Key:    key,
		Type:   BoolType,
		String: fmt.Sprintf("%v", value),
	}
}

// AnyField creates a field from any value
func AnyField(key string, value interface{}) LogField {
	return LogField{
		Key:    key,
		Type:   AnyType,
		String: fmt.Sprintf("%v", value),
	}
}

// UintField creates an unsigned integer field
func UintField(key string, value uint) LogField {
	return LogField{
		Key:    key,
		Type:   UintType,
		String: fmt.Sprintf("%d", value),
	}
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string, fields ...LogField)
	Info(msg string, fields ...LogField)
	Warn(msg string, fields ...LogField)
	Error(msg string, fields ...LogField)
	Fatal(msg string, fields ...LogField)
	With(fields ...LogField) Logger
}
