# Logging Package

This package provides unified logging functionality using zap.

## Responsibilities
- Structured logging interface
- Log level management
- Field-based logging
- Fx event logging
- Testing utilities

## Key Components
- `Logger`: Main logging interface
- `FxEventLogger`: Uber Fx event logging
- `LoggerFactory`: Logger creation and configuration
- Field helpers (String, Int, Bool, Error, etc.)

## Usage
```go
logger := logging.NewLogger(debug, appName)
logger.Info("message", logging.String("key", "value"))
``` 