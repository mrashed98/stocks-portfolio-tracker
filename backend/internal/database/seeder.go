package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

type Seeder struct {
	db *DB
}

func NewSeeder(db *DB) *Seeder {
	return &Seeder{db: db}
}

// SeedDevelopmentData seeds the database with development data
func (s *Seeder) SeedDevelopmentData() error {
	log.Println("Starting database seeding...")

	// Check if data already exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if count > 0 {
		log.Println("Database already contains data, skipping seeding")
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Seed users
	userIDs, err := s.seedUsers(tx)
	if err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// Seed stocks
	stockIDs, err := s.seedStocks(tx)
	if err != nil {
		return fmt.Errorf("failed to seed stocks: %w", err)
	}

	// Seed strategies
	strategyIDs, err := s.seedStrategies(tx, userIDs[0])
	if err != nil {
		return fmt.Errorf("failed to seed strategies: %w", err)
	}

	// Seed strategy-stock relationships
	if err := s.seedStrategyStocks(tx, strategyIDs, stockIDs); err != nil {
		return fmt.Errorf("failed to seed strategy stocks: %w", err)
	}

	// Seed signals
	if err := s.seedSignals(tx, stockIDs); err != nil {
		return fmt.Errorf("failed to seed signals: %w", err)
	}

	// Seed portfolios
	portfolioIDs, err := s.seedPortfolios(tx, userIDs[0])
	if err != nil {
		return fmt.Errorf("failed to seed portfolios: %w", err)
	}

	// Seed positions
	if err := s.seedPositions(tx, portfolioIDs[0], stockIDs[:3], strategyIDs[0]); err != nil {
		return fmt.Errorf("failed to seed positions: %w", err)
	}

	// Seed NAV history
	if err := s.seedNAVHistory(tx, portfolioIDs[0]); err != nil {
		return fmt.Errorf("failed to seed NAV history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Database seeding completed successfully")
	return nil
}

func (s *Seeder) seedUsers(tx *sql.Tx) ([]uuid.UUID, error) {
	users := []struct {
		name     string
		email    string
		password string
	}{
		{"John Doe", "john@example.com", "password123"},
		{"Jane Smith", "jane@example.com", "password123"},
		{"Bob Johnson", "bob@example.com", "password123"},
	}

	var userIDs []uuid.UUID

	for _, user := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}

		var userID uuid.UUID
		err = tx.QueryRow(`
			INSERT INTO users (name, email, password_hash)
			VALUES ($1, $2, $3)
			RETURNING id
		`, user.name, user.email, string(hashedPassword)).Scan(&userID)

		if err != nil {
			return nil, fmt.Errorf("failed to insert user %s: %w", user.email, err)
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (s *Seeder) seedStocks(tx *sql.Tx) ([]uuid.UUID, error) {
	stocks := []struct {
		ticker   string
		name     string
		sector   string
		exchange string
	}{
		{"AAPL", "Apple Inc.", "Technology", "NASDAQ"},
		{"GOOGL", "Alphabet Inc.", "Technology", "NASDAQ"},
		{"MSFT", "Microsoft Corporation", "Technology", "NASDAQ"},
		{"AMZN", "Amazon.com Inc.", "Consumer Discretionary", "NASDAQ"},
		{"TSLA", "Tesla Inc.", "Consumer Discretionary", "NASDAQ"},
		{"NVDA", "NVIDIA Corporation", "Technology", "NASDAQ"},
		{"META", "Meta Platforms Inc.", "Technology", "NASDAQ"},
		{"JPM", "JPMorgan Chase & Co.", "Financial Services", "NYSE"},
		{"JNJ", "Johnson & Johnson", "Healthcare", "NYSE"},
		{"V", "Visa Inc.", "Financial Services", "NYSE"},
	}

	var stockIDs []uuid.UUID

	for _, stock := range stocks {
		var stockID uuid.UUID
		err := tx.QueryRow(`
			INSERT INTO stocks (ticker, name, sector, exchange)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, stock.ticker, stock.name, stock.sector, stock.exchange).Scan(&stockID)

		if err != nil {
			return nil, fmt.Errorf("failed to insert stock %s: %w", stock.ticker, err)
		}

		stockIDs = append(stockIDs, stockID)
	}

	return stockIDs, nil
}

func (s *Seeder) seedStrategies(tx *sql.Tx, userID uuid.UUID) ([]uuid.UUID, error) {
	strategies := []struct {
		name        string
		weightMode  string
		weightValue decimal.Decimal
	}{
		{"Growth Strategy", "percent", decimal.NewFromFloat(60.0)},
		{"Value Strategy", "percent", decimal.NewFromFloat(30.0)},
		{"Dividend Strategy", "percent", decimal.NewFromFloat(10.0)},
	}

	var strategyIDs []uuid.UUID

	for _, strategy := range strategies {
		var strategyID uuid.UUID
		err := tx.QueryRow(`
			INSERT INTO strategies (user_id, name, weight_mode, weight_value)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, userID, strategy.name, strategy.weightMode, strategy.weightValue).Scan(&strategyID)

		if err != nil {
			return nil, fmt.Errorf("failed to insert strategy %s: %w", strategy.name, err)
		}

		strategyIDs = append(strategyIDs, strategyID)
	}

	return strategyIDs, nil
}

func (s *Seeder) seedStrategyStocks(tx *sql.Tx, strategyIDs, stockIDs []uuid.UUID) error {
	// Growth Strategy - Tech stocks
	growthStocks := stockIDs[:6] // AAPL, GOOGL, MSFT, AMZN, TSLA, NVDA
	for _, stockID := range growthStocks {
		_, err := tx.Exec(`
			INSERT INTO strategy_stocks (strategy_id, stock_id, eligible)
			VALUES ($1, $2, $3)
		`, strategyIDs[0], stockID, true)
		if err != nil {
			return fmt.Errorf("failed to insert strategy stock: %w", err)
		}
	}

	// Value Strategy - Financial and Healthcare
	valueStocks := []uuid.UUID{stockIDs[7], stockIDs[8], stockIDs[9]} // JPM, JNJ, V
	for _, stockID := range valueStocks {
		_, err := tx.Exec(`
			INSERT INTO strategy_stocks (strategy_id, stock_id, eligible)
			VALUES ($1, $2, $3)
		`, strategyIDs[1], stockID, true)
		if err != nil {
			return fmt.Errorf("failed to insert strategy stock: %w", err)
		}
	}

	// Dividend Strategy - Mix of stable stocks
	dividendStocks := []uuid.UUID{stockIDs[0], stockIDs[2], stockIDs[7], stockIDs[8]} // AAPL, MSFT, JPM, JNJ
	for _, stockID := range dividendStocks {
		_, err := tx.Exec(`
			INSERT INTO strategy_stocks (strategy_id, stock_id, eligible)
			VALUES ($1, $2, $3)
		`, strategyIDs[2], stockID, true)
		if err != nil {
			return fmt.Errorf("failed to insert strategy stock: %w", err)
		}
	}

	return nil
}

func (s *Seeder) seedSignals(tx *sql.Tx, stockIDs []uuid.UUID) error {
	today := time.Now().Format("2006-01-02")

	signals := []struct {
		stockIndex int
		signal     string
	}{
		{0, "Buy"},   // AAPL
		{1, "Buy"},   // GOOGL
		{2, "Buy"},   // MSFT
		{3, "Hold"},  // AMZN
		{4, "Buy"},   // TSLA
		{5, "Buy"},   // NVDA
		{6, "Hold"},  // META
		{7, "Buy"},   // JPM
		{8, "Buy"},   // JNJ
		{9, "Buy"},   // V
	}

	for _, signal := range signals {
		_, err := tx.Exec(`
			INSERT INTO signals (stock_id, signal, date)
			VALUES ($1, $2, $3)
		`, stockIDs[signal.stockIndex], signal.signal, today)
		if err != nil {
			return fmt.Errorf("failed to insert signal: %w", err)
		}
	}

	return nil
}

func (s *Seeder) seedPortfolios(tx *sql.Tx, userID uuid.UUID) ([]uuid.UUID, error) {
	portfolios := []struct {
		name            string
		totalInvestment decimal.Decimal
	}{
		{"My Growth Portfolio", decimal.NewFromFloat(100000.0)},
		{"Conservative Portfolio", decimal.NewFromFloat(50000.0)},
	}

	var portfolioIDs []uuid.UUID

	for _, portfolio := range portfolios {
		var portfolioID uuid.UUID
		err := tx.QueryRow(`
			INSERT INTO portfolios (user_id, name, total_investment)
			VALUES ($1, $2, $3)
			RETURNING id
		`, userID, portfolio.name, portfolio.totalInvestment).Scan(&portfolioID)

		if err != nil {
			return nil, fmt.Errorf("failed to insert portfolio %s: %w", portfolio.name, err)
		}

		portfolioIDs = append(portfolioIDs, portfolioID)
	}

	return portfolioIDs, nil
}

func (s *Seeder) seedPositions(tx *sql.Tx, portfolioID uuid.UUID, stockIDs []uuid.UUID, strategyID uuid.UUID) error {
	positions := []struct {
		stockIndex      int
		quantity        int
		entryPrice      decimal.Decimal
		allocationValue decimal.Decimal
	}{
		{0, 200, decimal.NewFromFloat(150.25), decimal.NewFromFloat(30050.0)}, // AAPL
		{1, 100, decimal.NewFromFloat(140.50), decimal.NewFromFloat(14050.0)}, // GOOGL
		{2, 150, decimal.NewFromFloat(380.75), decimal.NewFromFloat(57112.5)}, // MSFT
	}

	strategyContrib := fmt.Sprintf(`{"%s": %%f}`, strategyID.String())

	for _, position := range positions {
		contrib := fmt.Sprintf(strategyContrib, position.allocationValue)
		_, err := tx.Exec(`
			INSERT INTO positions (portfolio_id, stock_id, quantity, entry_price, allocation_value, strategy_contrib)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, portfolioID, stockIDs[position.stockIndex], position.quantity, position.entryPrice, position.allocationValue, contrib)

		if err != nil {
			return fmt.Errorf("failed to insert position: %w", err)
		}
	}

	return nil
}

func (s *Seeder) seedNAVHistory(tx *sql.Tx, portfolioID uuid.UUID) error {
	baseNAV := decimal.NewFromFloat(101212.5)
	now := time.Now()

	// Create 30 days of NAV history
	for i := 29; i >= 0; i-- {
		timestamp := now.AddDate(0, 0, -i)
		
		// Simulate some price movement
		variation := decimal.NewFromFloat(float64(i%7-3) * 500.0) // Random variation
		nav := baseNAV.Add(variation)
		pnl := nav.Sub(decimal.NewFromFloat(100000.0)) // Initial investment was 100k
		
		// Simple drawdown calculation (percentage from peak)
		drawdown := decimal.NewFromFloat(0.0)
		if i > 15 { // Simulate some drawdown in the middle period
			drawdown = decimal.NewFromFloat(-2.5)
		}

		_, err := tx.Exec(`
			INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown)
			VALUES ($1, $2, $3, $4, $5)
		`, portfolioID, timestamp, nav, pnl, drawdown)

		if err != nil {
			return fmt.Errorf("failed to insert NAV history: %w", err)
		}
	}

	return nil
}