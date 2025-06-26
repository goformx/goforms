package config

// FormConfig holds form-related configuration
type FormConfig struct {
	MaxFileSize    int64    `envconfig:"GOFORMS_MAX_FILE_SIZE" default:"10485760"` // 10MB
	AllowedTypes   []string `envconfig:"GOFORMS_ALLOWED_FILE_TYPES" default:"image/jpeg,image/png,application/pdf"`
	MaxSubmissions int      `envconfig:"GOFORMS_MAX_SUBMISSIONS" default:"1000"`
	RetentionDays  int      `envconfig:"GOFORMS_RETENTION_DAYS" default:"90"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	Version   string `envconfig:"GOFORMS_API_VERSION" default:"v1"`
	Prefix    string `envconfig:"GOFORMS_API_PREFIX" default:"/api"`
	RateLimit int    `envconfig:"GOFORMS_API_RATE_LIMIT" default:"100"`
	Timeout   int    `envconfig:"GOFORMS_API_TIMEOUT" default:"30"`
}

// WebConfig holds web-related configuration
type WebConfig struct {
	BaseURL      string `envconfig:"GOFORMS_WEB_BASE_URL" default:"http://localhost:8090"`
	AssetsDir    string `envconfig:"GOFORMS_WEB_ASSETS_DIR" default:"./assets"`
	TemplatesDir string `envconfig:"GOFORMS_WEB_TEMPLATES_DIR" default:"./templates"`
}

// UserConfig holds user-related configuration
type UserConfig struct {
	Admin   AdminUserConfig   `envconfig:"GOFORMS_ADMIN"`
	Default DefaultUserConfig `envconfig:"GOFORMS_USER"`
}

// AdminUserConfig holds admin user configuration
type AdminUserConfig struct {
	Email     string `envconfig:"GOFORMS_ADMIN_EMAIL" validate:"required,email"`
	Password  string `envconfig:"GOFORMS_ADMIN_PASSWORD" validate:"required"`
	FirstName string `envconfig:"GOFORMS_ADMIN_FIRST_NAME" validate:"required"`
	LastName  string `envconfig:"GOFORMS_ADMIN_LAST_NAME" validate:"required"`
}

// DefaultUserConfig holds default user configuration
type DefaultUserConfig struct {
	Email     string `envconfig:"GOFORMS_USER_EMAIL" validate:"required,email"`
	Password  string `envconfig:"GOFORMS_USER_PASSWORD" validate:"required"`
	FirstName string `envconfig:"GOFORMS_USER_FIRST_NAME" validate:"required"`
	LastName  string `envconfig:"GOFORMS_USER_LAST_NAME" validate:"required"`
}
