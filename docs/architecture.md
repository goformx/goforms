# Architecture Guide

## Observability

The application includes several observability features:
- Structured logging with Zap
- Request ID tracking
- Health check endpoints
- Detailed error reporting
- Performance metrics

## Middleware Stack

The middleware is configured in the following order for optimal security and functionality:
1. Recovery middleware (panic recovery)
2. Logging middleware (request logging)
3. Request ID middleware (request tracking)
4. Security middleware (HTTP security headers)
5. CORS middleware (Cross-Origin Resource Sharing)
6. Rate limiting middleware (request rate limiting)

[Additional architecture details...] 