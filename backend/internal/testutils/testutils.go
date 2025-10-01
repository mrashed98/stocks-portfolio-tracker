package testutils

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

// SetupTestDB sets up a test database connection
func SetupTestDB(t *testing.T) *sqlx.DB {
	if os.Getenv("TEST_DB_URL") == "" {
		t.Skip("TEST_DB_URL not set, skipping database tests")
	}

	db, err := sqlx.Connect("postgres", os.Getenv("TEST_DB_URL"))
	require.NoError(t, err)

	// Clean up any existing test data
	CleanupTestDB(t, db)

	return db
}

// CleanupTestDB cleans up test database
func CleanupTestDB(t *testing.T, db *sqlx.DB) {
	// Delete in reverse order of dependencies
	_, err := db.Exec("DELETE FROM nav_history")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM positions")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM portfolios")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM signals")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM strategy_stocks")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM strategies")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM stocks")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM users")
	require.NoError(t, err)
}

// CreateTestStock creates a test stock in the database
func CreateTestStock(t *testing.T, db *sqlx.DB, stockID uuid.UUID, ticker, name, sector string) {
	query := `
		INSERT INTO stocks (id, ticker, name, sector, exchange, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING`

	_, err := db.Exec(query, stockID, ticker, name, sector, "NASDAQ")
	require.NoError(t, err)
}