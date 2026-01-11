// Package adapters provides adapter implementations for external interfaces.
package adapters

import (
	"fmt"
	"io"
	"os"

	"github.com/labstack/gommon/log"

	"github.com/goformx/goforms/internal/application/constants"
	"github.com/goformx/goforms/internal/infrastructure/logging"
)

// EchoLogger implements echo.Logger interface using our custom logger
type EchoLogger struct {
	logger logging.Logger
}

// NewEchoLogger creates a new EchoLogger adapter
func NewEchoLogger(logger logging.Logger) *EchoLogger {
	return &EchoLogger{logger: logger}
}

// Print logs a message at info level
func (l *EchoLogger) Print(i ...any) {
	l.logger.Info(fmt.Sprint(i...))
}

// Printf logs a formatted message at info level
func (l *EchoLogger) Printf(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Printj logs a JSON message at info level
func (l *EchoLogger) Printj(j log.JSON) {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}

	l.logger.Info("", fields...)
}

// Debug logs a message at debug level
func (l *EchoLogger) Debug(i ...any) {
	if len(i) > 1 {
		message := fmt.Sprint(i[0])
		fields := i[1:]
		l.logger.Debug(message, fields...)
	} else {
		l.logger.Debug(fmt.Sprint(i...))
	}
}

// Debugf logs a formatted message at debug level
func (l *EchoLogger) Debugf(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

// Debugj logs a JSON message at debug level
func (l *EchoLogger) Debugj(j log.JSON) {
	l.logger.Debug("", l.jsonToFields(j)...)
}

// Info logs a message at info level
func (l *EchoLogger) Info(i ...any) {
	if len(i) > 1 {
		message := fmt.Sprint(i[0])
		fields := i[1:]
		l.logger.Info(message, fields...)
	} else {
		l.logger.Info(fmt.Sprint(i...))
	}
}

// Infof logs a formatted message at info level
func (l *EchoLogger) Infof(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Infoj logs a JSON message at info level
func (l *EchoLogger) Infoj(j log.JSON) {
	l.logger.Info("", l.jsonToFields(j)...)
}

// Warn logs a message at warn level
func (l *EchoLogger) Warn(i ...any) {
	if len(i) > 1 {
		message := fmt.Sprint(i[0])
		fields := i[1:]
		l.logger.Warn(message, fields...)
	} else {
		l.logger.Warn(fmt.Sprint(i...))
	}
}

// Warnf logs a formatted message at warn level
func (l *EchoLogger) Warnf(format string, args ...any) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

// Warnj logs a JSON message at warn level
func (l *EchoLogger) Warnj(j log.JSON) {
	l.logger.Warn("", l.jsonToFields(j)...)
}

// Error logs a message at error level
func (l *EchoLogger) Error(i ...any) {
	if len(i) > 1 {
		message := fmt.Sprint(i[0])
		fields := i[1:]
		l.logger.Error(message, fields...)
	} else {
		l.logger.Error(fmt.Sprint(i...))
	}
}

// Errorf logs a formatted message at error level
func (l *EchoLogger) Errorf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

// Errorj logs a JSON message at error level
func (l *EchoLogger) Errorj(j log.JSON) {
	l.logger.Error("", l.jsonToFields(j)...)
}

// Fatal logs a message at fatal level
func (l *EchoLogger) Fatal(i ...any) {
	l.logger.Fatal(fmt.Sprint(i...))
}

// Fatalf logs a formatted message at fatal level
func (l *EchoLogger) Fatalf(format string, args ...any) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

// Fatalj logs a JSON message at fatal level
func (l *EchoLogger) Fatalj(j log.JSON) {
	l.logger.Fatal("", l.jsonToFields(j)...)
}

// Panic logs a message at error level and panics
func (l *EchoLogger) Panic(i ...any) {
	l.logger.Error(fmt.Sprint(i...))
	panic(fmt.Sprint(i...))
}

// Panicf logs a formatted message at error level and panics
func (l *EchoLogger) Panicf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
	panic(fmt.Sprintf(format, args...))
}

// Panicj logs a JSON message at error level and panics
func (l *EchoLogger) Panicj(j log.JSON) {
	l.logger.Error("", l.jsonToFields(j)...)
	panic(fmt.Sprintf("%v", j))
}

// Level returns the current log level
func (l *EchoLogger) Level() log.Lvl {
	return log.INFO
}

// SetLevel sets the log level (no-op as we use our own configuration)
func (l *EchoLogger) SetLevel(_ log.Lvl) {}

// SetHeader sets the log header (no-op as we use our own format)
func (l *EchoLogger) SetHeader(_ string) {}

// SetPrefix sets the log prefix (no-op as we use our own format)
func (l *EchoLogger) SetPrefix(_ string) {}

// Prefix returns the current log prefix
func (l *EchoLogger) Prefix() string {
	return ""
}

// SetOutput sets the log output (no-op as we use our own configuration)
func (l *EchoLogger) SetOutput(_ io.Writer) {}

// Output returns the current log output writer
func (l *EchoLogger) Output() io.Writer {
	return os.Stdout
}

// jsonToFields converts log.JSON to a slice of key-value pairs
func (l *EchoLogger) jsonToFields(j log.JSON) []any {
	fields := make([]any, 0, len(j)*constants.FieldPairSize)
	for k, v := range j {
		fields = append(fields, k, fmt.Sprint(v))
	}
	return fields
}
