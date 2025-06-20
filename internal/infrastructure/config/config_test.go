package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityConfig_DefaultsAndValidation(t *testing.T) {
	os.Setenv("GOFORMS_SECURITY_CSRF_ENABLED", "true")
	os.Setenv("GOFORMS_SECURITY_CSRF_SECRET", "testsecret123456789012345678901234567890")
	cfg, err := New()
	assert.NoError(t, err)
	assert.True(t, cfg.Security.CSRF.Enabled)
	assert.Equal(t, "testsecret123456789012345678901234567890", cfg.Security.CSRF.Secret)
}
