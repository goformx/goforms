// Package logging provides a unified logging interface using zap
package logging

import (
	"strconv"
	"time"

	"go.uber.org/zap"
)

// LogField represents a field in a log entry
type LogField struct {
	Type   FieldType
	Key    string
	String string
	Int    int
	Error  error
	Uint   uint
	Any    any
}

// FieldType represents the type of a log field
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
func StringField(key, value string) LogField {
	return LogField{
		Type:   StringType,
		Key:    key,
		String: value,
	}
}

// IntField creates an integer field
func IntField(key string, value int) LogField {
	return LogField{
		Type: IntType,
		Key:  key,
		Int:  value,
	}
}

// ErrorField creates an error field
func ErrorField(key string, err error) LogField {
	return LogField{
		Type:  ErrorType,
		Key:   key,
		Error: err,
	}
}

// DurationField creates a duration field
func DurationField(key string, value time.Duration) LogField {
	return LogField{
		Type:   DurationType,
		Key:    key,
		String: value.String(),
	}
}

// BoolField creates a boolean field
func BoolField(key string, value bool) LogField {
	return LogField{
		Type:   BoolType,
		Key:    key,
		String: strconv.FormatBool(value),
	}
}

// AnyField creates a field for any type
func AnyField(key string, value any) LogField {
	return LogField{
		Type: AnyType,
		Key:  key,
		Any:  value,
	}
}

// UintField creates an unsigned integer field
func UintField(key string, value uint) LogField {
	return LogField{
		Type: UintType,
		Key:  key,
		Uint: value,
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
