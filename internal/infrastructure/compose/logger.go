package compose

import (
	"io"

	infralogging "github.com/goformx/goforms/internal/infrastructure/logging"
)

// LoggerAdapter adapts the infrastructure logger to the compose Logger interface.
type LoggerAdapter struct {
	logger infralogging.Logger
}

// NewLoggerAdapter creates a new logger adapter from an infrastructure logger.
func NewLoggerAdapter(logger infralogging.Logger) Logger {
	if logger == nil {
		return &NullLogger{}
	}
	return &LoggerAdapter{logger: logger}
}

// Info logs an info message.
func (l *LoggerAdapter) Info(msg string) {
	l.logger.Info(msg, "component", "compose")
}

// Warn logs a warning message.
func (l *LoggerAdapter) Warn(msg string) {
	l.logger.Warn(msg, "component", "compose")
}

// Error logs an error message.
func (l *LoggerAdapter) Error(msg string) {
	l.logger.Error(msg, "component", "compose")
}

// Debug logs a debug message.
func (l *LoggerAdapter) Debug(msg string) {
	l.logger.Debug(msg, "component", "compose")
}

// NullLogger is a no-op logger implementation.
type NullLogger struct{}

// Info is a no-op.
func (n *NullLogger) Info(msg string) {}

// Warn is a no-op.
func (n *NullLogger) Warn(msg string) {}

// Error is a no-op.
func (n *NullLogger) Error(msg string) {}

// Debug is a no-op.
func (n *NullLogger) Debug(msg string) {}

// SimpleLogger is a basic logger that writes to stdout/stderr.
type SimpleLogger struct {
	infoWriter  io.Writer
	warnWriter  io.Writer
	errorWriter io.Writer
	debugWriter io.Writer
}

// NewSimpleLogger creates a simple logger that writes to the given writers.
func NewSimpleLogger(info, warn, err, debug io.Writer) Logger {
	if info == nil {
		info = io.Discard
	}
	if warn == nil {
		warn = io.Discard
	}
	if err == nil {
		err = io.Discard
	}
	if debug == nil {
		debug = io.Discard
	}
	return &SimpleLogger{
		infoWriter:  info,
		warnWriter:  warn,
		errorWriter: err,
		debugWriter: debug,
	}
}

// Info writes an info message.
func (s *SimpleLogger) Info(msg string) {
	_, _ = s.infoWriter.Write([]byte("[INFO] " + msg + "\n"))
}

// Warn writes a warning message.
func (s *SimpleLogger) Warn(msg string) {
	_, _ = s.warnWriter.Write([]byte("[WARN] " + msg + "\n"))
}

// Error writes an error message.
func (s *SimpleLogger) Error(msg string) {
	_, _ = s.errorWriter.Write([]byte("[ERROR] " + msg + "\n"))
}

// Debug writes a debug message.
func (s *SimpleLogger) Debug(msg string) {
	_, _ = s.debugWriter.Write([]byte("[DEBUG] " + msg + "\n"))
}
