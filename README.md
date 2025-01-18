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

```shell
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

```shell
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

```http
POST /api/subscriptions
- Create new email subscription
- Rate limited
- Validates email format

GET /health
- Health check endpoint
- Returns service status
```

### Marketing Website API

```http
GET /v1/marketing/pages
- List all marketing pages
- Supports pagination
- Optional filtering by status
- Cached responses

GET /v1/marketing/pages/{id}
- Get specific marketing page
- Includes SEO metadata
- Cached responses

POST /v1/marketing/pages
- Create new marketing page
- Requires authentication
- Validates content format
- Rate limited

PUT /v1/marketing/pages/{id}
- Update existing page
- Requires authentication
- Validates content format
- Rate limited

GET /v1/marketing/stats
- Get marketing statistics
- Requires authentication
- Supports date range filtering
- Rate limited
```

### API Versioning

All new endpoints will be versioned under `/v1`:

```http
POST /v1/forms
GET  /v1/forms/{id}
PUT  /v1/forms/{id}
POST /v1/forms/{id}/submissions
```

### Development Guidelines

### Code Organization
- Follow clean architecture principles with distinct layers:
  - `internal/core/` - Domain logic and interfaces
  - `internal/platform/` - Infrastructure implementations
  - `internal/api/` - REST API endpoints
  - `internal/web/` - Web UI and templates

### Coding Standards
- Use Go 1.23 features appropriately
- Follow Go standard project layout
- Implement proper error handling and logging
- Write idiomatic Go code
- Use interfaces for dependency inversion
- Keep functions focused and small

### Dependencies
- Use Uber's fx for dependency injection
- Use Echo/v4 for web routing
- Use Zap for structured logging
- Use sqlx for database operations
- Use testify for testing
- Use templ for server-side rendering

### Testing Requirements
- Write unit tests for core domain logic
- Write integration tests for API endpoints
- Mock external dependencies in tests
- Aim for high test coverage
- Use table-driven tests where appropriate

### API Development
- Version all new endpoints under `/v1`
- Include OpenAPI/Swagger annotations
- Implement proper input validation
- Use consistent error response format
- Add rate limiting for public endpoints
- Group related functionality into fx.Module

### Database
- Use MariaDB as primary database
- Implement database migrations using golang-migrate
- Use sqlx for database operations
- Implement proper connection pooling
- Handle database errors appropriately

### Code Style
- Use lowercase with underscores for directories
- Favor named exports for functions
- Use clear, descriptive names for API endpoints
- Structure files: exported functions, subfunctions, helpers

### Observability
- Use structured logging with Zap
- Implement request tracking with unique request IDs
- Add health check endpoints
- Include detailed error reporting
- Use appropriate log levels (debug, info, warn, error)

### Error Handling
- Use custom error types when beneficial
- Include context in error messages
- Log errors with appropriate stack traces
- Return consistent error responses
- Handle all error cases explicitly

### Security
- Implement proper input validation
- Use secure headers middleware
- Configure CORS appropriately
- Implement rate limiting
- Follow security best practices

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
│   ├── api/          # API endpoints and handlers
│   │   └── v1/       # API version 1 endpoints
│   ├── app/          # Application setup and initialization
│   ├── config/       # Configuration management
│   ├── core/         # Core domain logic
│   │   ├── contact/     # Contact domain models and interfaces
│   │   └── subscription/# Subscription domain models and interfaces
│   ├── platform/     # Infrastructure implementations
│   │   ├── database/    # Database implementations
│   │   └── server/      # Server setup and configuration
│   ├── web/          # Web UI layer
│   │   ├── components/  # Reusable UI components
│   │   ├── handlers/    # Web request handlers
│   │   ├── layouts/     # Page layouts
│   │   └── pages/       # Page templates
│   ├── logger/       # Logging infrastructure
│   ├── middleware/   # HTTP middleware
│   └── response/     # Common response types
├── migrations/       # Database migrations
├── static/          # Static assets
│   └── css/         # CSS files
├── test/            # Test helpers and fixtures
└── Taskfile.yml     # Task automation configuration
```

This structure better reflects our clean architecture approach, with clear separation between:
- API layer (`internal/api/v1/`)
- Core domain logic (`internal/core/`)
- Infrastructure concerns (`internal/platform/`)
- Web UI layer (`internal/web/`)

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

## Architecture

The application follows a clean architecture approach:

### Core Layer
- Contains domain models and business logic
- No external dependencies
- Defines interfaces for infrastructure

### Platform Layer
- Implements infrastructure concerns
- Database access
- External services integration

### API Layer
- REST API endpoints
- Request/response handling
- Input validation

### Web Layer
- Server-side rendering with templ
- UI components and layouts
- Static assets

### Key Features

- Clean Architecture
- Domain-Driven Design
- Dependency Injection with Uber FX
- Type-safe templates with templ
- Structured logging with Zap
- MariaDB with sqlx
- Echo web framework
