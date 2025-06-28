# Configuration Management

This package provides a comprehensive configuration management system for the GoForms application using [Viper](https://github.com/spf13/viper).

## Overview

The configuration system supports multiple formats and sources:
- **Environment Variables** (highest priority)
- **Configuration Files** (YAML, JSON, TOML, ENV)
- **Default Values** (lowest priority)

## File Structure

```
internal/infrastructure/config/
├── config.go      # Main configuration structs and validation
├── viper.go       # Viper-based configuration loader
├── module.go      # Fx dependency injection module
├── utils.go       # Configuration utilities
├── app.go         # Application configuration
├── database.go    # Database configuration
├── security.go    # Security configuration
├── services.go    # Service configurations
├── web.go         # Web server configuration
├── constants.go   # Configuration constants
└── README.md      # This file
```

## Configuration Sources

### 1. Environment Variables

Environment variables take the highest priority and override all other sources. They use the `GOFORMS_` prefix and dot notation is converted to underscores:

```bash
# App configuration
GOFORMS_APP_NAME=goforms
GOFORMS_APP_ENVIRONMENT=development
GOFORMS_APP_PORT=8090

# Database configuration
GOFORMS_DATABASE_HOST=postgres
GOFORMS_DATABASE_PORT=5432
GOFORMS_DATABASE_NAME=goforms
```

### 2. Configuration Files

Configuration files are searched in the following order:
1. Current directory (`.`)
2. `./config/` directory
3. `/etc/goforms/` (system-wide)
4. `$HOME/.goforms/` (user-specific)

Supported formats:
- `config.yaml` (default)
- `config.yml`
- `config.json`
- `config.toml`
- `config.env`

### 3. Default Values

Default values are set in the `setDefaults()` function in `viper.go` and serve as fallbacks when no other source provides a value.

## Configuration Structure

### App Configuration
```yaml
app:
  name: "GoForms"
  version: "1.0.0"
  environment: "development"
  debug: true
  log_level: "debug"
  url: "http://localhost:8090"
  scheme: "http"
  port: 8090
  host: "0.0.0.0"
  read_timeout: "5s"
  write_timeout: "10s"
  idle_timeout: "120s"
  request_timeout: "30s"
  vite_dev_host: "0.0.0.0"
  vite_dev_port: "5173"
```

### Database Configuration
```yaml
database:
  driver: "postgres"
  host: "postgres"
  port: 5432
  name: "goforms"
  username: "goforms"
  password: "goforms"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "5m"
  conn_max_idle_time: "5m"
```

### Security Configuration
```yaml
security:
  csrf:
    enabled: true
    secret: "your-csrf-secret"
    token_name: "_token"
    header_name: "X-CSRF-Token"
  cors:
    enabled: true
    allowed_origins:
      - "http://localhost:5173"
      - "http://localhost:8090"
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allowed_headers:
      - "Content-Type"
      - "Authorization"
      - "X-CSRF-Token"
      - "X-Requested-With"
    allow_credentials: true
    max_age: 3600
  rate_limit:
    enabled: true
    rps: 100
    burst: 5
  csp:
    enabled: true
    default_src: "'self'"
    script_src: "'self' 'unsafe-inline'"
    style_src: "'self' 'unsafe-inline'"
    img_src: "'self' data: https:"
    connect_src: "'self'"
    font_src: "'self'"
    object_src: "'none'"
    media_src: "'self'"
    frame_src: "'none'"
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
  encryption:
    key: ""
  secure_cookie: false
  debug: false
```

## Usage

### Basic Usage

```go
import "github.com/goformx/goforms/internal/infrastructure/config"

// The configuration is automatically loaded via Fx dependency injection
func MyHandler(cfg *config.Config) {
    fmt.Printf("App name: %s\n", cfg.App.Name)
    fmt.Printf("Database host: %s\n", cfg.Database.Host)
}
```

### Environment-Specific Configuration

```go
// Load configuration for a specific environment
vc := config.NewViperConfig()
cfg, err := vc.LoadForEnvironment("production")
if err != nil {
    log.Fatal(err)
}
```

### Configuration Utilities

```go
// Get configuration utilities
utils := config.NewConfigUtils()

// Find configuration file
configFile, err := utils.FindConfigFile()
if err != nil {
    log.Printf("No config file found: %v", err)
}

// Sanitize configuration for logging
sanitized := utils.SanitizeConfigForLogging(cfg)
log.Printf("Config: %+v", sanitized)
```

### Hot Reloading (Development)

```go
// Watch for configuration changes
vc := config.NewViperConfig()
vc.WatchConfig(func() {
    log.Println("Configuration changed, reloading...")
    // Reload your application configuration
})
```

## Validation

Configuration validation is performed automatically when loading:

```go
// Check if configuration is valid
if !cfg.IsValid() {
    log.Fatal("Invalid configuration")
}

// Get validation errors
if err := cfg.validateConfig(); err != nil {
    log.Fatal("Configuration validation failed:", err)
}
```

## Environment Detection

```go
// Check current environment
if cfg.IsDevelopment() {
    // Development-specific logic
}

if cfg.IsProduction() {
    // Production-specific logic
}

if cfg.IsStaging() {
    // Staging-specific logic
}

if cfg.IsTest() {
    // Test-specific logic
}
```

## Configuration Summary

```go
// Get a summary of the current configuration
summary := cfg.GetConfigSummary()
for section, data := range summary {
    fmt.Printf("%s: %+v\n", section, data)
}
```

## Best Practices

1. **Environment Variables**: Use environment variables for sensitive data and deployment-specific settings
2. **Configuration Files**: Use configuration files for application defaults and non-sensitive settings
3. **Validation**: Always validate configuration before using it
4. **Sanitization**: Use `SanitizeConfigForLogging()` when logging configuration to avoid exposing secrets
5. **Hot Reloading**: Use hot reloading only in development environments
6. **Environment-Specific Files**: Use `config.{environment}.yaml` for environment-specific settings

## Migration from Old System

The old manual configuration loading system has been replaced with Viper. Key changes:

1. **Removed**: `LoadFromEnv()`, `loader.go`, manual environment variable parsing
2. **Added**: Viper-based loading, multi-format support, hot reloading
3. **Updated**: Environment variable naming convention (dot notation to underscores)
4. **Enhanced**: Better validation, utilities, and error handling

## Troubleshooting

### Common Issues

1. **Configuration not loading**: Check file paths and permissions
2. **Environment variables not working**: Ensure proper `GOFORMS_` prefix
3. **Validation errors**: Check required fields and data types
4. **Hot reloading not working**: Ensure file watching is enabled

### Debug Mode

Enable debug mode to see detailed configuration loading information:

```bash
GOFORMS_APP_DEBUG=true
GOFORMS_APP_LOG_LEVEL=debug
``` 