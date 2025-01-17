package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	once   sync.Once
	logger *zap.Logger
)

// GetLogger returns a singleton instance of the zap.Logger
func GetLogger() *zap.Logger {
	once.Do(func() {
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
	})
	return logger
}
