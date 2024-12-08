package database

type Config struct {
	Host           string
	Port           int
	Credentials    Credentials
	ConnectionPool PoolConfig
}
