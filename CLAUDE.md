# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Install dependencies (Go tools + npm packages)
task install

# Generate code (mocks)
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
└── presentation/     # Inertia.js rendering for Vue SPA
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

### Frontend (TypeScript/Vite/Vue 3)

- **Inertia.js** for SPA routing - Go backend renders pages, Vue handles client-side
- Entry point: `src/main.ts` with Inertia app setup
- Page components in `src/pages/` (e.g., `Dashboard/Index.vue`, `Forms/Edit.vue`)
- Path aliases: `@/`, `@/components`, `@/pages`, `@/composables`, `@/lib`
- **Tailwind CSS v4** with `@tailwindcss/postcss` plugin
- **Form.io** integration with custom components via `@goformx/formio`
- Build output: `dist/`

### Inertia.js Patterns

**IMPORTANT**: Web handlers must return Inertia-compatible responses:

```go
// ✅ Good - Render page for GET requests
return h.Inertia.Render(c, "Forms/Edit", inertia.Props{
    "title": "Edit Form",
    "form":  formData,
})

// ✅ Good - Redirect after mutations (POST/PUT/DELETE)
return c.Redirect(http.StatusSeeOther, "/dashboard")

// ✅ Good - Render page with error flash
return h.Inertia.Render(c, "Forms/New", inertia.Props{
    "title": "Create Form",
    "flash": map[string]string{"error": "Form title is required"},
})

// ❌ Bad - JSON response breaks Inertia
return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
```

The gonertia template uses `{{ .inertia }}` and `{{ .inertiaHead }}` placeholders (see `internal/presentation/inertia/inertia.go`).

### Code Generation

- **Mocks**: Generated in `test/mocks/` via `go generate ./...` (uses mockgen)

## Configuration

Uses Viper with environment variables. Viper maps nested config to env vars using underscore separator:

- Config key `database.host` → env var `DATABASE_HOST`
- Config key `security.csrf.enabled` → env var `SECURITY_CSRF_ENABLED`

```bash
APP_ENV=development
DATABASE_HOST=postgres-dev
SECURITY_CSRF_COOKIE_SAME_SITE=Lax
```

Configuration struct: `internal/infrastructure/config/`
Default values: `internal/infrastructure/config/viper.go` (see `setDatabaseDefaults`, etc.)

## Database

- **PostgreSQL** (primary) or MariaDB
- Migrations in `migrations/postgresql/` and `migrations/mariadb/`
- Uses GORM for ORM

## Development Environment

- Uses Docker Compose for local development
- Backend: `localhost:8090` (Go/Echo)
- Frontend dev server: `localhost:5173` (Vite)
- PostgreSQL: `localhost:5432`
- Access app via `localhost:8090` (not 5173) - Go serves HTML, Vite serves assets

### Docker Commands

```bash
docker compose up              # Start all services
docker compose down            # Stop all services
docker compose restart goforms-dev  # Restart just the app
docker compose logs -f goforms-dev  # Follow logs
```

### Environment Variables

Viper maps config keys to env vars: `database.host` → `DATABASE_HOST`

Key `.env` variables:
```bash
DATABASE_HOST=postgres-dev     # Docker service name
DATABASE_PORT=5432
DATABASE_NAME=goforms
DATABASE_USERNAME=goforms
DATABASE_PASSWORD=goforms
```

### CSP Configuration

Form.io requires CDN access. Update `.env` for development:
```bash
SECURITY_CSP_SCRIPT_SRC="'self' 'unsafe-inline' 'unsafe-eval' http://localhost:5173 https://localhost:5173 https://cdn.form.io blob:"
SECURITY_CSP_STYLE_SRC="'self' 'unsafe-inline' http://localhost:5173 https://localhost:5173 https://cdn.form.io"
SECURITY_CSP_CONNECT_SRC="'self' ws: wss: http://localhost:5173 https://localhost:5173 https://cdn.form.io"
SECURITY_CSP_FONT_SRC="'self' http://localhost:5173 https://localhost:5173 https://cdn.form.io"
```

## Code Style

- Go: snake_case files, standard Go naming conventions
- Error handling: Always wrap errors with `fmt.Errorf("context: %w", err)`
- Linting: golangci-lint v2 with 40+ linters enabled (see `.golangci.yml`)
- Frontend: ESLint + Prettier, strict TypeScript

## Linting Requirements

**IMPORTANT**: All code must pass linting before commit. Run `task lint` to verify.

### Go Linting Rules

1. **Use `any` instead of `interface{}`** (revive: use-any)
   ```go
   // ❌ Bad
   data := map[string]interface{}{"key": "value"}
   
   // ✅ Good
   data := map[string]any{"key": "value"}
   ```

2. **No magic numbers** (mnd) - Extract to named constants
   ```go
   // ❌ Bad
   if len(token) > 20 { ... }
   
   // ✅ Good
   const tokenPreviewLength = 20
   if len(token) > tokenPreviewLength { ... }
   ```

3. **Line length max 150 characters** (lll) - Break long lines
   ```go
   // ❌ Bad
   println("[DEBUG] Very long message with many params:", param1, ", param2=", param2, ", param3=", param3, ", param4=", param4)
   
   // ✅ Good
   println("[DEBUG] Message:", param1, ", param2=", param2)
   println("[DEBUG] More params:", param3, ", param4=", param4)
   ```

4. **Security: Avoid unsafe template functions** (gosec G203)
   ```go
   // When using template.JS/HTML with trusted data, add nosec comment
   return template.JS(trustedData) // #nosec G203 - data is from trusted source
   ```

### TypeScript/Frontend Linting Rules

1. **Handle promises properly** (@typescript-eslint/no-floating-promises)
   ```typescript
   // ❌ Bad
   onMounted(() => {
     initializeAsync();
   });
   
   // ✅ Good - use void for intentionally ignored promises
   onMounted(() => {
     void initializeAsync();
   });
   ```

2. **Use nullish coalescing** (@typescript-eslint/prefer-nullish-coalescing)
   ```typescript
   // ❌ Bad
   const value = data.items || [];
   
   // ✅ Good
   const value = data.items ?? [];
   ```

### Pre-commit Checklist

Before committing, ensure:
- `task lint` passes (both backend and frontend)
- `task test` passes
- No new linter warnings introduced
