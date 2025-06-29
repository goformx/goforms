package logging

import (
	"errors"
	"fmt"
	"time"

	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

// Example demonstrates how to use the improved logging system
func Example() {
	// Create a sanitizer
	sanitizer := sanitization.NewService()

	// Create factory configuration
	cfg := &FactoryConfig{
		AppName:     "MyApp",
		Version:     "1.0.0",
		Environment: "development",
		LogLevel:    "debug",
		Fields: map[string]any{
			"service": "user-service",
			"version": "1.0.0",
		},
	}

	// Create factory
	factory, err := NewFactory(cfg, sanitizer)
	if err != nil {
		panic(fmt.Sprintf("failed to create factory: %v", err))
	}

	// Create logger
	logger, err := factory.CreateLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	// Example 1: Legacy API with variadic fields
	logger.Info("User logged in", "user_id", "12345", "ip", "192.168.1.1")
	logger.Error("Database connection failed", "error", errors.New("connection timeout"))

	// Example 2: New Field-based API (recommended)
	logger.InfoWithFields("User logged in",
		Int("user_id", 12345),
		String("ip", "192.168.1.1"),
		UUID("session_id", "550e8400-e29b-41d4-a716-446655440000"),
	)

	logger.ErrorWithFields("Database connection failed",
		Error("error", errors.New("connection timeout")),
		String("database", "users_db"),
		Int("retry_count", 3),
	)

	// Example 3: Structured logging with context
	userLogger := logger.WithFieldsStructured(
		Int("user_id", 12345),
		String("action", "profile_update"),
	)

	userLogger.InfoWithFields("Profile updated",
		String("field", "email"),
		String("old_value", "old@example.com"),
		String("new_value", "new@example.com"),
	)

	// Example 4: Sensitive data handling
	logger.InfoWithFields("Payment processed",
		Sensitive("card_number", "4111111111111111"),
		Sensitive("cvv", "123"),
		Float("amount", 99.99),
		String("currency", "USD"),
	)

	// Example 5: Path and user agent sanitization
	logger.InfoWithFields("Request processed",
		Path("request_path", "/api/users/12345"),
		UserAgent("user_agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
		String("method", "GET"),
		Int("status_code", 200),
	)

	// Example 6: Performance logging with timing
	start := time.Now()
	// ... perform some operation ...
	duration := time.Since(start)

	logger.InfoWithFields("Operation completed",
		String("operation", "database_query"),
		Int("duration_ms", int(duration.Milliseconds())),
		Int("rows_returned", 150),
		Bool("cached", false),
	)

	// Example 7: Error logging with context
	err = errors.New("file not found")
	logger.ErrorWithFields("Failed to process file",
		Error("error", err),
		Path("file_path", "/var/log/app.log"),
		String("operation", "read_file"),
		Int("attempt", 1),
	)

	// Example 8: Debug logging with complex objects
	type User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	user := User{
		ID:    12345,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	logger.DebugWithFields("User object created",
		Object("user", user),
		String("source", "database"),
		Bool("cached", true),
	)

	// Example 9: Batch operations
	logger.InfoWithFields("Batch operation started",
		Int("batch_size", 1000),
		String("operation", "user_import"),
		String("source", "csv_file"),
	)

	// Example 10: Security events
	logger.WarnWithFields("Suspicious activity detected",
		String("event_type", "failed_login"),
		String("ip_address", "192.168.1.100"),
		Int("failed_attempts", 5),
		String("user_agent", "Mozilla/5.0 (compatible; Bot)"),
		Bool("blocked", true),
	)
}

// ExampleBenchmark demonstrates performance improvements
func ExampleBenchmark() {
	sanitizer := sanitization.NewService()
	cfg := &FactoryConfig{
		AppName:     "BenchmarkApp",
		Environment: "development",
		LogLevel:    "info",
	}

	factory, err := NewFactory(cfg, sanitizer)
	if err != nil {
		panic(err)
	}

	logger, err := factory.CreateLogger()
	if err != nil {
		panic(err)
	}

	// Benchmark: Legacy API vs New Field-based API
	// The new API preserves native types and is more efficient

	// Legacy API (converts everything to strings)
	for i := 0; i < 1000; i++ {
		logger.Info("Processing item", "id", i, "status", "active", "count", i*2)
	}

	// New Field-based API (preserves types, more efficient)
	for i := 0; i < 1000; i++ {
		logger.InfoWithFields("Processing item",
			Int("id", i),
			String("status", "active"),
			Int("count", i*2),
		)
	}
}

// ExampleConfiguration demonstrates different configuration options
func ExampleConfiguration() {
	// Development configuration
	devConfig := &FactoryConfig{
		AppName:     "MyApp",
		Version:     "1.0.0",
		Environment: "development",
		LogLevel:    "debug",
		OutputPaths: []string{"stdout"},
		Fields: map[string]any{
			"service": "user-service",
			"env":     "development",
		},
	}

	// Production configuration
	prodConfig := &FactoryConfig{
		AppName:          "MyApp",
		Version:          "1.0.0",
		Environment:      "production",
		LogLevel:         "info",
		OutputPaths:      []string{"stdout", "/var/log/app.log"},
		ErrorOutputPaths: []string{"stderr", "/var/log/app-error.log"},
		Fields: map[string]any{
			"service": "user-service",
			"env":     "production",
		},
	}

	// Validate configurations
	if err := devConfig.Validate(); err != nil {
		panic(fmt.Sprintf("invalid dev config: %v", err))
	}

	if err := prodConfig.Validate(); err != nil {
		panic(fmt.Sprintf("invalid prod config: %v", err))
	}

	sanitizer := sanitization.NewService()

	// Create factories
	devFactory, err := NewFactory(devConfig, sanitizer)
	if err != nil {
		panic(fmt.Sprintf("failed to create dev factory: %v", err))
	}
	prodFactory, err := NewFactory(prodConfig, sanitizer)
	if err != nil {
		panic(fmt.Sprintf("failed to create prod factory: %v", err))
	}

	// Create loggers
	devLogger, err := devFactory.CreateLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create dev logger: %v", err))
	}
	prodLogger, err := prodFactory.CreateLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create prod logger: %v", err))
	}

	// Development logger (console output, debug level)
	devLogger.DebugWithFields("Debug information",
		String("component", "auth"),
		Int("attempts", 3),
		Bool("success", true),
	)

	// Production logger (JSON output, info level)
	prodLogger.InfoWithFields("User authenticated",
		Int("user_id", 12345),
		String("method", "password"),
		Int("duration_ms", 150),
	)
}
