package logging_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestFactory_CreateLogger(t *testing.T) {
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)

	tests := []struct {
		name      string
		config    logging.FactoryConfig
		wantErr   bool
		checkFunc func(t *testing.T, logger logging.Logger)
	}{
		{
			name: "successful logger creation with default config",
			config: logging.FactoryConfig{
				AppName:     "test-app",
				Version:     "1.0.0",
				Environment: "development",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logger logging.Logger) {
				// Test basic logging
				logger.Info("test message", "key", "value")
				assert.Contains(t, buf.String(), "test message")
			},
		},
		{
			name: "logger with initial fields",
			config: logging.FactoryConfig{
				AppName:     "test-app",
				Version:     "1.0.0",
				Environment: "development",
				Fields: map[string]any{
					"service": "test-service",
					"region":  "test-region",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, logger logging.Logger) {
				// Test that initial fields are included
				logger.Info("test message")
				var output map[string]any
				err := json.Unmarshal(buf.Bytes(), &output)
				require.NoError(t, err)
				assert.Equal(t, "test-service", output["service"])
				assert.Equal(t, "test-region", output["region"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			factory := logging.NewFactory(tt.config).WithTestCore(core)
			logger, err := factory.CreateLogger()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, logger)
			if tt.checkFunc != nil {
				tt.checkFunc(t, logger)
			}
		})
	}
}

func TestLogger_Sanitization(t *testing.T) {
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	})
	logger, err := factory.CreateLogger()
	require.NoError(t, err)

	tests := []struct {
		name     string
		key      string
		value    any
		expected string
	}{
		{
			name:     "sensitive field - password",
			key:      "password",
			value:    "secret123",
			expected: "****",
		},
		{
			name:     "sensitive field - token",
			key:      "token",
			value:    "abc123",
			expected: "****",
		},
		{
			name:     "non-sensitive field",
			key:      "name",
			value:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "path field - valid",
			key:      "path",
			value:    "/api/v1/users",
			expected: "/api/v1/users",
		},
		{
			name:     "path field - invalid",
			key:      "path",
			value:    "/api/v1/users<script>alert(1)</script>",
			expected: "[invalid path]",
		},
		{
			name:     "long string field",
			key:      "description",
			value:    strings.Repeat("a", 2000),
			expected: strings.Repeat("a", 1000) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := logger.SanitizeField(tt.key, tt.value)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}

func TestLogger_ErrorHandling(t *testing.T) {
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	}).WithTestCore(core)
	logger, err := factory.CreateLogger()
	require.NoError(t, err)

	// Test error logging with context
	err = errors.New("test error")
	logger.Error("operation failed",
		"error", err,
		"user_id", "123",
		"path", "/api/test",
	)

	// Parse the JSON output
	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Verify the output
	assert.Equal(t, "operation failed", output["msg"])
	assert.Equal(t, "test error", output["error"])
	assert.Equal(t, "123", output["user_id"])
	assert.Equal(t, "/api/test", output["path"])
}

func TestLogger_WithMethods(t *testing.T) {
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	}).WithTestCore(core)
	logger, err := factory.CreateLogger()
	require.NoError(t, err)

	// Test WithComponent
	componentLogger := logger.WithComponent("test-component")
	componentLogger.Info("component message")

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, "test-component", output["component"])

	// Test WithOperation
	buf.Reset()
	operationLogger := logger.WithOperation("test-operation")
	operationLogger.Info("operation message")

	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, "test-operation", output["operation"])

	// Test WithRequestID
	buf.Reset()
	requestLogger := logger.WithRequestID("req-123")
	requestLogger.Info("request message")

	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, "req-123", output["request_id"])

	// Test WithUserID
	buf.Reset()
	userLogger := logger.WithUserID("user-123")
	userLogger.Info("user message")

	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, "user-123", output["user_id"])

	// Test WithError
	buf.Reset()
	errLogger := logger.WithError(errors.New("test error"))
	errLogger.Info("error message")

	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, "test error", output["error"])
}

func TestLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	}).WithTestCore(core)
	logger, err := factory.CreateLogger()
	require.NoError(t, err)

	// Test all log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")

	// Split the output into lines
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 4)

	// Verify each log level
	var output map[string]any
	err = json.Unmarshal([]byte(lines[0]), &output)
	require.NoError(t, err)
	assert.Equal(t, "debug message", output["msg"])
	assert.Equal(t, "debug", output["level"])

	err = json.Unmarshal([]byte(lines[1]), &output)
	require.NoError(t, err)
	assert.Equal(t, "info message", output["msg"])
	assert.Equal(t, "info", output["level"])

	err = json.Unmarshal([]byte(lines[2]), &output)
	require.NoError(t, err)
	assert.Equal(t, "warning message", output["msg"])
	assert.Equal(t, "warn", output["level"])

	err = json.Unmarshal([]byte(lines[3]), &output)
	require.NoError(t, err)
	assert.Equal(t, "error message", output["msg"])
	assert.Equal(t, "error", output["level"])
}
