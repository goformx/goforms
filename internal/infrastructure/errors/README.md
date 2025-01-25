# Errors Package

This package provides common infrastructure error types and utilities.

## Responsibilities
- Define common infrastructure errors
- Error wrapping utilities
- Error type checking
- Error context management

## Key Components
- Common error types
- Error wrapping functions
- Error checking utilities

## Usage
```go
if err != nil {
    return errors.Wrap(err, "operation failed")
}
``` 