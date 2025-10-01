package database

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"portfolio-app/config"
)

func setupTestDB(t *testing.T) *DB {
	if os.Getenv("TEST_DB_URL") == "" {
		t.Skip("TEST_DB_URL not set, skipping database tests")
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "portfolio_test",
			User:     "test_user",
			Password: "test_pass",
			SSLMode:  "disable",
		},
	}

	db, err := NewConnection(cfg)
	require.NoError(t, err)

	// Run migrations
	err = db.RunMigrations("../../migrations")
	require.NoError(t, err)

	return db
}

func TestUsersTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test user insertion
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	require.NoError(t, err)

	var userID uuid.UUID
	err = db.QueryRow(`
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, "Test User", "test@example.com", string(hashedPassword)).Scan(&userID)
	
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, userID)

	// Test unique email constraint
	_, err = db.Exec(`
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
	`, "Another User", "test@example.com", string(hashedPassword))
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value")

	// Test user retrieval
	var name, email string
	err = db.QueryRow(`
		SELECT name, email FROM users WHERE id = $1
	`, userID).Scan(&name, &email)
	
	require.NoError(t, err)
	assert.Equal(t, "Test User", name)
	assert.Equal(t, "test@example.com", email)
}

func TestStrategiesTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a user first
	userID := createTestUser(t, db)

	// Test strategy insertion
	var strategyID uuid.UUID
	err := db.QueryRow(`
		INSERT INTO strategies (user_id, name, weight_mode, weight_value)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, "Test Strategy", "percent", decimal.NewFromFloat(50.0)).Scan(&strategyID)
	
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, strategyID)

	// Test weight mode constraint
	_, err = db.Exec(`
		INSERT INTO strategies (user_id, name, weight_mode, weight_value)
		VALUES ($1, $2, $3, $4)
	`, userID, "Invalid Strategy", "invalid_mode", decimal.NewFromFloat(50.0))
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check constraint")

	// Test positive weight value constraint
	_, err = db.Exec(`
		INSERT INTO strategies (user_id, name, weight_mode, weight_value)
		VALUES ($1, $2, $3, $4)
	`, userID, "Negative Strategy", "percent", decimal.NewFromFloat(-10.0))
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check constraint")

	// Test foreign key constraint
	_, err = db.Exec(`
		INSERT INTO strategies (user_id, name, weight_mode, weight_value)
		VALUES ($1, $2, $3, $4)
	`, uuid.New(), "Orphan Strategy", "percent", decimal.NewFromFloat(50.0))
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "foreign key constraint")
}

func TestStocksTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test stock insertion
	var stockID uuid.UUID
	err := db.QueryRow(`
		INSERT INTO stocks (ticker, name, sector, exchange)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "AAPL", "Apple Inc.", "Technology", "NASDAQ").Scan(&stockID)
	
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, stockID)

	// Test unique ticker constraint
	_, err = db.Exec(`
		INSERT INTO stocks (ticker, name, sector, exchange)
		VALUES ($1, $2, $3, $4)
	`, "AAPL", "Another Apple", "Technology", "NASDAQ")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value")
}

func TestStrategyStocksTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)
	strategyID := createTestStrategy(t, db, userID)
	stockID := createTestStock(t, db)

	// Test strategy-stock relationship insertion
	_, err := db.Exec(`
		INSERT INTO strategy_stocks (strategy_id, stock_id, eligible)
		VALUES ($1, $2, $3)
	`, strategyID, stockID, true)
	
	require.NoError(t, err)

	// Test primary key constraint (duplicate relationship)
	_, err = db.Exec(`
		INSERT INTO strategy_stocks (strategy_id, stock_id, eligible)
		VALUES ($1, $2, $3)
	`, strategyID, stockID, false)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value")

	// Test foreign key constraints
	_, err = db.Exec(`
		INSERT INTO strategy_stocks (strategy_id, stock_id, eligible)
		VALUES ($1, $2, $3)
	`, uuid.New(), stockID, true)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "foreign key constraint")
}

func TestSignalsTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	stockID := createTestStock(t, db)

	// Test signal insertion
	_, err := db.Exec(`
		INSERT INTO signals (stock_id, signal, date)
		VALUES ($1, $2, $3)
	`, stockID, "Buy", "2024-01-01")
	
	require.NoError(t, err)

	// Test signal constraint
	_, err = db.Exec(`
		INSERT INTO signals (stock_id, signal, date)
		VALUES ($1, $2, $3)
	`, stockID, "InvalidSignal", "2024-01-02")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check constraint")

	// Test primary key constraint (same stock, same date)
	_, err = db.Exec(`
		INSERT INTO signals (stock_id, signal, date)
		VALUES ($1, $2, $3)
	`, stockID, "Hold", "2024-01-01")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value")
}

func TestPortfoliosTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)

	// Test portfolio insertion
	var portfolioID uuid.UUID
	err := db.QueryRow(`
		INSERT INTO portfolios (user_id, name, total_investment)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, "Test Portfolio", decimal.NewFromFloat(10000.0)).Scan(&portfolioID)
	
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, portfolioID)

	// Test positive investment constraint
	_, err = db.Exec(`
		INSERT INTO portfolios (user_id, name, total_investment)
		VALUES ($1, $2, $3)
	`, userID, "Invalid Portfolio", decimal.NewFromFloat(-1000.0))
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check constraint")
}

func TestPositionsTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)
	portfolioID := createTestPortfolio(t, db, userID)
	stockID := createTestStock(t, db)
	strategyID := createTestStrategy(t, db, userID)

	// Test position insertion
	strategyContrib := `{"` + strategyID.String() + `": 5000.0}`
	_, err := db.Exec(`
		INSERT INTO positions (portfolio_id, stock_id, quantity, entry_price, allocation_value, strategy_contrib)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, portfolioID, stockID, 100, decimal.NewFromFloat(50.0), decimal.NewFromFloat(5000.0), strategyContrib)
	
	require.NoError(t, err)

	// Test positive quantity constraint
	_, err = db.Exec(`
		INSERT INTO positions (portfolio_id, stock_id, quantity, entry_price, allocation_value, strategy_contrib)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, portfolioID, createTestStock(t, db), -10, decimal.NewFromFloat(50.0), decimal.NewFromFloat(5000.0), strategyContrib)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check constraint")

	// Test positive entry price constraint
	_, err = db.Exec(`
		INSERT INTO positions (portfolio_id, stock_id, quantity, entry_price, allocation_value, strategy_contrib)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, portfolioID, createTestStock(t, db), 100, decimal.NewFromFloat(-50.0), decimal.NewFromFloat(5000.0), strategyContrib)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check constraint")
}

func TestNAVHistoryTableSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)
	portfolioID := createTestPortfolio(t, db, userID)

	// Test NAV history insertion
	_, err := db.Exec(`
		INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown)
		VALUES ($1, $2, $3, $4, $5)
	`, portfolioID, "2024-01-01 10:00:00", decimal.NewFromFloat(10500.0), decimal.NewFromFloat(500.0), decimal.NewFromFloat(-2.5))
	
	require.NoError(t, err)

	// Test non-negative NAV constraint
	_, err = db.Exec(`
		INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown)
		VALUES ($1, $2, $3, $4, $5)
	`, portfolioID, "2024-01-02 10:00:00", decimal.NewFromFloat(-1000.0), decimal.NewFromFloat(-1000.0), decimal.NewFromFloat(-10.0))
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check constraint")

	// Test primary key constraint (same portfolio, same timestamp)
	_, err = db.Exec(`
		INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown)
		VALUES ($1, $2, $3, $4, $5)
	`, portfolioID, "2024-01-01 10:00:00", decimal.NewFromFloat(10600.0), decimal.NewFromFloat(600.0), decimal.NewFromFloat(-1.5))
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key value")
}

// Helper functions for creating test data
func createTestUser(t *testing.T, db *DB) uuid.UUID {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	require.NoError(t, err)

	var userID uuid.UUID
	err = db.QueryRow(`
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, "Test User", "test"+uuid.New().String()+"@example.com", string(hashedPassword)).Scan(&userID)
	
	require.NoError(t, err)
	return userID
}

func createTestStrategy(t *testing.T, db *DB, userID uuid.UUID) uuid.UUID {
	var strategyID uuid.UUID
	err := db.QueryRow(`
		INSERT INTO strategies (user_id, name, weight_mode, weight_value)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, "Test Strategy "+uuid.New().String(), "percent", decimal.NewFromFloat(50.0)).Scan(&strategyID)
	
	require.NoError(t, err)
	return strategyID
}

func createTestStock(t *testing.T, db *DB) uuid.UUID {
	ticker := "TST" + uuid.New().String()[:4]
	var stockID uuid.UUID
	err := db.QueryRow(`
		INSERT INTO stocks (ticker, name, sector, exchange)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, ticker, "Test Stock", "Technology", "NASDAQ").Scan(&stockID)
	
	require.NoError(t, err)
	return stockID
}

func createTestPortfolio(t *testing.T, db *DB, userID uuid.UUID) uuid.UUID {
	var portfolioID uuid.UUID
	err := db.QueryRow(`
		INSERT INTO portfolios (user_id, name, total_investment)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, "Test Portfolio "+uuid.New().String(), decimal.NewFromFloat(10000.0)).Scan(&portfolioID)
	
	require.NoError(t, err)
	return portfolioID
}