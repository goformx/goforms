# GoForms

A modern Go web application for form management with MariaDB backend.

## Features

✨ **Core Features**
- Form Management System
- Contact Submissions
- Email Subscriptions
- Modern UI with Dark Mode
- RESTful API
- MariaDB Database

🛠️ **Technical Features**
- Clean Architecture
- Dependency Injection (Uber FX)
- Type-safe Templates (templ)
- Structured Logging (Zap)
- Task Automation
- Docker Development

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
