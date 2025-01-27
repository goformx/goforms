package config

import (
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Save original env vars
	originalEnv := map[string]string{
		"APP_NAME":    os.Getenv("APP_NAME"),
		"APP_ENV":     os.Getenv("APP_ENV"),
		"APP_DEBUG":   os.Getenv("APP_DEBUG"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"APP_PORT":    os.Getenv("APP_PORT"),
		"APP_HOST":    os.Getenv("APP_HOST"),
	}

	// Cleanup function to restore original env vars
	defer func() {
		for k, v := range originalEnv {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	tests := []struct {
		name      string
		envVars   map[string]string
		wantError bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"APP_NAME":             "testapp",
				"APP_ENV":              "development",
				"APP_DEBUG":            "true",
				"DB_USER":              "testuser",
				"DB_PASSWORD":          "testpass",
				"DB_NAME":              "testdb",
				"DB_HOST":              "localhost",
				"DB_PORT":              "3306",
				"APP_PORT":             "8080",
				"APP_HOST":             "localhost",
				"CORS_ALLOWED_ORIGINS": "http://localhost:3000",
				"CORS_ALLOWED_METHODS": "GET,POST,PUT,DELETE,OPTIONS",
			},
			wantError: false,
		},
		{
			name: "invalid database port",
			envVars: map[string]string{
				"APP_NAME":    "testapp",
				"APP_ENV":     "development",
				"APP_DEBUG":   "true",
				"APP_PORT":    "8080",
				"APP_HOST":    "localhost",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_HOST":     "localhost",
				"DB_PORT":     "invalid", // Invalid port number
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing env vars first
			for k := range originalEnv {
				os.Unsetenv(k)
			}

			// Set environment variables for test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := New()
			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected config but got nil")
			}

			// Verify configuration values
			if tt.name == "valid configuration" {
				if cfg.App.Name != "testapp" {
					t.Errorf("expected App.Name to be %q, got %q", "testapp", cfg.App.Name)
				}
				if cfg.App.Env != "development" {
					t.Errorf("expected App.Env to be %q, got %q", "development", cfg.App.Env)
				}
				if !cfg.App.Debug {
					t.Error("expected App.Debug to be true")
				}
				if cfg.Database.User != "testuser" {
					t.Errorf("expected Database.User to be %q, got %q", "testuser", cfg.Database.User)
				}
				if cfg.Database.Password != "testpass" {
					t.Errorf("expected Database.Password to be %q, got %q", "testpass", cfg.Database.Password)
				}
				if cfg.Database.Name != "testdb" {
					t.Errorf("expected Database.Name to be %q, got %q", "testdb", cfg.Database.Name)
				}
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected Database.Host to be %q, got %q", "localhost", cfg.Database.Host)
				}
				if cfg.Database.Port != 3306 {
					t.Errorf("expected Database.Port to be %d, got %d", 3306, cfg.Database.Port)
				}
				if cfg.App.Port != 8080 {
					t.Errorf("expected App.Port to be %d, got %d", 8080, cfg.App.Port)
				}
				if cfg.App.Host != "localhost" {
					t.Errorf("expected App.Host to be %q, got %q", "localhost", cfg.App.Host)
				}
			}
		})
	}
}

func TestSecurityConfig(t *testing.T) {
	t.Run("default_security_settings", func(t *testing.T) {
		// Set required database config
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")

		// Clean up after test
		defer func() {
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_NAME")
		}()

		config, err := New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		if len(config.Security.CorsAllowedMethods) != len(expectedMethods) {
			t.Errorf("expected %d methods, got %d", len(expectedMethods), len(config.Security.CorsAllowedMethods))
		}
		for i, method := range expectedMethods {
			if config.Security.CorsAllowedMethods[i] != method {
				t.Errorf("expected method %q at index %d, got %q", method, i, config.Security.CorsAllowedMethods[i])
			}
		}
	})

	t.Run("custom_security_settings", func(t *testing.T) {
		// Set required database config
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")

		// Set custom security values
		os.Setenv("CORS_ALLOWED_METHODS", "GET,POST")

		// Clean up after test
		defer func() {
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("CORS_ALLOWED_METHODS")
		}()

		config, err := New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedMethods := []string{"GET", "POST"}
		if len(config.Security.CorsAllowedMethods) != len(expectedMethods) {
			t.Errorf("expected %d methods, got %d", len(expectedMethods), len(config.Security.CorsAllowedMethods))
		}
		for i, method := range expectedMethods {
			if config.Security.CorsAllowedMethods[i] != method {
				t.Errorf("expected method %q at index %d, got %q", method, i, config.Security.CorsAllowedMethods[i])
			}
		}
	})
}

func TestRateLimitConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		check   func(*testing.T, *Config)
	}{
		{
			name: "default rate limit settings",
			envVars: map[string]string{
				// Required database config
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
			},
			check: func(t *testing.T, cfg *Config) {
				if !cfg.RateLimit.Enabled {
					t.Error("expected RateLimit.Enabled to be true")
				}
				if cfg.RateLimit.Rate != 100 {
					t.Errorf("expected RateLimit.Rate to be %d, got %d", 100, cfg.RateLimit.Rate)
				}
				if cfg.RateLimit.Burst != 5 {
					t.Errorf("expected RateLimit.Burst to be %d, got %d", 5, cfg.RateLimit.Burst)
				}
				if cfg.RateLimit.TimeWindow != time.Minute {
					t.Errorf("expected RateLimit.TimeWindow to be %v, got %v", time.Minute, cfg.RateLimit.TimeWindow)
				}
				if !cfg.RateLimit.PerIP {
					t.Error("expected RateLimit.PerIP to be true")
				}
			},
		},
		{
			name: "custom rate limit settings",
			envVars: map[string]string{
				// Required database config
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				// Rate limit config
				"RATE_LIMIT_ENABLED":     "true",
				"RATE_LIMIT_PER_IP":      "true",
				"RATE_LIMIT":             "200",
				"RATE_BURST":             "10",
				"RATE_LIMIT_TIME_WINDOW": "2m",
			},
			check: func(t *testing.T, cfg *Config) {
				if !cfg.RateLimit.Enabled {
					t.Error("expected RateLimit.Enabled to be true")
				}
				if !cfg.RateLimit.PerIP {
					t.Error("expected RateLimit.PerIP to be true")
				}
				if cfg.RateLimit.Rate != 200 {
					t.Errorf("expected RateLimit.Rate to be %d, got %d", 200, cfg.RateLimit.Rate)
				}
				if cfg.RateLimit.Burst != 10 {
					t.Errorf("expected RateLimit.Burst to be %d, got %d", 10, cfg.RateLimit.Burst)
				}
				if cfg.RateLimit.TimeWindow != 2*time.Minute {
					t.Errorf("expected RateLimit.TimeWindow to be %v, got %v", 2*time.Minute, cfg.RateLimit.TimeWindow)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Clean up after test
			defer os.Clearenv()

			cfg, err := New()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected config but got nil")
			}

			tt.check(t, cfg)
		})
	}
}
