package logger

import (
	"sync"

	"go.uber.org/zap"
)

func getLoggerInstance() *zap.Logger {
	var once sync.Once
	var instance *zap.Logger

	once.Do(func() {
		var err error
		instance, err = zap.NewDevelopment()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
	})

	return instance
}

// GetLogger returns a singleton instance of the zap.Logger
func GetLogger() *zap.Logger {
	return getLoggerInstance()
}
