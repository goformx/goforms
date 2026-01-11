package logging

import (
	"go.uber.org/zap"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// logger implements the Logger interface using zap
type logger struct {
	zapLogger      *zap.Logger
	sanitizer      sanitization.ServiceInterface
	fieldSanitizer *Sanitizer
}

// newLogger creates a new logger instance
func newLogger(
	zapLogger *zap.Logger,
	sanitizer sanitization.ServiceInterface,
	fieldSanitizer *Sanitizer,
) Logger {
	return &logger{
		zapLogger:      zapLogger,
		sanitizer:      sanitizer,
		fieldSanitizer: fieldSanitizer,
	}
}

// With returns a new logger with the given fields
func (l *logger) With(fields ...any) Logger {
	zapFields := convertToZapFields(fields, l.fieldSanitizer)

	return newLogger(l.zapLogger.With(zapFields...), l.sanitizer, l.fieldSanitizer)
}

// WithFieldsStructured adds multiple fields to the logger using the new Field-based API
func (l *logger) WithFieldsStructured(fields ...Field) Logger {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = field.ToZapField()
	}

	return newLogger(l.zapLogger.With(zapFields...), l.sanitizer, l.fieldSanitizer)
}

// WithComponent returns a new logger with the given component
func (l *logger) WithComponent(component string) Logger {
	return l.With("component", component)
}

// WithOperation returns a new logger with the given operation
func (l *logger) WithOperation(operation string) Logger {
	return l.With("operation", operation)
}

// WithRequestID returns a new logger with the given request ID
func (l *logger) WithRequestID(requestID string) Logger {
	return l.With("request_id", requestID)
}

// WithUserID returns a new logger with the given user ID
func (l *logger) WithUserID(userID string) Logger {
	return l.With("user_id", l.SanitizeField("user_id", userID))
}

// WithError returns a new logger with the given error
func (l *logger) WithError(err error) Logger {
	return l.With("error", sanitizeError(err, l.sanitizer))
}

// WithFields adds multiple fields to the logger.
//
// Deprecated: Use With(fields ...any) with key-value pairs instead.
// This method will be removed in a future version.
//
// Example migration:
//
//	// Old:
//	logger.WithFields(map[string]any{"form_id": id, "user_id": uid})
//
//	// New:
//	logger.With("form_id", id, "user_id", uid)
func (l *logger) WithFields(fields map[string]any) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.String(k, l.SanitizeField(k, v)))
	}

	return newLogger(l.zapLogger.With(zapFields...), l.sanitizer, l.fieldSanitizer)
}

// SanitizeField returns a masked version of a sensitive field value
func (l *logger) SanitizeField(key string, value any) string {
	return l.fieldSanitizer.SanitizeField(key, value)
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...any) {
	l.zapLogger.Debug(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// DebugWithFields logs a debug message with Field-based API
func (l *logger) DebugWithFields(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = field.ToZapField()
	}

	l.zapLogger.Debug(msg, zapFields...)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...any) {
	l.zapLogger.Info(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// InfoWithFields logs an info message with Field-based API
func (l *logger) InfoWithFields(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = field.ToZapField()
	}

	l.zapLogger.Info(msg, zapFields...)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...any) {
	l.zapLogger.Warn(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// WarnWithFields logs a warning message with Field-based API
func (l *logger) WarnWithFields(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = field.ToZapField()
	}

	l.zapLogger.Warn(msg, zapFields...)
}

// Error logs an error message
func (l *logger) Error(msg string, fields ...any) {
	l.zapLogger.Error(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// ErrorWithFields logs an error message with Field-based API
func (l *logger) ErrorWithFields(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = field.ToZapField()
	}

	l.zapLogger.Error(msg, zapFields...)
}

// Fatal logs a fatal message
func (l *logger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// FatalWithFields logs a fatal message with Field-based API
func (l *logger) FatalWithFields(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = field.ToZapField()
	}

	l.zapLogger.Fatal(msg, zapFields...)
}
