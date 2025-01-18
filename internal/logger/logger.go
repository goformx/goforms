package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once   sync.Once
	logger *zap.Logger
)

// GetLogger returns a singleton instance of the zap.Logger
func GetLogger() *zap.Logger {
	once.Do(func() {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.DisableStacktrace = true
		config.DisableCaller = false
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

		var err error
		logger, err = config.Build()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
	})
	return logger
}
