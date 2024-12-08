package server

import "time"

// TimeoutConfig contains server timeout settings
type TimeoutConfig struct {
	Read  time.Duration
	Write time.Duration
	Idle  time.Duration
}
