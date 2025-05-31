package config

// Config holds the configuration for creating a logger
type Config struct {
	Level   string
	AppName string
	Debug   bool
}

// New creates a new logging configuration
func New() *Config {
	return &Config{
		Level:   "debug",
		AppName: "goforms",
		Debug:   true,
	}
}
