# Config Package

This package handles application configuration management.

## Responsibilities
- Loading environment variables
- Validating configuration values
- Providing typed configuration structures
- Managing configuration defaults

## Key Components
- `Config`: Main configuration structure
- `AppConfig`: Application-specific settings
- `ServerConfig`: HTTP server settings
- `DatabaseConfig`: Database connection settings
- `SecurityConfig`: Security and authentication settings

## Usage
```go
cfg := config.New()
``` 