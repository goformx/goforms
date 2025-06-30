package middleware

import "github.com/goformx/goforms/internal/application/middleware/core"

type ChainConfig struct {
	Enabled         bool
	MiddlewareNames []string
	Paths           []string // Path patterns for this chain
	CustomConfig    map[string]interface{}
}

type MiddlewareConfig interface {
	IsMiddlewareEnabled(name string) bool
	GetMiddlewareConfig(name string) map[string]interface{}
	GetChainConfig(chainType core.ChainType) ChainConfig
}
