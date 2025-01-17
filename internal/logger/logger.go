package logger

import (
	"sync"

	"go.uber.org/zap"
)

type Logger struct {
	instance *zap.Logger
}

var (
	loggerInstance *Logger
	once           sync.Once
)

// GetLogger returns a singleton instance of the zap.Logger
func GetLogger() *zap.Logger {
	once.Do(func() {
		var err error
		zapLogger, err := zap.NewDevelopment()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
		loggerInstance = &Logger{instance: zapLogger}
	})
	return loggerInstance.instance
}
