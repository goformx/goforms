# Database Package

This package manages database connections and lifecycle.

## Responsibilities
- Database connection management
- Connection pooling configuration
- Transaction support
- Database lifecycle hooks
- Error handling and logging

## Key Components
- `Database`: Main database wrapper
- `NewDB`: Database connection factory
- Transaction methods
- Connection pool management

## Usage
```go
db, err := database.NewDB(cfg, logger)
``` 