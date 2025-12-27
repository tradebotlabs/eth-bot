# Database Migrations

This directory contains SQL migration files for the ETH Trading Bot database schema.

## Migration Files

Migrations follow the naming convention: `{version}_{description}.{up|down}.sql`

- **Up migrations** (`*.up.sql`): Apply schema changes
- **Down migrations** (`*.down.sql`): Rollback schema changes

## Available Migrations

| Version | Description | Files |
|---------|-------------|-------|
| 001 | Initial schema (users, trading_accounts, sessions, audit_logs) | `001_initial_schema.{up\|down}.sql` |

## Running Migrations

### Using Docker Compose

The schema is automatically applied when starting PostgreSQL with Docker Compose:

```bash
docker-compose up -d postgres
```

The initial schema (`001_initial_schema.up.sql`) is automatically loaded via the `docker-entrypoint-initdb.d` volume mount.

### Manual Migration

To manually apply migrations:

```bash
# Connect to PostgreSQL
psql -h localhost -U postgres -d eth_trading

# Apply migration
\i migrations/001_initial_schema.up.sql

# Rollback migration
\i migrations/001_initial_schema.down.sql
```

### Using Go Migration Tools

You can use tools like `golang-migrate` or `goose` for automated migration management:

**Install golang-migrate:**
```bash
brew install golang-migrate  # macOS
# or
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**Run migrations:**
```bash
migrate -database "postgres://postgres:postgres@localhost:5432/eth_trading?sslmode=disable" -path migrations up

# Rollback last migration
migrate -database "postgres://postgres:postgres@localhost:5432/eth_trading?sslmode=disable" -path migrations down 1
```

## Seed Data

For development and testing, use the seed data script:

```bash
psql -h localhost -U postgres -d eth_trading < migrations/seed_data.sql
```

## Creating New Migrations

1. Create a new migration file pair:
   ```bash
   # Example: 002_add_two_factor_auth.up.sql
   # Example: 002_add_two_factor_auth.down.sql
   ```

2. Write the migration:
   - **Up migration**: Add your schema changes
   - **Down migration**: Reverse those changes

3. Test both directions:
   ```bash
   # Apply
   psql -h localhost -U postgres -d eth_trading < migrations/002_add_two_factor_auth.up.sql

   # Rollback
   psql -h localhost -U postgres -d eth_trading < migrations/002_add_two_factor_auth.down.sql
   ```

## Best Practices

- **Always create both up and down migrations**
- **Test migrations on a development database first**
- **Make migrations atomic** - one logical change per migration
- **Never modify existing migrations** - create a new one instead
- **Use transactions** when possible for rollback safety
- **Add comments** explaining complex changes
- **Version control** all migration files

## Schema Verification

To verify the current schema:

```bash
# List all tables
psql -h localhost -U postgres -d eth_trading -c "\dt"

# Describe a specific table
psql -h localhost -U postgres -d eth_trading -c "\d users"

# Check indexes
psql -h localhost -U postgres -d eth_trading -c "\di"
```
