package config_test

import (
	"testing"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityConfig_DefaultsAndValidation(t *testing.T) {
	t.Setenv("GOFORMS_SECURITY_CSRF_ENABLED", "true")
	t.Setenv("GOFORMS_SECURITY_CSRF_SECRET", "testsecret123456789012345678901234567890")
	cfg, err := config.New()
	require.NoError(t, err)
	assert.True(t, cfg.Security.CSRF.Enabled)
	assert.Equal(t, "testsecret123456789012345678901234567890", cfg.Security.CSRF.Secret)
}
