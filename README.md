# GoFormX

A modern Go web application for form management with MariaDB backend.

## Features

- Email subscription system with validation
- RESTful API using Echo framework
- PostgreSQL database with migrations
- Dependency injection using Uber FX
- Structured logging with Zap
- Rate limiting and CORS support
- Comprehensive test coverage
- Docker-based development environment
- Health check monitoring

## Tech Stack

- Go 1.25
- PostgreSQL 17
- Echo v4 web framework
- Uber FX for dependency injection
- Zap for structured logging
- Testify for testing
- Task for automation

## Quick Start

1. Prerequisites:

   - Docker
   - VS Code with Dev Containers
   - Git

2. Clone and Setup:

   ```bash
   git clone https://github.com/goformx/goforms.git
   cd goforms
   ```

3. Start Development:

   - Click "Reopen in Container" when prompted
   - Copy environment file: `cp .env.example .env`
   - Install dependencies: `task install`
   - Start server: `task dev`

4. View the application at `http://localhost:8090`

## Documentation

Documentation is available in the `docs` directory:

- [API Documentation](docs/api/README.md)
- [Development Guide](docs/development/README.md)
- [Architecture Overview](docs/architecture/README.md)

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/development/README.md#git-workflow) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Development Setup

### CSRF Configuration for Development

When running the frontend (localhost:5173) and backend (localhost:8090) on different ports, you need to configure CSRF properly for cross-origin requests:

1. **Set CSRF Cookie SameSite to Lax**: This allows cookies to be sent in cross-origin requests
2. **Disable Secure Flag**: In development, cookies don't need to be HTTPS-only
3. **Include CSRF Headers in CORS**: Allow the `X-Csrf-Token` header

The application automatically configures these settings in development mode, but you can override them with environment variables:

```bash
# CSRF Configuration for Development
SECURITY_CSRF_COOKIE_SAME_SITE=Lax
SECURITY_SECURE_COOKIE=false

# CORS Configuration
SECURITY_CORS_ENABLED=true
SECURITY_CORS_ORIGINS=http://localhost:5173
SECURITY_CORS_CREDENTIALS=true
```

### Troubleshooting CSRF Issues

If you encounter 403 Forbidden errors with CSRF token mismatch:

1. **Clear Browser Cookies**: Old CSRF cookies may be invalid
2. **Restart the Backend**: Ensure new CSRF configuration is loaded
3. **Check Browser Console**: Verify CSRF token is being sent in headers
4. **Check Network Tab**: Ensure cookies are being sent with requests

The frontend automatically includes CSRF tokens in the `X-Csrf-Token` header for all non-GET requests.
