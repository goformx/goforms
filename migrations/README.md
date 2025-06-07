# Database Migrations

This directory contains database migrations for both PostgreSQL and MariaDB. The migrations are organized in separate directories for each database type.

## Directory Structure

```
migrations/
├── mariadb/           # MariaDB-specific migrations
│   ├── 1983010101_create_users_table.up.sql
│   ├── 1983010101_create_users_table.down.sql
│   ├── 2004020401_create_forms_table.up.sql
│   └── 2004020401_create_forms_table.down.sql
├── postgres/          # PostgreSQL-specific migrations
│   ├── 1983010101_create_users_table.up.sql
│   ├── 1983010101_create_users_table.down.sql
│   ├── 2004020401_create_forms_table.up.sql
│   └── 2004020401_create_forms_table.down.sql
└── README.md
```

## Running Migrations

Use the `migrate` command-line tool to run migrations:

```bash
# Apply migrations to PostgreSQL
go run cmd/migrate/main.go -db postgres -dsn "postgres://user:password@localhost:5432/dbname?sslmode=disable"

# Apply migrations to MariaDB
go run cmd/migrate/main.go -db mariadb -dsn "user:password@tcp(localhost:3306)/dbname"

# Revert the last migration
go run cmd/migrate/main.go -db postgres -dsn "postgres://user:password@localhost:5432/dbname?sslmode=disable" -down
```

## Migration Files

Each migration consists of two files:
- `{version}_{name}.up.sql`: Contains the SQL to apply the migration
- `{version}_{name}.down.sql`: Contains the SQL to revert the migration

The version number should be in the format `YYYYMMDDHH` (e.g., `1983010101`).

## Creating New Migrations

1. Create a new migration file in both `mariadb/` and `postgres/` directories
2. Use the same version number for both database types
3. Ensure the SQL is compatible with the target database
4. Include both up and down migrations
5. Test the migrations on both database types

## Best Practices

1. Always include both up and down migrations
2. Use transactions in migrations
3. Keep migrations idempotent when possible
4. Test migrations on both database types
5. Use descriptive names for migration files
6. Follow the versioning scheme strictly 