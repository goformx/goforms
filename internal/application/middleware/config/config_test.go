package config

import (
	"testing"

	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
)

func TestNewMiddlewareConfig_ReferencesInfrastructureConfig(t *testing.T) {
	infra := &config.Config{}
	mc := NewMiddlewareConfig(infra)
	assert.Equal(t, &infra.Security, mc.Security)
	assert.Equal(t, &infra.Logging, mc.Logging)
}
