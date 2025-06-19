# Database Migrations

This directory contains database migrations for GoForms, organized by database type to support both PostgreSQL and MariaDB.

## Structure

```
migrations/
├── postgresql/          # PostgreSQL-specific migrations
│   ├── 1970010101_create_users_table.up.sql
│   ├── 1970010101_create_users_table.down.sql
│   ├── 1983010101_create_forms_table.up.sql
│   ├── 1983010101_create_forms_table.down.sql
│   ├── 1991080601_create_form_submissions.up.sql
│   ├── 1991080601_create_form_submissions.down.sql
│   ├── 2004020401_create_form_schemas.up.sql
│   └── 2004020401_create_form_schemas.down.sql
├── mariadb/             # MariaDB-specific migrations
│   ├── 1970010101_create_users_table.up.sql
│   ├── 1970010101_create_users_table.down.sql
│   ├── 1983010101_create_forms_table.up.sql
│   ├── 1983010101_create_forms_table.down.sql
│   ├── 1991080601_create_form_submissions.up.sql
│   ├── 1991080601_create_form_submissions.down.sql
│   ├── 2004020401_create_form_schemas.up.sql
│   └── 2004020401_create_form_schemas.down.sql
└── README.md
```

## Database Support

### PostgreSQL
- Uses triggers and functions for `updated_at` timestamp updates
- Supports JSON data types natively
- Uses `DO $$` blocks for conditional logic

### MariaDB
- Uses `ON UPDATE CURRENT_TIMESTAMP` for automatic timestamp updates
- Supports JSON data types
- Simpler migration structure without triggers

## Usage

The migration system automatically selects the appropriate migration directory based on the `GOFORMS_DB_CONNECTION` environment variable:

- `postgres` → uses `migrations/postgresql/`
- `mariadb` → uses `migrations/mariadb/`

### Commands

```bash
# Run migrations
task migrate:up

# Rollback last migration
task migrate:down

# Rollback all migrations
task migrate:down-all

# Show current version
task migrate:version

# Create new migration
task migrate:create name=add_new_table

# Force migration version (fix dirty state)
task migrate:force version=1970010101
```

## Environment Variables

Set these environment variables to configure the database connection:

```bash
# Database type (postgres or mariadb)
GOFORMS_DB_CONNECTION=postgres

# Database connection details
GOFORMS_DB_HOST=localhost
GOFORMS_DB_PORT=5432
GOFORMS_DB_DATABASE=goforms
GOFORMS_DB_USERNAME=goforms
GOFORMS_DB_PASSWORD=goforms

# PostgreSQL specific
GOFORMS_DB_SSLMODE=disable
```

## Creating New Migrations

When creating new migrations, you'll need to create files in both directories:

1. Create the migration in the appropriate directory:
   ```bash
   task migrate:create name=add_new_feature
   ```

2. Copy the generated files to both directories:
   ```bash
   cp migrations/postgresql/*.sql migrations/mariadb/
   ```

3. Modify each file to use the appropriate database syntax:
   - **PostgreSQL**: Use triggers and functions for `updated_at`
   - **MariaDB**: Use `ON UPDATE CURRENT_TIMESTAMP`

## Migration Naming Convention

Migrations follow the format: `YYYYMMDDHHMMSS_description.up.sql`

Example: `1970010101_create_users_table.up.sql`

## Best Practices

1. **Always create both PostgreSQL and MariaDB versions** of each migration
2. **Test migrations on both database types** before deploying
3. **Use descriptive names** for migration files
4. **Keep migrations atomic** - one logical change per migration
5. **Include down migrations** for rollback capability
6. **Document any database-specific logic** in comments

## Migration Files

Each migration consists of two files:
- `*.up.sql`: Contains the SQL to apply the migration
- `*.down.sql`: Contains the SQL to revert the migration

## Running Migrations

The application supports both PostgreSQL and MariaDB. The database type is determined by the `GOFORMS_DB_CONNECTION` environment variable.

### Environment Variables

```bash
# Database Type (postgres or mariadb)
GOFORMS_DB_CONNECTION=postgres

# Database Connection
GOFORMS_DB_HOST=localhost
GOFORMS_DB_PORT=5432  # 3306 for MariaDB
GOFORMS_DB_DATABASE=goforms
GOFORMS_DB_USERNAME=goforms
GOFORMS_DB_PASSWORD=goforms
GOFORMS_DB_SSLMODE=disable  # Only used for PostgreSQL
```

### Apply Migrations

```bash
# Apply migrations using task
task migrate:up

# Or directly using migrate command
migrate -path migrations -database "postgresql://goforms:goforms@localhost:5432/goforms?sslmode=disable" up
migrate -path migrations -database "goforms:goforms@tcp(localhost:3306)/goforms?multiStatements=true" up
```

## Creating New Migrations

When creating new migrations:

1. Create a single migration file that works for both databases
2. Use database-agnostic SQL when possible
3. Use conditional SQL for database-specific features
4. Test the migrations on both database types

## Best Practices

1. Keep migrations idempotent when possible
2. Test both up and down migrations
3. Use database-agnostic SQL when possible
4. Document any database-specific features
5. Maintain consistent naming
6. Test migrations in both development and production environments

## Database-Specific Features

### PostgreSQL
- Uses triggers and functions for `updated_at` timestamps
- Supports JSONB data type
- Case-sensitive identifiers by default

### MariaDB
- Uses `ON UPDATE CURRENT_TIMESTAMP` for `updated_at` timestamps
- Supports JSON data type
- Case-insensitive identifiers by default

## Troubleshooting

If you encounter issues with migrations:

1. Check the database connection settings
2. Verify the migration files exist
3. Ensure the SQL syntax is compatible with the target database
4. Check the migration logs for specific errors
5. Test the migration in a development environment first 