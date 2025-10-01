package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DatabaseIntegrationTestSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *sqlx.DB
	ctx       context.Context
}

func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	
	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)
	suite.container = container
	
	// Get container connection details
	host, err := container.Host(suite.ctx)
	suite.Require().NoError(err)
	
	port, err := container.MappedPort(suite.ctx, "5432")
	suite.Require().NoError(err)
	
	// Connect to database
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err := sqlx.Connect("postgres", dsn)
	suite.Require().NoError(err)
	suite.db = db
	
	// Run migrations
	err = suite.runMigrations(dsn)
	suite.Require().NoError(err)
}

func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
	}
}

func (suite *DatabaseIntegrationTestSuite) SetupTest() {
	// Clean up tables before each test
	tables := []string{
		"nav_history", "positions", "portfolios", "signals", 
		"strategy_stocks", "strategies", "stocks", "users",
	}
	
	for _, table := range tables {
		_, err := suite.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		suite.Require().NoError(err)
	}
}

func (suite *DatabaseIntegrationTestSuite) runMigrations(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver)
	if err != nil {
		return err
	}
	
	return m.Up()
}

func (suite *DatabaseIntegrationTestSuite) TestDatabaseConnection() {
	err := suite.db.Ping()
	suite.NoError(err)
}

func (suite *DatabaseIntegrationTestSuite) TestUserCRUD() {
	// Test user creation
	userID := "550e8400-e29b-41d4-a716-446655440000"
	_, err := suite.db.Exec(`
		INSERT INTO users (id, name, email, password_hash) 
		VALUES ($1, $2, $3, $4)`,
		userID, "Test User", "test@example.com", "hashedpassword")
	suite.NoError(err)
	
	// Test user retrieval
	var user struct {
		ID           string    `db:"id"`
		Name         string    `db:"name"`
		Email        string    `db:"email"`
		PasswordHash string    `db:"password_hash"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}
	
	err = suite.db.Get(&user, "SELECT * FROM users WHERE id = $1", userID)
	suite.NoError(err)
	suite.Equal("Test User", user.Name)
	suite.Equal("test@example.com", user.Email)
	
	// Test user update
	_, err = suite.db.Exec("UPDATE users SET name = $1 WHERE id = $2", "Updated User", userID)
	suite.NoError(err)
	
	err = suite.db.Get(&user, "SELECT * FROM users WHERE id = $1", userID)
	suite.NoError(err)
	suite.Equal("Updated User", user.Name)
	
	// Test user deletion
	_, err = suite.db.Exec("DELETE FROM users WHERE id = $1", userID)
	suite.NoError(err)
	
	err = suite.db.Get(&user, "SELECT * FROM users WHERE id = $1", userID)
	suite.Error(err)
}

func (suite *DatabaseIntegrationTestSuite) TestStrategyWithStocks() {
	// Create user
	userID := "550e8400-e29b-41d4-a716-446655440000"
	_, err := suite.db.Exec(`
		INSERT INTO users (id, name, email, password_hash) 
		VALUES ($1, $2, $3, $4)`,
		userID, "Test User", "test@example.com", "hashedpassword")
	suite.NoError(err)
	
	// Create strategy
	strategyID := "550e8400-e29b-41d4-a716-446655440001"
	_, err = suite.db.Exec(`
		INSERT INTO strategies (id, user_id, name, weight_mode, weight_value) 
		VALUES ($1, $2, $3, $4, $5)`,
		strategyID, userID, "Growth Strategy", "percent", 60.0)
	suite.NoError(err)
	
	// Create stocks
	stockID1 := "550e8400-e29b-41d4-a716-446655440002"
	stockID2 := "550e8400-e29b-41d4-a716-446655440003"
	
	_, err = suite.db.Exec(`
		INSERT INTO stocks (id, ticker, name, sector) 
		VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)`,
		stockID1, "AAPL", "Apple Inc.", "Technology",
		stockID2, "GOOGL", "Alphabet Inc.", "Technology")
	suite.NoError(err)
	
	// Create strategy-stock relationships
	_, err = suite.db.Exec(`
		INSERT INTO strategy_stocks (strategy_id, stock_id, eligible) 
		VALUES ($1, $2, $3), ($4, $5, $6)`,
		strategyID, stockID1, true,
		strategyID, stockID2, false)
	suite.NoError(err)
	
	// Test complex query: Get strategy with eligible stocks
	var results []struct {
		StrategyName string `db:"strategy_name"`
		StockTicker  string `db:"stock_ticker"`
		Eligible     bool   `db:"eligible"`
	}
	
	err = suite.db.Select(&results, `
		SELECT s.name as strategy_name, st.ticker as stock_ticker, ss.eligible
		FROM strategies s
		JOIN strategy_stocks ss ON s.id = ss.strategy_id
		JOIN stocks st ON ss.stock_id = st.id
		WHERE s.id = $1 AND ss.eligible = true
		ORDER BY st.ticker`, strategyID)
	suite.NoError(err)
	
	suite.Len(results, 1)
	suite.Equal("Growth Strategy", results[0].StrategyName)
	suite.Equal("AAPL", results[0].StockTicker)
	suite.True(results[0].Eligible)
}

func (suite *DatabaseIntegrationTestSuite) TestPortfolioWithPositions() {
	// Create user
	userID := "550e8400-e29b-41d4-a716-446655440000"
	_, err := suite.db.Exec(`
		INSERT INTO users (id, name, email, password_hash) 
		VALUES ($1, $2, $3, $4)`,
		userID, "Test User", "test@example.com", "hashedpassword")
	suite.NoError(err)
	
	// Create portfolio
	portfolioID := "550e8400-e29b-41d4-a716-446655440001"
	_, err = suite.db.Exec(`
		INSERT INTO portfolios (id, user_id, name, total_investment) 
		VALUES ($1, $2, $3, $4)`,
		portfolioID, userID, "Test Portfolio", 10000.00)
	suite.NoError(err)
	
	// Create stocks
	stockID1 := "550e8400-e29b-41d4-a716-446655440002"
	stockID2 := "550e8400-e29b-41d4-a716-446655440003"
	
	_, err = suite.db.Exec(`
		INSERT INTO stocks (id, ticker, name) 
		VALUES ($1, $2, $3), ($4, $5, $6)`,
		stockID1, "AAPL", "Apple Inc.",
		stockID2, "GOOGL", "Alphabet Inc.")
	suite.NoError(err)
	
	// Create positions
	_, err = suite.db.Exec(`
		INSERT INTO positions (portfolio_id, stock_id, quantity, entry_price, allocation_value, strategy_contrib) 
		VALUES ($1, $2, $3, $4, $5, $6), ($7, $8, $9, $10, $11, $12)`,
		portfolioID, stockID1, 33, 150.50, 5000.00, `{"strategy1": 5000}`,
		portfolioID, stockID2, 1, 2800.75, 3000.00, `{"strategy1": 3000}`)
	suite.NoError(err)
	
	// Test portfolio value calculation
	var totalValue float64
	err = suite.db.Get(&totalValue, `
		SELECT SUM(allocation_value) 
		FROM positions 
		WHERE portfolio_id = $1`, portfolioID)
	suite.NoError(err)
	suite.Equal(8000.00, totalValue)
	
	// Test position details query
	var positions []struct {
		Ticker          string  `db:"ticker"`
		Quantity        int     `db:"quantity"`
		EntryPrice      float64 `db:"entry_price"`
		AllocationValue float64 `db:"allocation_value"`
	}
	
	err = suite.db.Select(&positions, `
		SELECT s.ticker, p.quantity, p.entry_price, p.allocation_value
		FROM positions p
		JOIN stocks s ON p.stock_id = s.id
		WHERE p.portfolio_id = $1
		ORDER BY s.ticker`, portfolioID)
	suite.NoError(err)
	
	suite.Len(positions, 2)
	suite.Equal("AAPL", positions[0].Ticker)
	suite.Equal(33, positions[0].Quantity)
	suite.Equal(150.50, positions[0].EntryPrice)
}

func (suite *DatabaseIntegrationTestSuite) TestNAVHistoryTracking() {
	// Create user and portfolio
	userID := "550e8400-e29b-41d4-a716-446655440000"
	portfolioID := "550e8400-e29b-41d4-a716-446655440001"
	
	_, err := suite.db.Exec(`
		INSERT INTO users (id, name, email, password_hash) 
		VALUES ($1, $2, $3, $4)`,
		userID, "Test User", "test@example.com", "hashedpassword")
	suite.NoError(err)
	
	_, err = suite.db.Exec(`
		INSERT INTO portfolios (id, user_id, name, total_investment) 
		VALUES ($1, $2, $3, $4)`,
		portfolioID, userID, "Test Portfolio", 10000.00)
	suite.NoError(err)
	
	// Insert NAV history entries
	timestamps := []time.Time{
		time.Now().Add(-48 * time.Hour),
		time.Now().Add(-24 * time.Hour),
		time.Now(),
	}
	
	navValues := []float64{10000.00, 10500.00, 9800.00}
	pnlValues := []float64{0.00, 500.00, -200.00}
	drawdowns := []float64{0.00, 0.00, 6.67}
	
	for i, timestamp := range timestamps {
		_, err = suite.db.Exec(`
			INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown) 
			VALUES ($1, $2, $3, $4, $5)`,
			portfolioID, timestamp, navValues[i], pnlValues[i], drawdowns[i])
		suite.NoError(err)
	}
	
	// Test NAV history retrieval
	var history []struct {
		Timestamp time.Time `db:"timestamp"`
		NAV       float64   `db:"nav"`
		PnL       float64   `db:"pnl"`
		Drawdown  float64   `db:"drawdown"`
	}
	
	err = suite.db.Select(&history, `
		SELECT timestamp, nav, pnl, drawdown 
		FROM nav_history 
		WHERE portfolio_id = $1 
		ORDER BY timestamp`, portfolioID)
	suite.NoError(err)
	
	suite.Len(history, 3)
	suite.Equal(10000.00, history[0].NAV)
	suite.Equal(10500.00, history[1].NAV)
	suite.Equal(9800.00, history[2].NAV)
	
	// Test performance metrics calculation
	var maxNAV, minNAV, currentNAV float64
	err = suite.db.Get(&maxNAV, `
		SELECT MAX(nav) FROM nav_history WHERE portfolio_id = $1`, portfolioID)
	suite.NoError(err)
	
	err = suite.db.Get(&minNAV, `
		SELECT MIN(nav) FROM nav_history WHERE portfolio_id = $1`, portfolioID)
	suite.NoError(err)
	
	err = suite.db.Get(&currentNAV, `
		SELECT nav FROM nav_history 
		WHERE portfolio_id = $1 
		ORDER BY timestamp DESC LIMIT 1`, portfolioID)
	suite.NoError(err)
	
	suite.Equal(10500.00, maxNAV)
	suite.Equal(9800.00, minNAV)
	suite.Equal(9800.00, currentNAV)
}

func (suite *DatabaseIntegrationTestSuite) TestSignalManagement() {
	// Create stock
	stockID := "550e8400-e29b-41d4-a716-446655440000"
	_, err := suite.db.Exec(`
		INSERT INTO stocks (id, ticker, name) 
		VALUES ($1, $2, $3)`,
		stockID, "AAPL", "Apple Inc.")
	suite.NoError(err)
	
	// Insert signals over time
	dates := []time.Time{
		time.Now().Add(-72 * time.Hour),
		time.Now().Add(-48 * time.Hour),
		time.Now().Add(-24 * time.Hour),
	}
	
	signals := []string{"Buy", "Hold", "Buy"}
	
	for i, date := range dates {
		_, err = suite.db.Exec(`
			INSERT INTO signals (stock_id, signal, date) 
			VALUES ($1, $2, $3)`,
			stockID, signals[i], date.Format("2006-01-02"))
		suite.NoError(err)
	}
	
	// Test latest signal retrieval
	var latestSignal string
	err = suite.db.Get(&latestSignal, `
		SELECT signal FROM signals 
		WHERE stock_id = $1 
		ORDER BY date DESC LIMIT 1`, stockID)
	suite.NoError(err)
	
	suite.Equal("Buy", latestSignal)
	
	// Test signal history
	var signalHistory []struct {
		Signal string    `db:"signal"`
		Date   time.Time `db:"date"`
	}
	
	err = suite.db.Select(&signalHistory, `
		SELECT signal, date FROM signals 
		WHERE stock_id = $1 
		ORDER BY date DESC`, stockID)
	suite.NoError(err)
	
	suite.Len(signalHistory, 3)
	suite.Equal("Buy", signalHistory[0].Signal)
	suite.Equal("Hold", signalHistory[1].Signal)
	suite.Equal("Buy", signalHistory[2].Signal)
}

func (suite *DatabaseIntegrationTestSuite) TestTransactionRollback() {
	// Test transaction rollback on error
	tx, err := suite.db.Beginx()
	suite.NoError(err)
	
	// Insert user
	userID := "550e8400-e29b-41d4-a716-446655440000"
	_, err = tx.Exec(`
		INSERT INTO users (id, name, email, password_hash) 
		VALUES ($1, $2, $3, $4)`,
		userID, "Test User", "test@example.com", "hashedpassword")
	suite.NoError(err)
	
	// Try to insert duplicate user (should fail)
	_, err = tx.Exec(`
		INSERT INTO users (id, name, email, password_hash) 
		VALUES ($1, $2, $3, $4)`,
		userID, "Duplicate User", "test@example.com", "hashedpassword")
	suite.Error(err)
	
	// Rollback transaction
	err = tx.Rollback()
	suite.NoError(err)
	
	// Verify user was not inserted
	var count int
	err = suite.db.Get(&count, "SELECT COUNT(*) FROM users WHERE id = $1", userID)
	suite.NoError(err)
	suite.Equal(0, count)
}

func (suite *DatabaseIntegrationTestSuite) TestConcurrentAccess() {
	// Test concurrent portfolio updates
	userID := "550e8400-e29b-41d4-a716-446655440000"
	portfolioID := "550e8400-e29b-41d4-a716-446655440001"
	
	// Setup
	_, err := suite.db.Exec(`
		INSERT INTO users (id, name, email, password_hash) 
		VALUES ($1, $2, $3, $4)`,
		userID, "Test User", "test@example.com", "hashedpassword")
	suite.NoError(err)
	
	_, err = suite.db.Exec(`
		INSERT INTO portfolios (id, user_id, name, total_investment) 
		VALUES ($1, $2, $3, $4)`,
		portfolioID, userID, "Test Portfolio", 10000.00)
	suite.NoError(err)
	
	// Simulate concurrent NAV updates
	done := make(chan bool, 2)
	
	updateNAV := func(nav float64) {
		_, err := suite.db.Exec(`
			INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown) 
			VALUES ($1, $2, $3, $4, $5)`,
			portfolioID, time.Now(), nav, nav-10000, 0.0)
		assert.NoError(suite.T(), err)
		done <- true
	}
	
	go updateNAV(10100.00)
	go updateNAV(10200.00)
	
	// Wait for both updates
	<-done
	<-done
	
	// Verify both updates were successful
	var count int
	err = suite.db.Get(&count, `
		SELECT COUNT(*) FROM nav_history WHERE portfolio_id = $1`, portfolioID)
	suite.NoError(err)
	suite.Equal(2, count)
}

func TestDatabaseIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationTestSuite))
}

// Benchmark tests for database operations
func BenchmarkDatabaseOperations(b *testing.B) {
	ctx := context.Background()
	
	// Setup container (simplified for benchmark)
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "benchdb",
			"POSTGRES_USER":     "benchuser",
			"POSTGRES_PASSWORD": "benchpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer container.Terminate(ctx)
	
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://benchuser:benchpass@%s:%s/benchdb?sslmode=disable", host, port.Port())
	
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	
	// Create simple table for benchmark
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bench_users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255),
			email VARCHAR(255)
		)`)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	
	b.Run("InsertUser", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := db.Exec("INSERT INTO bench_users (name, email) VALUES ($1, $2)",
				fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("SelectUser", func(b *testing.B) {
		// Insert some test data first
		for i := 0; i < 1000; i++ {
			db.Exec("INSERT INTO bench_users (name, email) VALUES ($1, $2)",
				fmt.Sprintf("BenchUser%d", i), fmt.Sprintf("benchuser%d@example.com", i))
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var user struct {
				ID    int    `db:"id"`
				Name  string `db:"name"`
				Email string `db:"email"`
			}
			err := db.Get(&user, "SELECT * FROM bench_users WHERE id = $1", i%1000+1)
			if err != nil && err != sql.ErrNoRows {
				b.Fatal(err)
			}
		}
	})
}