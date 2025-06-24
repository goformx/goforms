package logging_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
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
	mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
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
	mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
		return input
	}).AnyTimes()

	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test-app",
		Version:     "1.0.0",
		Environment: "development",
	}, mockSanitizer)
	_, err := factory.CreateLogger()
	require.NoError(t, err)

	// Test the new field wrappers instead of the old SanitizeField method
	t.Run("sensitive field - password", func(t *testing.T) {
		field := logging.Sensitive("password", "secret123")
		assert.Equal(t, "password", field.Key)
		assert.Equal(t, "****", field.String)
	})

	t.Run("sensitive field - token", func(t *testing.T) {
		field := logging.Sensitive("token", "abc123")
		assert.Equal(t, "token", field.Key)
		assert.Equal(t, "****", field.String)
	})

	t.Run("non-sensitive field", func(t *testing.T) {
		field := logging.Sensitive("name", "John Doe")
		assert.Equal(t, "name", field.Key)
		assert.Equal(t, "John Doe", field.String)
	})

	t.Run("path field - valid", func(t *testing.T) {
		field := logging.Path("path", "/api/v1/users")
		assert.Equal(t, "path", field.Key)
		assert.Equal(t, "/api/v1/users", field.String)
	})

	t.Run("path field - invalid", func(t *testing.T) {
		field := logging.Path("path", "/api/v1/users<script>alert(1)</script>")
		assert.Equal(t, "path", field.Key)
		assert.Equal(t, "[invalid path]", field.String)
	})

	t.Run("long string field", func(t *testing.T) {
		longString := strings.Repeat("a", 2000)
		field := logging.TruncatedField("description", longString, 1000)
		assert.Equal(t, "description", field.Key)
		assert.Equal(t, strings.Repeat("a", 1000)+"...", field.String)
	})
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
	mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
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
		"user_id", "550e8400-e29b-41d4-a716-446655440000", // Use a valid UUID
		"path", "/api/test",
	)

	// Parse the JSON output
	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Verify the output
	assert.Equal(t, "operation failed", output["msg"])
	assert.Equal(t, "test error", output["error"])
	assert.Equal(t, "550e8400...0000", output["user_id"]) // UUID should be masked
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
	mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
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
	mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
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

func TestLogger_LogInjectionProtection(t *testing.T) {
	factory := logging.NewFactory(logging.FactoryConfig{
		AppName:     "test",
		Version:     "1.0.0",
		Environment: "development",
	}, sanitization.NewService())

	// Test malicious inputs that could be used for log injection
	maliciousInputs := []struct {
		name     string
		input    string
		expected string // What we expect after sanitization
	}{
		{
			name:     "newline injection",
			input:    "normal message\nmalicious log entry",
			expected: "normal message malicious log entry",
		},
		{
			name:     "carriage return injection",
			input:    "normal message\rmalicious log entry",
			expected: "normal message malicious log entry",
		},
		{
			name:     "null byte injection",
			input:    "normal message\x00malicious log entry",
			expected: "normal messagemalicious log entry",
		},
		{
			name:     "HTML injection",
			input:    "normal message<script>alert('xss')</script>",
			expected: "normal message&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "mixed injection",
			input:    "normal\nmessage<script>alert('xss')</script>\r\n",
			expected: "normal message&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
	}

	for _, tt := range maliciousInputs {
		t.Run(tt.name, func(t *testing.T) {
			var captured string
			testCore := zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.AddSync(&testWriter{&captured}),
				zapcore.DebugLevel,
			)

			testLogger, err := factory.WithTestCore(testCore).CreateLogger()
			require.NoError(t, err)

			testLogger.Info(tt.input)

			// The log output will always end with a newline, so allow it
			trimmed := strings.TrimSuffix(captured, "\n")
			assert.Contains(t, trimmed, tt.expected)
			// Only check for newlines in the message part, not the whole log line
			fields := strings.Split(trimmed, "\t")
			if len(fields) > 3 {
				msg := fields[3]
				assert.NotContains(t, msg, "\n")
				assert.NotContains(t, msg, "\r")
				assert.NotContains(t, msg, "\x00")
			}
			assert.NotContains(t, trimmed, "<script>")
		})
	}
}

// testWriter is a simple writer for capturing log output in tests
type testWriter struct {
	content *string
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	*w.content += string(p)
	return len(p), nil
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"valid path", "/api/v1/users", true},
		{"valid path with params", "/api/v1/users?page=1", true},
		{"empty path", "", false},
		{"path without leading slash", "api/v1/users", false},
		{"path with dangerous chars", "/api/v1/users<script>", false},
		{"path with newlines", "/api/v1/users\n", false},
		{"path with null bytes", "/api/v1/users\x00", false},
		{"path traversal attempt", "/api/v1/../etc/passwd", false},
		{"double slash", "/api//v1/users", false},
		{"path too long", "/" + strings.Repeat("a", 501), false}, // MaxPathLength is 500
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to test the validatePath function, but it's not exported
			// So we'll test it indirectly through the logging system
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockSanitizer := mocksanitization.NewMockService(ctrl)
			mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
				return input
			}).AnyTimes()

			factory := logging.NewFactory(logging.FactoryConfig{
				AppName:     "test-app",
				Version:     "1.0.0",
				Environment: "development",
			}, mockSanitizer)
			logger, err := factory.CreateLogger()
			require.NoError(t, err)

			// Test path validation through the logging system
			sanitized := logger.SanitizeField("path", tt.path)
			if tt.expected {
				assert.NotEqual(t, "[invalid path]", sanitized)
			} else {
				assert.Equal(t, "[invalid path]", sanitized)
			}
		})
	}
}

func TestValidateUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{"valid user agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", true},
		{"empty user agent", "", true},
		{"user agent with dangerous chars", "Mozilla<script>", false},
		{"user agent with newlines", "Mozilla\n", false},
		{"user agent with null bytes", "Mozilla\x00", false},
		{"user agent with javascript", "Mozilla javascript:alert(1)", false},
		{"user agent with script tag", "Mozilla <script>alert(1)</script>", false},
		{"user agent with onload", "Mozilla onload=alert(1)", false},
		{"user agent too long", strings.Repeat("a", 1001), false}, // MaxStringLength is 1000
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test user agent validation through the logging system
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockSanitizer := mocksanitization.NewMockService(ctrl)
			mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
				return input
			}).AnyTimes()

			factory := logging.NewFactory(logging.FactoryConfig{
				AppName:     "test-app",
				Version:     "1.0.0",
				Environment: "development",
			}, mockSanitizer)
			logger, err := factory.CreateLogger()
			require.NoError(t, err)

			// Test user agent validation through the logging system
			sanitized := logger.SanitizeField("user_agent", tt.userAgent)
			if tt.expected {
				assert.NotEqual(t, "[invalid user agent]", sanitized)
			} else {
				assert.Equal(t, "[invalid user agent]", sanitized)
			}
		})
	}
}

func TestSanitizeValueWithUserAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSanitizer := mocksanitization.NewMockService(ctrl)
	mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
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
		{"valid user agent", "user_agent", "Mozilla/5.0", "Mozilla/5.0"},
		{"invalid user agent", "user_agent", "Mozilla<script>", "[invalid user agent]"},
		{"user agent with newlines", "user_agent", "Mozilla\n", "[invalid user agent]"},
		{"non-string user agent", "user_agent", 123, "[invalid user agent type]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.SanitizeField(tt.key, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeValueWithUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSanitizer := mocksanitization.NewMockService(ctrl)
	mockSanitizer.EXPECT().SanitizeForLogging(gomock.Any()).DoAndReturn(func(input string) string {
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
		{"sensitive form_id should be masked", "form_id", "550e8400-e29b-41d4-a716-446655440000", "****"},
		{"valid user_id (not sensitive)", "user_id", "550e8400-e29b-41d4-a716-446655440001", "550e8400...0001"},
		{"invalid uuid format for user_id", "user_id", "invalid-uuid", "[invalid uuid format]"},
		{"uuid too short for user_id", "user_id", "550e8400", "[invalid uuid format]"},
		{"non-string uuid for user_id", "user_id", 123, "[invalid uuid type]"},
		{"length field should not be masked", "id_length", "36", "36"},
		{"other id field should be masked", "other_id", "550e8400-e29b-41d4-a716-446655440002", "550e8400...0002"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.SanitizeField(tt.key, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}
