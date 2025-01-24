# GoForms

A modern Go web application for form management with MariaDB backend.

## Features

✨ **Core Features**

- Form Management System
  - Contact Form Submissions
  - Email Subscription Management
  - Status Tracking
  - Validation
- RESTful API
  - OpenAPI/Swagger Documentation
  - Versioned Endpoints (v1)
  - Standardized Response Format
- Modern UI
  - Server-side Rendering
  - Dark Mode Support
  - Responsive Design
- MariaDB Database
  - Connection Pooling
  - Migration System

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

## Dependencies

The project follows a strict dependency rule where dependencies point inward:
- Core Domain: No external dependencies
- Platform: Implements core interfaces
- Application: Depends on core and platform
- Infrastructure: Configures and connects components
- Presentation: Consumes application services

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
   git clone https://github.com/jonesrussell/goforms.git
   cd goforms
   code .
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
