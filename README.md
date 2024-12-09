# Goforms

A modern Go web application for form management with MariaDB backend.

## Features

Current:
- ✅ Email subscription system with validation
- ✅ RESTful API using Echo framework
- ✅ MariaDB database with migrations
- ✅ Dependency injection using Uber FX
- ✅ Structured logging with Zap
- ✅ Rate limiting and CORS support
- ✅ Comprehensive test coverage
- ✅ Docker-based development environment

Coming Soon:
- 🚧 Form Management API
- 🚧 Custom Form Fields
- 🚧 Form Analytics
- 🚧 Advanced Security Features

## Development Setup

This project uses VS Code Dev Containers for development. Make sure you have:
- Docker installed
- VS Code with Dev Containers extension
- Git

### Getting Started

1. Clone the repository
2. Open in VS Code
3. When prompted, click "Reopen in Container"
   - Or use Command Palette: "Dev Containers: Reopen in Container"
4. Copy `.env.example` to `.env` and adjust values if needed

The container will:
- Set up Go 1.23 environment
- Initialize MariaDB database
- Install required tools (migrate, MariaDB client, task)

### Task Commands

We use [Task](https://taskfile.dev) for project automation:

```bash
# Install dependencies
task install

# Run the application
task run

# Run tests
task test

# Run integration tests
task test:integration

# View test coverage
task test:coverage

# Database operations
task migrate:up      # Run migrations
task migrate:down    # Rollback migrations
task migrate:create  # Create new migration
```

### Environment Variables

Key configuration options in `.env`:

```env
# Server Configuration
SERVER_PORT=8090
SERVER_HOST=localhost

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=goforms
DB_PASSWORD=goforms
DB_NAME=goforms

# Security
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://jonesrussell.github.io
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RATE=100
```

### API Endpoints

Current endpoints:

```
POST /api/subscriptions
- Create new email subscription
- Rate limited
- Validates email format

GET /health
- Health check endpoint
- Returns service status
```

### API Versioning

All new endpoints will be versioned under `/v1`:
```
POST /v1/forms
GET  /v1/forms/{id}
PUT  /v1/forms/{id}
POST /v1/forms/{id}/submissions
```

### Development Guidelines

- All new endpoints must include OpenAPI/Swagger annotations
- Use fx.Module for feature grouping
- Follow REST best practices for resource naming
- Include rate limiting per endpoint
- Add comprehensive test coverage

### Tech Stack

- Go 1.23
- MariaDB 10.11
- Echo v4 web framework
- Uber FX for dependency injection
- Zap for structured logging
- Testify for testing
- Task for automation

## Project Structure

```
.
├── .devcontainer/     # Development container configuration
├── .github/           # GitHub workflows and configuration
├── cmd/              
│   └── server/        # Application entrypoint
├── internal/          
│   ├── app/          # Application setup and initialization
│   ├── config/       # Configuration management
│   ├── database/     # Database connection and utilities
│   ├── handlers/     # HTTP handlers
│   ├── middleware/   # Custom middleware
│   └── models/       # Data models and business logic
├── migrations/        # Database migrations
├── test/             # Test helpers and fixtures
└── Taskfile.yml      # Task automation configuration
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### Observability

The application includes several observability features:
- Structured logging with Zap
- Request ID tracking
- Health check endpoints
- Detailed error reporting
- Performance metrics

### Middleware Stack

The middleware is configured in the following order for optimal security and functionality:
1. Recovery middleware (panic recovery)
2. Logging middleware (request logging)
3. Request ID middleware (request tracking)
4. Security middleware (HTTP security headers)
5. CORS middleware (Cross-Origin Resource Sharing)
6. Rate limiting middleware (request rate limiting)
