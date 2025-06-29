// Package logging provides a unified logging interface
package logging

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
)

// setDefaultPaths sets default output paths if not specified
func setDefaultPaths(cfg *FactoryConfig) {
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}

	if len(cfg.ErrorOutputPaths) == 0 {
		cfg.ErrorOutputPaths = []string{"stderr"}
	}
}

// Validate validates the factory configuration
func (cfg *FactoryConfig) Validate() error {
	// Validate required fields
	if err := cfg.validateRequiredFields(); err != nil {
		return err
	}

	// Validate log level
	if err := cfg.validateLogLevel(); err != nil {
		return err
	}

	// Validate output paths
	if err := cfg.validateOutputPaths(); err != nil {
		return err
	}

	return nil
}

// validateRequiredFields validates required configuration fields
func (cfg *FactoryConfig) validateRequiredFields() error {
	if cfg.AppName == "" {
		return errors.New("app name is required")
	}

	if cfg.Environment == "" {
		cfg.Environment = "production"
	}

	return nil
}

// validateLogLevel validates the log level configuration
func (cfg *FactoryConfig) validateLogLevel() error {
	if cfg.LogLevel != "" && !isValidLogLevel(cfg.LogLevel) {
		return fmt.Errorf("invalid log level: %s", cfg.LogLevel)
	}

	return nil
}

// validateOutputPaths validates output path configurations
func (cfg *FactoryConfig) validateOutputPaths() error {
	// Validate regular output paths
	for _, path := range cfg.OutputPaths {
		if !isValidOutputPath(path) {
			return fmt.Errorf("invalid output path: %s", path)
		}
	}

	// Validate error output paths
	for _, path := range cfg.ErrorOutputPaths {
		if !isValidOutputPath(path) {
			return fmt.Errorf("invalid error output path: %s", path)
		}
	}

	return nil
}

// isValidOutputPath checks if an output path is valid
func isValidOutputPath(path string) bool {
	return path == "stdout" || path == "stderr" || strings.HasSuffix(path, ".log")
}

// isValidLogLevel checks if the log level is valid
func isValidLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	levelLower := strings.ToLower(level)

	for _, valid := range validLevels {
		if levelLower == valid {
			return true
		}
	}

	return false
}

// parseLogLevel converts string level to zap level
func parseLogLevel(level, environment string) zapcore.Level {
	if level != "" {
		switch strings.ToLower(level) {
		case "debug":
			return zapcore.DebugLevel
		case "info":
			return zapcore.InfoLevel
		case "warn":
			return zapcore.WarnLevel
		case string(LogLevelError):
			return zapcore.ErrorLevel
		case "fatal":
			return zapcore.FatalLevel
		default:
			return zapcore.InfoLevel
		}
	}

	// Fallback to environment-based level
	switch strings.ToLower(environment) {
	case "development":
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
}

// createEncoderConfig creates the zap encoder configuration
func createEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			// Show only the last two parts of the file path
			parts := strings.Split(caller.File, "/")
			if len(parts) > MaxPartsLength {
				parts = parts[len(parts)-MaxPartsLength:]
			}
			file := strings.Join(parts, "/")
			enc.AppendString(fmt.Sprintf("%s:%d", file, caller.Line))
		},
	}
}

// createZapCore creates the zap core with the appropriate level and output
func createZapCore(level zapcore.Level, testCore zapcore.Core) zapcore.Core {
	if testCore != nil {
		return testCore
	}

	encoder := zapcore.NewConsoleEncoder(createEncoderConfig())

	return zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)
}

// createJSONEncoder creates a JSON encoder for production environments
func createJSONEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

// createProductionCore creates a production-optimized zap core with stdout output
func createProductionCore(level zapcore.Level) zapcore.Core {
	encoder := createJSONEncoder()
	writeSyncer := zapcore.AddSync(os.Stdout)

	return zapcore.NewCore(encoder, writeSyncer, level)
}
