# OpenAPI Validation Middleware

The OpenAPI Validation Middleware provides automatic validation of HTTP requests and responses against your OpenAPI 3.1 specification. This ensures API consistency and helps catch contract violations early in development.

## Features

- **Request Validation**: Validates incoming requests against the OpenAPI spec
- **Response Validation**: Validates outgoing responses against the OpenAPI spec
- **Configurable Behavior**: Choose between logging violations or blocking requests/responses
- **Selective Application**: Apply validation to specific routes or the entire API
- **Performance Optimized**: Minimal overhead with efficient validation

## Quick Start

The middleware is already integrated into your GoFormX application. Here's how to use it:

### 1. Basic Usage

The middleware is automatically provided through the dependency injection system:

```go
// The middleware is already configured in main.go
app := fx.New(
    // ... other modules
    providers.OpenAPIValidationProvider(),
    // ... other providers
)
```

### 2. Integration with Echo

To add the middleware to your Echo instance:

```go
// In your setup function
func setupOpenAPIValidation(e *echo.Echo, validationMiddleware *middleware.OpenAPIValidationMiddleware) {
    // Add to all routes
    e.Use(validationMiddleware.Middleware())

    // Or add to specific API group
    apiGroup := e.Group("/api/v1")
    apiGroup.Use(validationMiddleware.Middleware())
}
```

### 3. Using the Integration Helper

For easier integration, use the provided helper:

```go
func setupOpenAPIValidation(e *echo.Echo, validationMiddleware *middleware.OpenAPIValidationMiddleware, logger logging.Logger) {
    integration := middleware.NewOpenAPIIntegration(validationMiddleware, logger)

    // Add to all routes
    integration.AddToEcho(e)

    // Or add to specific group
    apiGroup := e.Group("/api/v1")
    integration.AddToGroup(apiGroup)
}
```

## Configuration

The middleware supports various configuration options:

```go
config := &middleware.Config{
    // Enable request validation
    EnableRequestValidation: true,

    // Enable response validation
    EnableResponseValidation: true,

    // Log validation errors (doesn't block requests)
    LogValidationErrors: true,

    // Block requests that don't match the spec
    BlockInvalidRequests: false, // Start with logging only

    // Block responses that don't match the spec
    BlockInvalidResponses: false, // Start with logging only

    // Paths to skip validation
    SkipPaths: []string{
        "/health",
        "/metrics",
        "/docs",
        "/openapi.yaml",
        "/openapi.json",
    },

    // HTTP methods to skip validation
    SkipMethods: []string{
        "OPTIONS",
        "HEAD",
    },
}

validationMiddleware, err := middleware.NewOpenAPIValidationMiddleware(logger, config)
```

### Default Configuration

The default configuration is conservative and safe for production:

- ✅ Request validation enabled
- ✅ Response validation enabled
- ✅ Validation errors logged
- ❌ Invalid requests blocked (logs only)
- ❌ Invalid responses blocked (logs only)
- Skips health checks, metrics, and documentation routes

## Production Configuration

For production environments, you may want to enable blocking:

```go
config := middleware.DefaultConfig()
config.BlockInvalidRequests = true   // Block invalid requests
config.BlockInvalidResponses = true  // Block invalid responses
config.LogValidationErrors = true    // Keep logging for monitoring
```

## Validation Examples

### Request Validation

The middleware validates:

- HTTP method matches the spec
- Path parameters match the schema
- Query parameters match the schema
- Request body matches the schema
- Required headers are present

### Response Validation

The middleware validates:

- HTTP status code matches the spec
- Response headers match the schema
- Response body matches the schema

## Error Handling

### Logging Mode (Default)

When validation fails in logging mode:

```json
{
  "level": "warn",
  "msg": "Request validation failed",
  "error": "request validation failed: route not found in OpenAPI spec",
  "path": "/api/v1/forms",
  "method": "GET",
  "ip": "127.0.0.1"
}
```

### Blocking Mode

When validation fails in blocking mode:

```json
{
  "error": "Request validation failed: route not found in OpenAPI spec"
}
```

## Performance Considerations

- **Request Validation**: Minimal overhead, validates only when route is found in spec
- **Response Validation**: Captures response body, may have memory impact for large responses
- **Caching**: The OpenAPI spec is parsed once at startup
- **Skipping**: Use skip paths/methods to avoid unnecessary validation

## Troubleshooting

### Common Issues

1. **"route not found in OpenAPI spec"**

   - Ensure your route is defined in the OpenAPI specification
   - Check path parameter formats (e.g., `/forms/{id}` vs `/forms/:id`)

2. **"request validation failed"**

   - Verify request body matches the schema
   - Check required parameters are present
   - Ensure parameter types match (string vs integer)

3. **"response validation failed"**
   - Verify response body matches the schema
   - Check required fields are present in response
   - Ensure response status code is defined in spec

### Debug Mode

Enable debug logging to see detailed validation information:

```go
config := middleware.DefaultConfig()
config.LogValidationErrors = true
// Add debug logging to your logger configuration
```

## API Documentation

The middleware works with the existing OpenAPI documentation endpoints:

- `/api/v1/openapi.yaml` - Raw OpenAPI specification
- `/api/v1/openapi.json` - JSON format specification
- `/api/v1/docs` - Swagger UI documentation
- `/api/v1/validate` - Validate custom OpenAPI specs

## Best Practices

1. **Start with Logging**: Begin with `BlockInvalidRequests: false` to monitor violations
2. **Gradual Rollout**: Enable blocking for specific routes first
3. **Monitor Performance**: Watch for validation overhead in high-traffic scenarios
4. **Keep Spec Updated**: Ensure your OpenAPI spec matches your actual API
5. **Test Thoroughly**: Validate all endpoints before enabling blocking mode

## Integration with Testing

The middleware can be used in tests to validate API contracts:

```go
func TestAPIValidation(t *testing.T) {
    // Setup Echo with validation middleware
    e := echo.New()
    validationMiddleware, _ := middleware.NewOpenAPIValidationMiddleware(logger, config)
    e.Use(validationMiddleware.Middleware())

    // Your test requests will be validated against the spec
    req := httptest.NewRequest(http.MethodGet, "/api/v1/forms", nil)
    rec := httptest.NewRecorder()
    e.ServeHTTP(rec, req)

    // Validation errors will be logged or cause failures based on config
}
```
