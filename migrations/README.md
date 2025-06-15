# Database Migrations

This directory contains database migrations that work with both PostgreSQL and MariaDB. The migrations use database-agnostic SQL where possible and include database-specific features when needed.

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