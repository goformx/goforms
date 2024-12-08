# Goforms

A modern Go web application with PostgreSQL backend.

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
- Initialize PostgreSQL database
- Install required tools (migrate, PostgreSQL client)

### Database Configuration

Database credentials are configured in `.devcontainer/.env`:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=goforms
POSTGRES_HOSTNAME=localhost
```

### Running Migrations

```bash
# Create a new migration
migrate create -ext sql -dir migrations -seq migration_name

# Run migrations
migrate -database "postgres://postgres:postgres@localhost:5432/goforms?sslmode=disable" -path migrations up
```

### Tech Stack

- Go 1.23
- PostgreSQL
- Echo web framework
- Uber FX for dependency injection
- Zap for logging
- Testify for testing

## Project Structure

```
.
├── .devcontainer/     # Development container configuration
├── migrations/        # Database migrations
├── cmd/              # Application entrypoints
└── internal/         # Private application code
```
