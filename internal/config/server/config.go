package server

type Config struct {
	Host     string
	Port     int
	Timeouts TimeoutConfig
}
