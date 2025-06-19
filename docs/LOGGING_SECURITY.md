# Logging Security Guide

This document outlines the security measures implemented in the GoFormX logging system to protect against log injection attacks and ensure safe handling of user input.

## Overview

The logging system in GoFormX is designed with security as a primary concern. All user input that flows through the logging system is properly sanitized and validated to prevent various types of attacks.

## Security Measures

### 1. Input Sanitization

All log messages and fields are sanitized using the `SanitizeForLogging()` method which:

- Removes newline characters (`\n`, `\r`, `\r\n`) that could be used for log injection
- Removes null bytes (`\x00`) that could corrupt log files
- Removes HTML tags using the sanitization service
- HTML escapes content to prevent XSS if logs are displayed in web interfaces
- Trims extra whitespace

### 2. Sensitive Data Protection

The system maintains a comprehensive list of sensitive keys that are automatically masked as "****" when logged:

```go
var sensitiveKeys = map[string]struct{}{
    "password":           {},
    "token":              {},
    "secret":             {},
    "key":                {},
    "credential":         {},
    "authorization":      {},
    "cookie":             {},
    "session":            {},
    "api_key":            {},
    "access_token":       {},
    "refresh_token":      {},
    // ... and many more
}
```

### 3. Path Validation

URL paths are validated before logging to prevent path traversal attacks:

- Must start with `/`
- Cannot contain dangerous characters (`\`, `<`, `>`, `"`, `'`, `\x00`, `\n`, `\r`)
- Cannot contain path traversal attempts (`..`, `//`)
- Limited to maximum length (500 characters)

### 4. User Agent Validation

User agent strings are validated to prevent injection attacks:

- Cannot contain dangerous characters
- Cannot contain suspicious patterns (`<script`, `javascript:`, `vbscript:`, `onload=`, `onerror=`)
- Limited to maximum length (1000 characters)

### 5. String Length Limits

All string fields are truncated to prevent log flooding attacks:

- General strings: 1000 characters maximum
- Path fields: 500 characters maximum
- Truncated strings are marked with "..." suffix

### 6. Error Handling

Error messages are properly sanitized before logging:

- Error content is processed through the same sanitization pipeline
- Error context is preserved while ensuring safety
- Stack traces are handled appropriately

## Code Flow Example

Here's how user input flows through the secure logging system:

```go
// 1. User input from HTTP request
h.Logger.Info("LoginValidation endpoint called",
    "method", c.Request().Method,        // Sanitized
    "path", c.Request().URL.Path,        // Validated and sanitized
    "user_agent", c.Request().UserAgent(), // Validated and sanitized
    "remote_addr", c.RealIP())           // Sanitized

// 2. Internal processing in logging factory
func (l *logger) Info(msg string, fields ...any) {
    l.zapLogger.Info(sanitizeMessage(msg, l.sanitizer), 
        convertToZapFields(fields, l.sanitizer)...)
}

// 3. Field conversion with validation
func convertToZapFields(fields []any, sanitizer sanitization.ServiceInterface) []zap.Field {
    // Each field is processed through sanitizeValue()
    sanitizedValue := sanitizeValue(key, value, sanitizer)
    return zap.String(key, sanitizedValue)
}

// 4. Value sanitization with type-specific validation
func sanitizeValue(key string, value any, sanitizer sanitization.ServiceInterface) string {
    // Check for sensitive keys
    if _, ok := sensitiveKeys[strings.ToLower(key)]; ok {
        return "****"
    }
    
    // Type-specific validation
    if key == "path" {
        if !validatePath(path) {
            return "[invalid path]"
        }
    }
    
    if key == "user_agent" {
        if !validateUserAgent(ua) {
            return "[invalid user agent]"
        }
    }
    
    // General sanitization
    return sanitizeString(truncateString(str, MaxStringLength), sanitizer)
}
```

## Security Best Practices

### 1. Always Use the Logger Interface

```go
// ✅ Good
logger.Info("User action", "user_id", userID, "action", action)

// ❌ Bad - direct string concatenation
log.Printf("User action: user_id=%s action=%s", userID, action)
```

### 2. Avoid Logging Sensitive Data

```go
// ✅ Good
logger.Info("User logged in", "user_id", userID, "ip", ip)

// ❌ Bad
logger.Info("User logged in", "password", password, "token", token)
```

### 3. Use Appropriate Log Levels

```go
// ✅ Good
logger.Debug("Generated validation schema", "fields_count", len(schema))
logger.Info("User authentication successful", "user_id", userID)
logger.Error("Database connection failed", "error", err)

// ❌ Bad
logger.Info("Generated validation schema", "schema", schema) // Too verbose for info level
```

### 4. Validate Input Before Logging

```go
// ✅ Good
if validatePath(path) {
    logger.Info("Request received", "path", path)
} else {
    logger.Warn("Invalid path received", "path", "[invalid path]")
}
```

## Testing Security

The logging system includes comprehensive tests for security features:

```bash
# Run logging security tests
go test ./internal/infrastructure/logging/... -v
```

Tests cover:
- Path validation
- User agent validation
- Sensitive data masking
- Log injection protection
- String length limits
- Error handling

## Monitoring and Alerting

Consider implementing monitoring for:

1. **Log Injection Attempts**: Monitor for patterns like `[invalid path]` or `[invalid user agent]`
2. **Sensitive Data Exposure**: Alert if sensitive keys are logged (should be masked)
3. **Log Volume**: Monitor for unusual log volume that might indicate attacks
4. **Error Patterns**: Track repeated validation failures

## Configuration

The logging system can be configured through environment variables:

```bash
# Log level
GOFORMS_LOG_LEVEL=info

# Log format
GOFORMS_LOG_FORMAT=json

# Output paths
GOFORMS_LOG_OUTPUT_PATHS=stdout,file.log
```

## Conclusion

The GoFormX logging system provides comprehensive protection against log injection attacks while maintaining useful logging capabilities. All user input is properly validated and sanitized, sensitive data is automatically masked, and the system includes extensive testing to ensure security measures work correctly.

For additional security, consider:
- Implementing log rotation and retention policies
- Using centralized logging with additional security controls
- Regular security audits of log content
- Implementing log integrity monitoring 