# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Install dependencies (Go tools + npm packages)
task install

# Generate code (templ templates + mocks)
task generate

# Build entire application (frontend + backend)
task build

# Run development environment with hot reload
task dev

# Run only backend or frontend
task dev:backend    # Uses air for hot reload
task dev:frontend   # Vite dev server on :5173

# Linting
task lint           # All linters
task lint:backend   # Go: fmt, vet, golangci-lint
task lint:frontend  # ESLint

# Testing
task test                  # All tests
task test:backend          # Go unit tests
task test:backend:cover    # With coverage report
task test:integration      # Integration tests (build tag: integration)

# Run a single Go test
go test -v -run TestFunctionName ./path/to/package/...

# Database migrations
task migrate:up     # Apply migrations
task migrate:down   # Rollback one migration
```

## Architecture Overview

GoFormX follows **Clean Architecture** with Uber FX dependency injection:

```
internal/
├── domain/           # Business entities, interfaces, services (form/, user/, common/)
├── application/      # HTTP handlers, middleware, validation, response builders
├── infrastructure/   # Database, config, logging, server, event bus
└── presentation/     # Templ templates and view rendering
```

**Dependency flow**: Infrastructure → Application → Domain (dependencies point inward)

### Key Architectural Patterns

1. **Uber FX Modules** - DI is organized in modules loaded in `main.go`:

   - `config.Module` → `infrastructure.Module` → `domain.Module` → `application.Module` → `presentation.Module` → `web.Module`

2. **Handler Interface** - All HTTP handlers implement `web.Handler` with `Register()`, `Start()`, `Stop()` methods and are collected via FX groups.

3. **Service-Repository Pattern** - Handlers call services (business logic) which use repositories (data access) and may emit events via EventBus.

4. **Dual Middleware System** - Currently migrating from legacy `Manager` to new `Orchestrator` system:
   - Legacy: `internal/application/middleware/manager.go`
   - New: `internal/application/middleware/orchestrator.go`
   - Migration adapter provides fallback capability

### Frontend (TypeScript/Vite)

- Entry points in `src/js/` with multiple page-specific entries (main.ts, dashboard.ts, form-builder.ts, etc.)
- Path aliases: `@/core`, `@/features`, `@/pages`, `@/shared`
- Form.io integration with custom components via `@goformx/formio`
- Build output: `dist/`

### Code Generation

- **Templ templates**: `*.templ` → `*_templ.go` via `templ generate`
- **Mocks**: Generated in `test/mocks/` via `go generate ./...` (uses mockgen)

## Configuration

Uses Viper with environment variables. Key env var pattern: `<SECTION>_<KEY>`

```bash
APP_ENV=development
DB_HOST=localhost
SECURITY_CSRF_COOKIE_SAME_SITE=Lax
```

Configuration struct: `internal/infrastructure/config/`

## Database

- **PostgreSQL** (primary) or MariaDB
- Migrations in `migrations/postgresql/` and `migrations/mariadb/`
- Uses GORM for ORM

## Development Environment

- Uses Dev Containers (VS Code)
- Backend: `localhost:8090`
- Frontend dev server: `localhost:5173`
- CSRF configured for cross-origin development

## Code Style

- Go: snake_case files, standard Go naming conventions
- Error handling: Always wrap errors with `fmt.Errorf("context: %w", err)`
- Linting: golangci-lint v2 with 40+ linters enabled (see `.golangci.yml`)
- Frontend: ESLint + Prettier, strict TypeScript
