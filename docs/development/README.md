# Development Guide

## Getting Started

### Prerequisites

- Docker
- VS Code with Dev Containers extension
- Git

### Development Environment

1. Clone the repository:

   ```bash
   git clone https://github.com/jonesrussell/goforms.git
   cd goforms
   ```

2. Open in VS Code:

   ```bash
   code .
   ```

3. When prompted, click "Reopen in Container" or use Command Palette:

   ```plaintext
   Dev Containers: Reopen in Container
   ```

4. Set up environment:

   ```bash
   cp .env.example .env
   task install
   task migrate:up
   ```

## Development Workflow

### Running the Application

```bash
# Start development server with hot reload
task dev

# Run without hot reload
task run
```

### Testing

```bash
# Run all tests
task test

# Run integration tests
task test:integration

# View test coverage
task test:coverage
```

### Database Operations

```bash
# Create new migration
task migrate:create name=migration_name

# Run migrations
task migrate:up

# Rollback migrations
task migrate:down
```

## Code Organization

```plaintext
.
├── cmd/                  # Application entrypoints
├── internal/            
│   ├── api/             # API endpoints
│   ├── app/             # Application setup
│   ├── core/            # Domain logic
│   ├── platform/        # Infrastructure
│   └── web/             # Web UI
├── migrations/          # Database migrations
├── static/             # Static assets
└── test/               # Test helpers
```

## Development Guidelines

### Code Style

- Follow Go standard project layout
- Use interfaces for dependency inversion
- Keep functions focused and small
- Write idiomatic Go code

### Testing Requirements

- Write unit tests for core logic
- Write integration tests for APIs
- Mock external dependencies
- Use table-driven tests
- Aim for high coverage

### Git Workflow

1. Create feature branch:

   ```bash
   git checkout -b feature/name
   ```

2. Make changes and commit:

   ```bash
   git add .
   git commit -m "Description of changes"
   ```

3. Push and create PR:

   ```bash
   git push origin feature/name
   ```

### Documentation

- Update API docs for endpoint changes
- Document new features
- Keep README.md current
- Add code comments for complex logic

## Tooling

### Task Runner

[Task](https://taskfile.dev) commands are defined in `Taskfile.yml`:

```yaml
tasks:
  install:
    desc: Install dependencies
  
  dev:
    desc: Start development server
  
  test:
    desc: Run tests
  
  migrate:up:
    desc: Run migrations
```

### VS Code Extensions

Recommended extensions:

- Go
- Dev Containers
- GitLens
- Go Test Explorer

### Debugging

1. Set breakpoints in VS Code
2. Use "Run and Debug" panel
3. Select "Go: Launch Package"
4. Start debugging session

## Further Reading

- [API Development](./api.md)
- [Testing Guide](./testing.md)
- [Database Guide](./database.md)
- [Deployment Guide](./deployment.md)
