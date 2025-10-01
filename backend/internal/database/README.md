# Database Package

This package provides database connection utilities, migrations, and seeding functionality for the Portfolio Web App.

## Features

- PostgreSQL connection with connection pooling
- Database migration management using golang-migrate
- Development data seeding
- Comprehensive schema validation tests
- Health check functionality

## Database Schema

The application uses the following core tables:

- **users**: User accounts with authentication
- **strategies**: Investment strategies with weight modes
- **stocks**: Stock information and metadata
- **strategy_stocks**: Many-to-many relationship between strategies and stocks
- **signals**: Buy/Hold signals for stocks
- **portfolios**: User portfolios with investment amounts
- **positions**: Portfolio positions with entry prices and allocations
- **nav_history**: Historical NAV data for performance tracking

## Usage

### Basic Connection

```go
import "portfolio-app/backend/internal/database"

cfg := &config.Config{...}
db, err := database.NewConnection(cfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### Running Migrations

```go
err := db.RunMigrations("./migrations")
if err != nil {
    log.Fatal(err)
}
```

### Seeding Development Data

```go
seeder := database.NewSeeder(db)
err := seeder.SeedDevelopmentData()
if err != nil {
    log.Fatal(err)
}
```

## Migration Files

Migration files are located in `backend/migrations/` and follow the naming convention:
- `XXXXXX_description.up.sql` - Forward migration
- `XXXXXX_description.down.sql` - Rollback migration

### Creating New Migrations

Use the migrate tool to create new migration files:

```bash
migrate create -ext sql -dir ./migrations -seq add_new_table
```

## Testing

The package includes comprehensive tests for:
- Database connection and pooling
- Schema validation and constraints
- Foreign key relationships
- Data integrity checks

Run tests with:

```bash
# Unit tests only
go test ./internal/database/

# Integration tests (requires test database)
export TEST_DB_URL="postgres://test_user:test_pass@localhost:5432/portfolio_test?sslmode=disable"
go test -v ./internal/database/
```

## Configuration

The database connection is configured through environment variables:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=portfolio_db
DB_USER=portfolio_user
DB_PASSWORD=portfolio_pass
DB_SSLMODE=disable
```

## Connection Pooling

The connection pool is configured with:
- Max open connections: 25
- Max idle connections: 5
- Connection max lifetime: 5 minutes

These settings can be adjusted in the `NewConnection` function based on your application's needs.

## Development Data

The seeder creates sample data including:
- 3 test users with hashed passwords
- 10 popular stocks (AAPL, GOOGL, MSFT, etc.)
- 3 sample strategies (Growth, Value, Dividend)
- Stock assignments to strategies
- Buy/Hold signals for all stocks
- 2 sample portfolios with positions
- 30 days of NAV history

This data is automatically seeded when running in development mode.

## Error Handling

The package provides structured error handling for:
- Connection failures
- Migration errors
- Constraint violations
- Foreign key violations
- Data validation errors

All errors are wrapped with context to provide meaningful error messages.