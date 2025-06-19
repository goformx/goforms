package logging_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	mocksanitization "github.com/goformx/goforms/test/mocks/sanitization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSanitizer := mocksanitization.NewMockService(ctrl)
	// Configure mock expectations - use AnyTimes() for all possible calls
	mockSanitizer.EXPECT().SingleLine(gomock.Any()).DoAndReturn(func(input string) string {
		return input
	}).AnyTimes()

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
			factory := logging.NewFactory(tt.config, mockSanitizer).WithTestCore(core)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSanitizer := mocksanitization.NewMockService(ctrl)
	// Configure mock expectations for sanitization tests - use AnyTimes() for all possible calls
	mockSanitizer.EXPECT().SingleLine(gomock.Any()).DoAndReturn(func(input string) string {
		return input
	}).AnyTimes()

	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	}, mockSanitizer)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSanitizer := mocksanitization.NewMockService(ctrl)
	// Configure mock expectations - use AnyTimes() for all possible calls
	mockSanitizer.EXPECT().SingleLine(gomock.Any()).DoAndReturn(func(input string) string {
		return input
	}).AnyTimes()

	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	}, mockSanitizer).WithTestCore(core)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSanitizer := mocksanitization.NewMockService(ctrl)
	// Configure mock expectations - use AnyTimes() for all possible calls
	mockSanitizer.EXPECT().SingleLine(gomock.Any()).DoAndReturn(func(input string) string {
		return input
	}).AnyTimes()

	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	}, mockSanitizer).WithTestCore(core)
	logger, err := factory.CreateLogger()
	require.NoError(t, err)

	// Test WithComponent
	componentLogger := logger.WithComponent("test-component")
	componentLogger.Info("component message")

	// Parse the JSON output
	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Verify the output
	assert.Equal(t, "component message", output["msg"])
	assert.Equal(t, "test-component", output["component"])
}

func TestLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSanitizer := mocksanitization.NewMockService(ctrl)
	// Configure mock expectations - use AnyTimes() for all possible calls
	mockSanitizer.EXPECT().SingleLine(gomock.Any()).DoAndReturn(func(input string) string {
		return input
	}).AnyTimes()

	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	}, mockSanitizer).WithTestCore(core)
	logger, err := factory.CreateLogger()
	require.NoError(t, err)

	// Test different log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Verify all messages are logged
	output := buf.String()
	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}
