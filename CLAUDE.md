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
- Entry point: `src/main.ts` with Inertia app setup + Sonner toast provider
- Page components in `src/pages/` (e.g., `Dashboard/Index.vue`, `Forms/Edit.vue`)
- Path aliases: `@/`, `@/components`, `@/pages`, `@/composables`, `@/lib`
- **Tailwind CSS v4** with `@tailwindcss/postcss` plugin
- **Form.io** integration with custom components via `@goformx/formio`
- **shadcn-vue** component library for UI primitives
- Build output: `dist/`

#### Frontend Architecture

**Component Structure:**
```
src/
├── components/
│   ├── ui/                  # shadcn-vue UI primitives (22+ components)
│   │   ├── button/, card/, input/, badge/, alert/
│   │   ├── dialog/, sheet/, tabs/, tooltip/, popover/
│   │   ├── dropdown-menu/, select/, switch/
│   │   ├── command/, separator/, scroll-area/
│   │   ├── sonner/ (toast), skeleton/, table/
│   │   └── ... (see components.json for full list)
│   ├── layout/              # Layout wrappers
│   │   ├── AppLayout.vue, DashboardLayout.vue, GuestLayout.vue
│   ├── shared/              # Shared components
│   │   ├── Nav.vue, UserMenu.vue, DashboardHeader.vue
│   ├── form-builder/        # Form builder specific
│   │   ├── BuilderLayout.vue       # Three-panel builder layout
│   │   ├── FieldsPanel.vue         # Searchable field library
│   │   └── FieldSettingsPanel.vue  # Inline field settings
│   └── dashboard/           # Dashboard specific
│       └── FormCard.vue            # Form card component
├── composables/             # Vue 3 Composition API composables
│   ├── useFormBuilder.ts           # Form.io builder integration
│   ├── useFormValidation.ts        # Zod validation
│   ├── useFormBuilderState.ts      # Builder state + undo/redo
│   ├── useKeyboardShortcuts.ts     # Keyboard shortcuts system
│   ├── useThemeCustomization.ts    # Theme management
│   └── useCommandPalette.ts        # Command palette logic
├── pages/                   # Inertia.js pages
│   ├── Dashboard/Index.vue  # Grid layout with search/filter
│   └── Forms/Edit.vue       # Three-panel builder
└── lib/
    └── utils.ts             # Tailwind class utilities (cn)
```

**Key Composables:**

1. **`useFormBuilder.ts`** - Form.io builder integration with:
   - Auto-save with debounce (2s)
   - Undo/redo history (50 actions)
   - Field CRUD operations (duplicate, delete)
   - Schema import/export
   - Selected field state management

2. **`useFormBuilderState.ts`** - Centralized builder state:
   - Selected field tracking
   - Dirty state management
   - Undo/redo history with localStorage
   - Field CRUD operations

3. **`useKeyboardShortcuts.ts`** - Platform-aware shortcuts:
   - Cmd on Mac, Ctrl on Windows
   - Auto-cleanup on unmount
   - Enable/disable toggling

4. **`useThemeCustomization.ts`** - Theme management:
   - CSS variable injection
   - Presets: Linear, Stripe, Notion, Vercel
   - Load/save to server and localStorage

5. **`useCommandPalette.ts`** - Command palette:
   - Fuzzy search across commands
   - Recent commands tracking (last 10)
   - localStorage persistence

**Form Builder Architecture:**

The form builder uses a three-panel layout:
- **Left Panel**: Searchable field library (Basic, Layout, Advanced)
- **Center Canvas**: Form.io builder instance
- **Right Panel**: Tabbed field settings (Display, Data, Validation)

Keyboard shortcuts:
- `Cmd+S` - Save form
- `Cmd+P` - Preview form
- `Cmd+Z` / `Cmd+Shift+Z` - Undo/Redo
- `Cmd+D` - Duplicate selected field
- `Cmd+Backspace` - Delete selected field
- `Cmd+/` - Show shortcuts help

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

### UI Component Patterns

#### Toast Notifications

Use Sonner for all transient feedback (replaces inline alerts):

```typescript
import { toast } from "vue-sonner";

// Success
toast.success("Form saved successfully");

// Error
toast.error("Failed to save form");

// Info
toast.info("Processing...");

// With description
toast.success("Form saved", {
  description: "Your changes have been saved successfully"
});
```

**Toast Provider Setup:**
The Sonner toast provider is configured in `src/main.ts` with:
- Position: `top-right`
- Rich colors: enabled
- Auto-dismiss: default (4s)

#### shadcn-vue Component Usage

All UI components follow shadcn-vue patterns with Radix Vue primitives:

```vue
<script setup lang="ts">
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>Form Title</CardTitle>
    </CardHeader>
    <CardContent>
      <div class="space-y-4">
        <div class="space-y-2">
          <Label for="name">Name</Label>
          <Input id="name" v-model="form.name" type="text" />
        </div>
        <Button @click="save">Save</Button>
      </div>
    </CardContent>
  </Card>
</template>
```

**Component Variants:**
```vue
<!-- Buttons -->
<Button variant="default">Default</Button>
<Button variant="destructive">Delete</Button>
<Button variant="outline">Cancel</Button>
<Button variant="ghost">Ghost</Button>
<Button variant="link">Link</Button>
<Button size="sm">Small</Button>
<Button size="lg">Large</Button>
<Button size="icon">🔍</Button>

<!-- Badges -->
<Badge variant="default">Default</Badge>
<Badge variant="secondary">Draft</Badge>
<Badge variant="destructive">Error</Badge>
<Badge variant="outline">Outline</Badge>

<!-- Alerts -->
<Alert variant="default">Info message</Alert>
<Alert variant="destructive">Error message</Alert>
```

#### Keyboard Shortcuts Pattern

Use `useKeyboardShortcuts` composable for keyboard-first interactions:

```typescript
import { useKeyboardShortcuts } from "@/composables/useKeyboardShortcuts";

const shortcuts = [
  {
    key: "s",
    meta: true,  // Cmd on Mac, Ctrl on Windows
    handler: () => save(),
    description: "Save form"
  },
  {
    key: "z",
    meta: true,
    shift: true,  // Cmd+Shift+Z
    handler: () => redo(),
    description: "Redo"
  }
];

useKeyboardShortcuts(shortcuts);
```

#### Form Builder Integration

When integrating with the form builder:

```typescript
import { useFormBuilder } from "@/composables/useFormBuilder";

const {
  isLoading,
  error,
  isSaving,
  saveSchema,
  getSchema,
  selectedField,
  selectField,
  duplicateField,
  deleteField,
  undo,
  redo,
  canUndo,
  canRedo,
} = useFormBuilder({
  containerId: "form-schema-builder",
  formId: props.form.id,
  autoSave: false,  // Set to true for auto-save with 2s debounce
  onSchemaChange: (schema) => {
    // Handle schema changes
  },
});
```

#### Responsive Design Patterns

Use Tailwind responsive prefixes:

```vue
<template>
  <!-- Grid: 1 column mobile, 2 tablet, 3 desktop -->
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
    <FormCard v-for="form in forms" :key="form.id" :form="form" />
  </div>

  <!-- Stack on mobile, row on desktop -->
  <div class="flex flex-col sm:flex-row gap-4">
    <Input class="flex-1" />
    <Button>Submit</Button>
  </div>

  <!-- Hide on mobile, show on desktop -->
  <div class="hidden lg:block">Desktop only</div>
</template>
```

#### Component Communication

**Props & Emits Pattern:**
```typescript
interface Props {
  form: Form;
  readonly?: boolean;
}

interface Emits {
  (e: "update", form: Form): void;
  (e: "delete", formId: string): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

function handleUpdate() {
  emit("update", props.form);
}
```

**Composable Pattern:**
```typescript
// Composable for shared logic
export function useFormActions(formId: string) {
  const isLoading = ref(false);

  async function save() {
    isLoading.value = true;
    try {
      // Save logic
      toast.success("Saved");
    } catch (err) {
      toast.error("Failed to save");
    } finally {
      isLoading.value = false;
    }
  }

  return { isLoading, save };
}
```

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
func (h *FormWebHandler) handleUpdate(c echo.Context) error {
    logger := h.Logger.WithComponent("form_handler").
        WithOperation("update").
        With("form_id", form.ID)

    if err := h.FormService.UpdateForm(ctx, form, req); err != nil {
        logger.Error("form update failed", "error", err)
        return h.handleFormUpdateError(c, form, err)
    }

    logger.Info("form updated successfully")
    return c.Redirect(http.StatusSeeOther, redirectURL)
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

## Frontend Modernization

The frontend has been modernized with a Linear/Vercel/Stripe/Notion-inspired design language. Key improvements:

**New Features:**
- Three-panel form builder with collapsible sides
- Keyboard shortcuts for power users (Cmd+S, Cmd+Z, etc.)
- Toast notifications (via Sonner)
- Modern grid layouts with search/filter
- Undo/redo with 50-action history
- Auto-save with debounce
- Theme customization system

**When Building New Features:**
- Use shadcn-vue components for UI primitives
- Use composables for shared logic (not Pinia/Vuex)
- Implement keyboard shortcuts for common actions
- Use toast notifications instead of inline alerts
- Follow responsive design patterns (mobile-first)
- Add undo/redo for complex state changes

**Reference Documentation:**
- Full details in `MODERNIZATION_SUMMARY.md`
- Component examples in existing pages (`Dashboard/Index.vue`, `Forms/Edit.vue`)
- Composable patterns in `src/composables/`
