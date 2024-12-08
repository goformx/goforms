# Goforms

A modern Go web application for form management with MariaDB backend.

## Features

- ✅ Email subscription system with validation
- ✅ RESTful API using Echo framework
- ✅ MariaDB database with migrations
- ✅ Dependency injection using Uber FX
- ✅ Structured logging with Zap
- ✅ Rate limiting and CORS support
- ✅ Comprehensive test coverage
- ✅ Docker-based development environment

## Development Setup

This project uses VS Code Dev Containers for development. Make sure you have:
- Docker installed
- VS Code with Dev Containers extension

### Getting Started

1. Clone the repository
2. Open in VS Code
3. When prompted, click "Reopen in Container"
   - Or use Command Palette: "Dev Containers: Reopen in Container"

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

### Database Configuration

Database credentials are configured in `.devcontainer/.env`:

```env
DB_USER=goforms
DB_PASSWORD=goforms
DB_DATABASE=goforms
DB_HOST=db
```

### API Endpoints

Current endpoints:

```
POST /api/subscriptions
- Create new email subscription
- Rate limited
- Validates email format
```

### Tech Stack

- Go 1.23
- MariaDB
- Echo web framework
- Uber FX for dependency injection
- Zap for logging
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
│   └── models/       # Data models and business logic
├── migrations/        # Database migrations
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

### Middleware Stack

The middleware is configured in the following order for optimal security and functionality:
1. Recovery middleware (panic recovery)
2. Logging middleware (request logging)
3. Request ID middleware (request tracking)
4. Security middleware (HTTP security headers)
5. CORS middleware (Cross-Origin Resource Sharing)
6. Rate limiting middleware (request rate limiting)
