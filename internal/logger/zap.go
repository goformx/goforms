package logger

import (
	"time"

	"go.uber.org/zap"
)

type zapLogger struct {
	*zap.Logger
}

func newZapLogger(l *zap.Logger) Logger {
	return &zapLogger{l}
}

// UnderlyingZap returns the underlying zap logger
func UnderlyingZap(l Logger) *zap.Logger {
	if zl, ok := l.(*zapLogger); ok {
		return zl.Logger
	}
	return zap.NewNop()
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.Logger.Info(msg, toZapFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.Logger.Error(msg, toZapFields(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.Logger.Warn(msg, toZapFields(fields)...)
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.Logger.Debug(msg, toZapFields(fields)...)
}

func toZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		if zf, ok := field.(zap.Field); ok {
			zapFields[i] = zf
		}
	}
	return zapFields
}

func zapString(key string, value string) Field {
	return zap.String(key, value)
}

func zapInt(key string, value int) Field {
	return zap.Int(key, value)
}

func zapDuration(key string, value interface{}) Field {
	switch v := value.(type) {
	case time.Duration:
		return zap.Duration(key, v)
	default:
		return zap.Any(key, v)
	}
}

func zapError(err error) Field {
	return zap.Error(err)
}
