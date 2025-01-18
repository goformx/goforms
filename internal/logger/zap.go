package logger

import (
	"time"

	forbidden_zap "go.uber.org/zap"
)

type zapLogger struct {
	*forbidden_zap.Logger
}

func newZapLogger(l *forbidden_zap.Logger) Logger {
	return &zapLogger{l}
}

// UnderlyingZap returns the underlying zap logger
func UnderlyingZap(l Logger) *forbidden_zap.Logger {
	if zl, ok := l.(*zapLogger); ok {
		return zl.Logger
	}
	return forbidden_zap.NewNop()
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

func toZapFields(fields []Field) []forbidden_zap.Field {
	zapFields := make([]forbidden_zap.Field, len(fields))
	for i, field := range fields {
		if zf, ok := field.(forbidden_zap.Field); ok {
			zapFields[i] = zf
		}
	}
	return zapFields
}

func zapString(key string, value string) Field {
	return forbidden_zap.String(key, value)
}

func zapInt(key string, value int) Field {
	return forbidden_zap.Int(key, value)
}

func zapDuration(key string, value interface{}) Field {
	switch v := value.(type) {
	case time.Duration:
		return forbidden_zap.Duration(key, v)
	default:
		return forbidden_zap.Any(key, v)
	}
}

func zapError(err error) Field {
	return forbidden_zap.Error(err)
}
