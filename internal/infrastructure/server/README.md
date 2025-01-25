# Server Package

This package handles HTTP server lifecycle management.

## Responsibilities
- HTTP server initialization
- Server lifecycle management (start/stop)
- Graceful shutdown
- Server configuration
- Error handling and logging

## Key Components
- `Server`: Main server structure
- `New`: Server factory with lifecycle hooks
- Lifecycle hooks for Fx integration
- Graceful shutdown handling

## Usage
```go
srv := server.New(lc, logger, config)
``` 