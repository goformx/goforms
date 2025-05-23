# GoFormX

A modern Go web application for form management with MariaDB backend.

## Features

Current (MVP):
- ✅ Email subscription system with validation
- ✅ RESTful API using Echo framework
- ✅ MariaDB database with migrations
- ✅ Dependency injection using Uber FX
- ✅ Structured logging with Zap
- ✅ Rate limiting and CORS support
- ✅ Comprehensive test coverage
- ✅ Docker-based development environment
- ✅ Health check monitoring

Coming in V2:
- 🎯 Multi-tenant support
  - API key authentication
  - Per-tenant rate limiting
  - Domain/CORS management
- 🎯 Form Management
  - Form builder
  - Field validation
  - Form versioning
- 🎯 Submission Management
  - Data storage
  - Export capabilities

Future Enhancements:
- 🚧 Advanced Form Features (conditional logic, multi-page)
- 🚧 Integration Features (webhooks, notifications)
- 🚧 Analytics & Monitoring
- 🚧 Administration Features

🛠️ **Technical Features**

- Clean Architecture
  - Domain-Driven Design
  - Separation of Concerns
  - SOLID Principles
- Dependency Injection (Uber FX)
- Structured Logging (Zap)
- Comprehensive Testing
  - Unit Tests
  - Mock Implementations
  - Test Utilities
- Task Automation
- Docker Development

## Architecture

The project follows Clean Architecture principles with clear separation of concerns:

### Core Domain Layer (`/internal/core/`)
Contains business logic and domain entities:
- Domain Models and Interfaces
- Business Rules and Validation
- Use Cases and Services
- No External Dependencies

### Platform Layer (`/internal/platform/`)
Implements infrastructure and technical concerns:
- Database Implementations
- Server Configuration
- Error Handling
- External Integrations

### Application Layer
```
/internal/
├── api/          - API endpoints and versioning
├── handlers/     - Request handlers
├── middleware/   - HTTP middleware
├── response/     - Response formatting
└── validation/   - Input validation
```

### Infrastructure Layer
```
/internal/
├── app/          - Application bootstrapping
├── config/       - Configuration management
├── database/     - Database connections
├── logger/       - Logging infrastructure
└── web/         - Web server setup
```

### Presentation Layer
```
/internal/
├── components/   - UI components
├── ui/          - UI logic
└── view/        - View templates
```

## API Versioning

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

### Architecture Overview
- Echo web framework with middleware stack
- MariaDB with connection pooling
- Dependency injection with Uber FX
- Structured logging with Zap

### Technical Documentation
Detailed technical documentation can be found in the `/docs` directory:
- [Architecture Guide](docs/architecture.md)
- [API Documentation](docs/api.md)
- [Development Guide](docs/development.md)

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
/internal/
├── api/          - API endpoints and handlers
├── app/          - Application setup
├── components/   - UI components
├── config/       - Configuration
├── core/         - Business logic
├── database/     - Database layer
├── handlers/     - HTTP handlers
├── logger/       - Logging
├── middleware/   - HTTP middleware
├── models/       - Data models
├── platform/     - Platform code
├── response/     - API responses
├── ui/          - UI code
├── validation/   - Input validation
├── view/        - View templates
└── web/         - Web server
```

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

📚 **Comprehensive documentation is available in the `docs` directory:**

- [API Documentation](docs/api/README.md)
- [Development Guide](docs/development/README.md)
- [Architecture Overview](docs/architecture/README.md)

## Tech Stack

- Go 1.23
- MariaDB 10.11
- Echo v4
- Uber FX
- Zap Logger
- Task Runner

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/development/README.md#git-workflow) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
