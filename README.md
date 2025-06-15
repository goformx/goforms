# GoFormX

A modern Go web application for form management with MariaDB backend.

## Features

- Email subscription system with validation
- RESTful API using Echo framework
- MariaDB database with migrations
- Dependency injection using Uber FX
- Structured logging with Zap
- Rate limiting and CORS support
- Comprehensive test coverage
- Docker-based development environment
- Health check monitoring

## Tech Stack

- Go 1.24
- MariaDB 10.11
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
   - Install dependencies: `task install:all`
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
