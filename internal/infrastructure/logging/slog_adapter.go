// Package logging provides slog adapter for stdlib compatibility.
package logging

import (
	"context"
	"log/slog"
	"os"

	"go.uber.org/zap"
)

// SlogAdapter wraps our Logger interface to provide an slog.Handler
type SlogAdapter struct {
	zapLogger *zap.Logger
	attrs     []slog.Attr
	groups    []string
}

// NewSlogHandler creates a new slog.Handler from a zap logger
func NewSlogHandler(zapLogger *zap.Logger) slog.Handler {
	return &SlogAdapter{
		zapLogger: zapLogger,
		attrs:     make([]slog.Attr, 0),
		groups:    make([]string, 0),
	}
}

// Enabled returns true if the level is enabled
func (s *SlogAdapter) Enabled(_ context.Context, level slog.Level) bool {
	return s.zapLevelEnabled(level)
}

// Handle handles the log record
func (s *SlogAdapter) Handle(_ context.Context, record slog.Record) error {
	fields := s.buildFields(record)

	switch record.Level {
	case slog.LevelDebug:
		s.zapLogger.Debug(record.Message, fields...)
	case slog.LevelInfo:
		s.zapLogger.Info(record.Message, fields...)
	case slog.LevelWarn:
		s.zapLogger.Warn(record.Message, fields...)
	case slog.LevelError:
		s.zapLogger.Error(record.Message, fields...)
	default:
		s.zapLogger.Info(record.Message, fields...)
	}

	return nil
}

// WithAttrs returns a new handler with the given attributes
func (s *SlogAdapter) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(s.attrs)+len(attrs))
	copy(newAttrs, s.attrs)
	copy(newAttrs[len(s.attrs):], attrs)

	return &SlogAdapter{
		zapLogger: s.zapLogger,
		attrs:     newAttrs,
		groups:    s.groups,
	}
}

// WithGroup returns a new handler with the given group
func (s *SlogAdapter) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(s.groups)+1)
	copy(newGroups, s.groups)
	newGroups[len(s.groups)] = name

	return &SlogAdapter{
		zapLogger: s.zapLogger,
		attrs:     s.attrs,
		groups:    newGroups,
	}
}

// buildFields converts slog attributes to zap fields
func (s *SlogAdapter) buildFields(record slog.Record) []zap.Field {
	fields := make([]zap.Field, 0, record.NumAttrs()+len(s.attrs))

	// Add pre-existing attrs
	for _, attr := range s.attrs {
		fields = append(fields, s.attrToZapField(attr))
	}

	// Add record attrs
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, s.attrToZapField(attr))
		return true
	})

	return fields
}

// attrToZapField converts an slog.Attr to a zap.Field
func (s *SlogAdapter) attrToZapField(attr slog.Attr) zap.Field {
	key := attr.Key
	if len(s.groups) > 0 {
		for _, g := range s.groups {
			key = g + "." + key
		}
	}

	switch attr.Value.Kind() {
	case slog.KindString:
		return zap.String(key, attr.Value.String())
	case slog.KindInt64:
		return zap.Int64(key, attr.Value.Int64())
	case slog.KindUint64:
		return zap.Uint64(key, attr.Value.Uint64())
	case slog.KindFloat64:
		return zap.Float64(key, attr.Value.Float64())
	case slog.KindBool:
		return zap.Bool(key, attr.Value.Bool())
	case slog.KindTime:
		return zap.Time(key, attr.Value.Time())
	case slog.KindDuration:
		return zap.Duration(key, attr.Value.Duration())
	case slog.KindGroup:
		// For groups, we flatten the attributes
		groupAttrs := attr.Value.Group()
		if len(groupAttrs) == 1 {
			return s.attrToZapField(slog.Attr{Key: key + "." + groupAttrs[0].Key, Value: groupAttrs[0].Value})
		}
		return zap.Any(key, attr.Value.Any())
	case slog.KindAny:
		return zap.Any(key, attr.Value.Any())
	case slog.KindLogValuer:
		return zap.Any(key, attr.Value.Any())
	default:
		return zap.Any(key, attr.Value.Any())
	}
}

// zapLevelEnabled checks if a level is enabled in the zap logger
func (s *SlogAdapter) zapLevelEnabled(level slog.Level) bool {
	// Map slog levels to zap levels and check
	zapCore := s.zapLogger.Core()
	switch level {
	case slog.LevelDebug:
		return zapCore.Enabled(zap.DebugLevel)
	case slog.LevelInfo:
		return zapCore.Enabled(zap.InfoLevel)
	case slog.LevelWarn:
		return zapCore.Enabled(zap.WarnLevel)
	case slog.LevelError:
		return zapCore.Enabled(zap.ErrorLevel)
	default:
		return true
	}
}

// SetDefaultSlogLogger sets the default slog logger to use our adapter
func SetDefaultSlogLogger(zapLogger *zap.Logger) {
	handler := NewSlogHandler(zapLogger)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// NewSlogLogger creates a new slog.Logger using our adapter
func NewSlogLogger(zapLogger *zap.Logger) *slog.Logger {
	return slog.New(NewSlogHandler(zapLogger))
}

// SlogLoggerFromFactory creates an slog.Logger from a logging Factory
func SlogLoggerFromFactory(factory *Factory) (*slog.Logger, error) {
	loggerImpl, err := factory.CreateLogger()
	if err != nil {
		return nil, err
	}

	// Get the underlying zap logger from the factory
	zapLogger := factory.GetZapLogger()
	if zapLogger != nil {
		return NewSlogLogger(zapLogger), nil
	}

	// If we can't get the zap logger, use the interface methods
	// This provides basic compatibility but loses some zap features
	_ = loggerImpl // Use the logger interface if needed
	return slog.New(slog.NewJSONHandler(os.Stdout, nil)), nil
}
