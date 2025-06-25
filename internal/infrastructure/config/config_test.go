package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goformx/goforms/internal/infrastructure/config"
)

func TestSecurityConfig_DefaultsAndValidation(t *testing.T) {
	// Set required environment variables for the test
	t.Setenv("GOFORMS_SECURITY_CSRF_ENABLED", "true")
	t.Setenv("GOFORMS_SECURITY_CSRF_SECRET", "testsecret123456789012345678901234567890")

	// Set required database environment variables to prevent validation errors
	t.Setenv("GOFORMS_DB_HOST", "localhost")
	t.Setenv("GOFORMS_DB_PORT", "3306")
	t.Setenv("GOFORMS_DB_DATABASE", "testdb")
	t.Setenv("GOFORMS_DB_USERNAME", "testuser")
	t.Setenv("GOFORMS_DB_PASSWORD", "testpass")
	t.Setenv("GOFORMS_DB_ROOT_PASSWORD", "rootpass")

	// Set required user environment variables
	t.Setenv("GOFORMS_ADMIN_EMAIL", "admin@test.com")
	t.Setenv("GOFORMS_ADMIN_PASSWORD", "adminpass")
	t.Setenv("GOFORMS_ADMIN_FIRST_NAME", "Admin")
	t.Setenv("GOFORMS_ADMIN_LAST_NAME", "User")
	t.Setenv("GOFORMS_USER_EMAIL", "user@test.com")
	t.Setenv("GOFORMS_USER_PASSWORD", "userpass")
	t.Setenv("GOFORMS_USER_FIRST_NAME", "Test")
	t.Setenv("GOFORMS_USER_LAST_NAME", "User")

	cfg, err := config.New()
	require.NoError(t, err)
	assert.True(t, cfg.Security.CSRF.Enabled)
	assert.Equal(t, "testsecret123456789012345678901234567890", cfg.Security.CSRF.Secret)
}
