package config

const (
	// Environment constants
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvStaging     = "staging"
	EnvTest        = "test"

	// Database connection types
	DBConnectionPostgreSQL = "postgres"
	DBConnectionMariaDB    = "mariadb"
	DBConnectionMySQL      = "mysql"

	// Cache types
	CacheTypeMemory = "memory"
	CacheTypeRedis  = "redis"

	// Storage types
	StorageTypeLocal = "local"
	StorageTypeS3    = "s3"
	StorageTypeGCS   = "gcs"

	// Session types
	SessionTypeNone   = "none"
	SessionTypeMemory = "memory"
	SessionTypeFile   = "file"
	SessionTypeRedis  = "redis"

	// Log levels
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"

	// Log formats
	LogFormatJSON = "json"
	LogFormatText = "text"

	// SSL modes for PostgreSQL
	SSLModeDisable    = "disable"
	SSLModeRequire    = "require"
	SSLModeVerifyCA   = "verify-ca"
	SSLModeVerifyFull = "verify-full"

	// Default file size limits
	DefaultMaxFileSize = 10 * 1024 * 1024 // 10MB

	// Default timeouts
	DefaultReadTimeout    = "5s"
	DefaultWriteTimeout   = "10s"
	DefaultIdleTimeout    = "120s"
	DefaultRequestTimeout = "30s"

	// Default ports
	DefaultHTTPPort  = 8090
	DefaultHTTPSPort = 8443
	DefaultVitePort  = 5173

	// Security defaults
	DefaultCSRFTokenLength = 32
	DefaultPasswordCost    = 12

	// Rate limiting defaults
	DefaultRateLimitRequests = 100
	DefaultRateLimitWindow   = "1m"
	DefaultRateLimitBurst    = 20
)
