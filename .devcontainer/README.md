# Devcontainer Setup for GoForms

This directory contains configuration for the VS Code Devcontainer, which provides a reproducible development environment for GoForms.

## Overview

The devcontainer uses Docker Compose to spin up the application, database(s), and supporting services. It supports both **PostgreSQL** and **MariaDB** for local development and testing.

## Services

- **app**: The main GoForms application container.
- **postgres**: PostgreSQL database container (default for development).
- **mariadb**: MariaDB database container (optional, for MariaDB support).
- **adminer**: Database management UI (supports both databases).

## Database Initialization Scripts

- `.devcontainer/init-scripts/init.sql`: Initialization script for MariaDB.
  - Creates the `goforms` and `goforms_test` databases.
  - Creates users and grants privileges.
  - Sets some global settings for development.
- `.devcontainer/init-scripts/init-postgres.sql`: Initialization script for PostgreSQL.
  - Creates the `goforms` and `goforms_test` databases.
  - Creates users and grants privileges.
  - Sets up schemas and default privileges.
  - Sets timezone and connection limits.

**These scripts are executed automatically by Docker when the database containers are created for the first time.**

- If you delete your database volumes, these scripts will run again on the next container startup.
- If you keep your volumes, the scripts are not re-run.
- If you remove these scripts, you will need to create databases and users manually.

## Switching Between Databases

- The application supports both PostgreSQL and MariaDB.
- The database used is controlled by environment variables in `docker-compose.yml`:
  - `GOFORMS_DB_CONNECTION=postgres` (for PostgreSQL)
  - `GOFORMS_DB_CONNECTION=mariadb` (for MariaDB)
- Update the `app` service environment variables to switch between databases.
- Make sure the corresponding database service (`postgres` or `mariadb`) is enabled in `docker-compose.yml`.

## Usage

1. Open the project in VS Code and reopen in the devcontainer when prompted.
2. The containers will start and initialize the databases if needed.
3. Use the `adminer` service at [http://localhost:8098](http://localhost:8098) to inspect your databases.
4. Run migrations and start the app as described in the project README.

## Notes

- The initialization scripts are for development and CI convenience. In production, database setup should be handled separately and securely.
- If you change the initialization scripts, you must remove the database volumes to see the changes take effect. 