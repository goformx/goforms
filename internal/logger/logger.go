package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once   sync.Once
	logger Logger
)

// GetLogger returns a singleton instance of the Logger
func GetLogger() Logger {
	once.Do(func() {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.DisableStacktrace = true
		config.DisableCaller = false
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

		zapLogger, err := config.Build()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
		logger = newZapLogger(zapLogger)
	})
	return logger
}
