package logger

// Logger defines the interface for logging operations
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
}

// Field represents a log field
type Field interface{}

// String creates a string field
func String(key string, value string) Field {
	return zapString(key, value)
}

// Int creates an integer field
func Int(key string, value int) Field {
	return zapInt(key, value)
}

// Duration creates a duration field
func Duration(key string, value interface{}) Field {
	return zapDuration(key, value)
}

// Error creates an error field
func Error(err error) Field {
	return zapError(err)
}
