// Package logging provides a unified logging interface using zap
package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger implements the Logger interface using zap
// (Logger interface is defined in types.go)
type logger struct {
	log *zap.Logger
}

type noopLogger struct{}

func NewNoopLogger() Logger {
	return &noopLogger{}
}

func NewLogger(logLevel, appName string) (Logger, error) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var level zapcore.Level
	levelErr := level.UnmarshalText([]byte(logLevel))
	if levelErr != nil {
		level = zapcore.InfoLevel
	}

	config := zap.NewDevelopmentConfig()
	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string{"stdout"}
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(level)

	zapLog, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("app", appName)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return &logger{log: zapLog}, nil
}

func NewTestLogger() (Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	zapLog, err := config.Build(zap.Fields(zap.String("app", "test")))
	if err != nil {
		return nil, fmt.Errorf("failed to create test logger: %w", err)
	}
	return &logger{log: zapLog}, nil
}

func NewZapLogger(zapLog *zap.Logger) Logger {
	return &logger{log: zapLog}
}

func (l *logger) Debug(msg string, fields ...any) { l.log.Debug(msg, convertToZapFields(fields)...) }
func (l *logger) Info(msg string, fields ...any)  { l.log.Info(msg, convertToZapFields(fields)...) }
func (l *logger) Warn(msg string, fields ...any)  { l.log.Warn(msg, convertToZapFields(fields)...) }
func (l *logger) Error(msg string, fields ...any) { l.log.Error(msg, convertToZapFields(fields)...) }
func (l *logger) Fatal(msg string, fields ...any) { l.log.Fatal(msg, convertToZapFields(fields)...) }

func (l *logger) With(fields ...any) Logger {
	return &logger{log: l.log.With(convertToZapFields(fields)...)}
}

func (l *noopLogger) Debug(msg string, fields ...any) {}
func (l *noopLogger) Info(msg string, fields ...any)  {}
func (l *noopLogger) Warn(msg string, fields ...any)  {}
func (l *noopLogger) Error(msg string, fields ...any) {}
func (l *noopLogger) Fatal(msg string, fields ...any) {}
func (l *noopLogger) With(fields ...any) Logger       { return l }
