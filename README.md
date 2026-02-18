# GoFormX (Go Forms API)

Forms API backend for GoFormX. Handles form CRUD, schema storage, submissions, and public embed/submit. The web UI (dashboard, form builder) lives in [goformx-laravel](https://github.com/goformx/goformx-laravel); this repo is API-only.

## Architecture

- **Authenticated API** (`/api/forms`): Used by Laravel. Requires signed headers `X-User-Id`, `X-Timestamp`, `X-Signature` (HMAC-SHA256). Laravel sends these after authenticating the user.
- **Public API** (`/forms/:id/...`): No auth. Embed page, schema, validation rules, and form submission for external sites. CORS and rate limiting apply.
- **Database**: PostgreSQL. Go owns forms, submissions, and related tables; Laravel has its own DB for users and sessions.

See the [split design doc](https://github.com/goformx/goformx-laravel/blob/main/docs/plans/2026-02-18-goformx-laravel-go-split-design.md) in goformx-laravel for the full architecture.

## Features

- Form CRUD and schema (Form.io–compatible)
- Submissions and event bus
- Laravel assertion auth (signed headers)
- Public embed and submit with CORS
- PostgreSQL, migrations (GORM)
- Uber FX, Echo, Zap, Testify, Task

## Tech Stack

- Go 1.25+
- PostgreSQL 17
- Echo v4
- Uber FX, GORM, Zap, Testify, Task

## Quick Start

1. **Prerequisites**

   - Go 1.25+
   - PostgreSQL
   - Task (optional; see `Taskfile.yml`)

2. **Clone and setup**

   ```bash
   git clone https://github.com/goformx/goforms.git
   cd goforms
   cp .env.example .env
   ```

3. **Environment**

   Set database and shared secret (must match Laravel):

   ```bash
   DATABASE_HOST=localhost
   DATABASE_PORT=5432
   DATABASE_NAME=goforms
   DATABASE_USERNAME=goforms
   DATABASE_PASSWORD=goforms
   GOFORMS_SHARED_SECRET=your-shared-secret
   ```

4. **Run**

   ```bash
   task migrate:up
   task dev:backend
   ```

   API: `http://localhost:8090`. Use with goformx-laravel (`GOFORMS_API_URL=http://localhost:8090`, same `GOFORMS_SHARED_SECRET`).

## API Overview

| Route | Auth | Purpose |
|-------|------|---------|
| `GET/POST /api/forms`, `GET/PUT/DELETE /api/forms/:id` | Assertion | Laravel form CRUD |
| `GET /api/forms/:id/submissions` | Assertion | List/get submissions |
| `GET /forms/:id/schema` | None | Public schema |
| `POST /forms/:id/submit` | None | Public submit |
| `GET /forms/:id/embed` | None | Embeddable form page |
| `GET /health` | None | Health check |

## Documentation

- [CLAUDE.md](CLAUDE.md) — development and architecture notes
- [API Documentation](docs/api/README.md) (if present)
- [Development Guide](docs/development/README.md) (if present)

## License

MIT — see [LICENSE](LICENSE).
