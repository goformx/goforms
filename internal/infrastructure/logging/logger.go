package logging

import (
	"go.uber.org/zap"

	loggingsanitization "github.com/goformx/goforms/internal/infrastructure/logging/sanitization"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// logger implements the Logger interface using zap
type logger struct {
	zapLogger      *zap.Logger
	sanitizer      sanitization.ServiceInterface
	fieldSanitizer *loggingsanitization.FieldSanitizer
}

// newLogger creates a new logger instance
func newLogger(
	zapLogger *zap.Logger,
	sanitizer sanitization.ServiceInterface,
	fieldSanitizer *loggingsanitization.FieldSanitizer,
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

// WithFields adds multiple fields to the logger
func (l *logger) WithFields(fields map[string]any) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.String(k, l.SanitizeField(k, v)))
	}
	return newLogger(l.zapLogger.With(zapFields...), l.sanitizer, l.fieldSanitizer)
}

// SanitizeField returns a masked version of a sensitive field value
func (l *logger) SanitizeField(key string, value any) string {
	return l.fieldSanitizer.Sanitize(key, value, l.sanitizer)
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...any) {
	l.zapLogger.Debug(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...any) {
	l.zapLogger.Info(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...any) {
	l.zapLogger.Warn(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// Error logs an error message
func (l *logger) Error(msg string, fields ...any) {
	l.zapLogger.Error(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}

// Fatal logs a fatal message
func (l *logger) Fatal(msg string, fields ...any) {
	l.zapLogger.Fatal(msg, convertToZapFields(fields, l.fieldSanitizer)...)
}
