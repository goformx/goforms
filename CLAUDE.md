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

GoFormX is the **forms API backend**. The web UI (dashboard, form builder) lives in **goformx-laravel** (Laravel + Inertia/Vue). This repo is API-only: form domain, public embed/submit, and assertion auth.

**Clean Architecture** with Uber FX:

```
internal/
├── domain/           # Business entities, interfaces, services (form/, common/)
├── application/      # HTTP handlers, middleware, validation, response builders
└── infrastructure/   # Database, config, logging, server, event bus
```

**Dependency flow**: Infrastructure → Application → Domain (dependencies point inward)

### Key Architectural Patterns

1. **Uber FX Modules** - DI in `main.go`: `config.Module` → `infrastructure.Module` → `domain.Module` → `application.Module` → `web.Module`.

2. **Handler Interface** - HTTP handlers implement `web.Handler` with `Register()`, `Start()`, `Stop()` and are collected via FX groups.

3. **Service-Repository Pattern** - Handlers call domain services; services use repositories and may emit events via EventBus.

4. **Laravel assertion auth** - Authenticated form API uses signed headers from Laravel: `X-User-Id`, `X-Timestamp`, `X-Signature` (HMAC-SHA256 of `user_id:timestamp`). Middleware: `internal/application/middleware/assertion/`. Config: `security.assertion.secret` (env `GOFORMS_SHARED_SECRET`), `timestamp_skew_seconds`.

### API Surface

- **Authenticated** (require assertion headers): `GET/POST /api/forms`, `GET/PUT/DELETE /api/forms/:id`, `GET /api/forms/:id/submissions`, `GET /api/forms/:id/submissions/:sid`. Used by Laravel only.
- **Public** (no auth): `GET /forms/:id/schema`, `GET /forms/:id/validation`, `POST /forms/:id/submit`, `GET /forms/:id/embed`. For embedded forms and public submission. CORS and rate limiting apply.

### Frontend (goformx-laravel)

The **UI lives in the Laravel app** (goformx-laravel). This repo has **no** Inertia, no page rendering, no auth UI, no Vite assets. For Form.io embed HTML and public endpoints only.

### Code Generation

- **Mocks**: Generated in `test/mocks/` via `go generate ./...` (uses mockgen)

## Configuration

Uses Viper with environment variables. Viper maps nested config to env vars using underscore separator:

- Config key `database.host` → env var `DATABASE_HOST`
- Config key `security.assertion.secret` → env var `GOFORMS_SHARED_SECRET` (must match Laravel)

```bash
APP_ENV=development
DATABASE_HOST=postgres-dev
GOFORMS_SHARED_SECRET=    # Same value as in goformx-laravel .env
SECURITY_CSRF_COOKIE_SAME_SITE=Lax
```

Configuration struct: `internal/infrastructure/config/`
Default values: `internal/infrastructure/config/viper.go` (see `setDatabaseDefaults`, etc.)

## Database

- **PostgreSQL** (primary) or MariaDB
- Migrations in `migrations/postgresql/` and `migrations/mariadb/`
- Uses GORM for ORM

## Development Environment

- **Go only**: Runs standalone at `localhost:8090`. No built-in frontend; UI is in goformx-laravel.
- Laravel runs separately (e.g. `localhost:8000`) and calls Go with `GOFORMS_API_URL=http://localhost:8090` and the same `GOFORMS_SHARED_SECRET`.
- PostgreSQL: `localhost:5432` (or per docker-compose). Go owns form tables; Laravel has its own DB for users/sessions.

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
GOFORMS_SHARED_SECRET=         # Same as Laravel; used for assertion verification
```

### Embed / Form.io

Public embed route `GET /forms/:id/embed` serves HTML that loads Form.io from CDN. CSP may need to allow `https://cdn.form.io` for embed pages.

## Code Style

- Go: snake_case files, standard Go naming conventions
- Error handling: Always wrap errors with `fmt.Errorf("context: %w", err)`
- Linting: golangci-lint v2 with 40+ linters enabled (see `.golangci.yml`)

## Logging Conventions

### DO:

- Log at handler/boundary level, not inside helpers
- Use structured key-value pairs: `logger.Info("msg", "key", value)`
- Include contextual fields: request_id, user_id, form_id
- Use error wrapping: `fmt.Errorf("context: %w", err)`
- Log once per error (at the boundary that handles it)
- Use `logging.LoggerFromContext(ctx)` to get an enriched logger

### DON'T:

- Use `println`, `fmt.Printf`, or `log.Printf`
- Use `c.Logger()` - use the structured logger instead
- Log entire request bodies
- Log in tight loops or low-level utilities
- Log secrets (password, token, key, secret, credential)
- Duplicate logs (if returning error, don't also log it)

### Log Levels:

- **DEBUG**: Development-only, verbose tracing
- **INFO**: Normal operations (request completed, form created)
- **WARN**: Recoverable issues (slow request, rate limited)
- **ERROR**: Failures requiring attention (DB errors, auth failures)
- **FATAL**: Unrecoverable (startup failures only)

### Required Fields by Context:

- All requests: `request_id`, `method`, `path`, `status`, `latency_ms`
- Authenticated: + `user_id`
- Form operations: + `form_id`
- Errors: + `error`, `error_type`

### Logging Patterns:

```go
// Handler-level logging with context enrichment
func (h *FormAPIHandler) handleUpdateForm(c echo.Context) error {
    logger := h.Logger.WithComponent("form_api").WithOperation("update").With("form_id", formID)
    if err := h.FormServiceHandler.UpdateForm(ctx, formID, req); err != nil {
        logger.Error("form update failed", "error", err)
        return h.ErrorHandler.HandleError(c, err)
    }
    logger.Info("form updated successfully")
    return c.NoContent(http.StatusNoContent)
}

// Type-safe field construction for complex scenarios
logger.InfoWithFields("form created",
    logging.String("form_id", form.ID),
    logging.String("user_id", userID),
    logging.Int("field_count", len(form.Fields)),
)
```

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

### Pre-commit Checklist

Before committing, ensure:
- `task lint` passes
- `task test` passes
- No new linter warnings introduced
