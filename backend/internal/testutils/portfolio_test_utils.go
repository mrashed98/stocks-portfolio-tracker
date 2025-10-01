package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"portfolio-app/internal/models"
)

// CreateTestPortfolio creates a test portfolio in the database
func CreateTestPortfolio(t *testing.T, db *sqlx.DB, userID uuid.UUID, name string, totalInvestment decimal.Decimal) *models.Portfolio {
	portfolio := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            name,
		TotalInvestment: totalInvestment,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	query := `
		INSERT INTO portfolios (id, user_id, name, total_investment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Exec(query, portfolio.ID, portfolio.UserID, portfolio.Name, 
		portfolio.TotalInvestment, portfolio.CreatedAt, portfolio.UpdatedAt)
	require.NoError(t, err)

	return portfolio
}

// CreateTestPosition creates a test position in the database
func CreateTestPosition(t *testing.T, db *sqlx.DB, portfolioID, stockID uuid.UUID, quantity int, entryPrice, allocationValue decimal.Decimal) *models.Position {
	position := &models.Position{
		PortfolioID:     portfolioID,
		StockID:         stockID,
		Quantity:        quantity,
		EntryPrice:      entryPrice,
		AllocationValue: allocationValue,
		StrategyContrib: []byte(`{"strategy1": ` + allocationValue.String() + `}`),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	query := `
		INSERT INTO positions (portfolio_id, stock_id, quantity, entry_price, allocation_value, strategy_contrib, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := db.Exec(query, position.PortfolioID, position.StockID, position.Quantity,
		position.EntryPrice, position.AllocationValue, position.StrategyContrib,
		position.CreatedAt, position.UpdatedAt)
	require.NoError(t, err)

	return position
}

// CreateTestNAVHistory creates a test NAV history entry in the database
func CreateTestNAVHistory(t *testing.T, db *sqlx.DB, portfolioID uuid.UUID, timestamp time.Time, nav, pnl decimal.Decimal, drawdown *decimal.Decimal) *models.NAVHistory {
	navHistory := &models.NAVHistory{
		PortfolioID: portfolioID,
		Timestamp:   timestamp,
		NAV:         nav,
		PnL:         pnl,
		Drawdown:    drawdown,
		CreatedAt:   time.Now(),
	}

	query := `
		INSERT INTO nav_history (portfolio_id, timestamp, nav, pnl, drawdown, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Exec(query, navHistory.PortfolioID, navHistory.Timestamp, navHistory.NAV,
		navHistory.PnL, navHistory.Drawdown, navHistory.CreatedAt)
	require.NoError(t, err)

	return navHistory
}

// CreateTestPortfolioWithPositions creates a test portfolio with positions
func CreateTestPortfolioWithPositions(t *testing.T, db *sqlx.DB, userID uuid.UUID, name string, totalInvestment decimal.Decimal, stockIDs []uuid.UUID) *models.Portfolio {
	// Create portfolio
	portfolio := CreateTestPortfolio(t, db, userID, name, totalInvestment)

	// Create positions
	allocationPerStock := totalInvestment.Div(decimal.NewFromInt(int64(len(stockIDs))))
	for i, stockID := range stockIDs {
		// Create test stock if it doesn't exist
		CreateTestStock(t, db, stockID, fmt.Sprintf("STOCK%d", i+1), fmt.Sprintf("Test Stock %d", i+1), "Technology")
		
		// Create position
		CreateTestPosition(t, db, portfolio.ID, stockID, 100, decimal.NewFromFloat(100.0), allocationPerStock)
	}

	// Create initial NAV history
	CreateTestNAVHistory(t, db, portfolio.ID, portfolio.CreatedAt, totalInvestment, decimal.Zero, nil)

	return portfolio
}

// CreateTestUser creates a test user and returns the user ID
func CreateTestUser(t *testing.T, db *sqlx.DB, email, name string) uuid.UUID {
	userID := uuid.New()
	
	query := `
		INSERT INTO users (id, email, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(query, userID, email, name, time.Now(), time.Now())
	require.NoError(t, err)

	return userID
}

// CreateTestStrategy creates a test strategy
func CreateTestStrategy(t *testing.T, db *sqlx.DB, userID uuid.UUID, name string, weightMode models.WeightMode, weightValue decimal.Decimal) *models.Strategy {
	strategy := &models.Strategy{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		WeightMode:  weightMode,
		WeightValue: weightValue,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO strategies (id, user_id, name, weight_mode, weight_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.Exec(query, strategy.ID, strategy.UserID, strategy.Name,
		strategy.WeightMode, strategy.WeightValue, strategy.CreatedAt, strategy.UpdatedAt)
	require.NoError(t, err)

	return strategy
}

// CleanupPortfolioTestData removes all test data related to portfolios
func CleanupPortfolioTestData(t *testing.T, db *sqlx.DB) {
	// Delete in reverse order of dependencies
	_, err := db.Exec("DELETE FROM nav_history")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM positions")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM portfolios")
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

// AssertPortfolioEqual asserts that two portfolios are equal
func AssertPortfolioEqual(t *testing.T, expected, actual *models.Portfolio) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.UserID, actual.UserID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.True(t, expected.TotalInvestment.Equal(actual.TotalInvestment))
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Second)
}

// AssertPositionEqual asserts that two positions are equal
func AssertPositionEqual(t *testing.T, expected, actual *models.Position) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)

	assert.Equal(t, expected.PortfolioID, actual.PortfolioID)
	assert.Equal(t, expected.StockID, actual.StockID)
	assert.Equal(t, expected.Quantity, actual.Quantity)
	assert.True(t, expected.EntryPrice.Equal(actual.EntryPrice))
	assert.True(t, expected.AllocationValue.Equal(actual.AllocationValue))
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Second)
}

// AssertNAVHistoryEqual asserts that two NAV history entries are equal
func AssertNAVHistoryEqual(t *testing.T, expected, actual *models.NAVHistory) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)

	assert.Equal(t, expected.PortfolioID, actual.PortfolioID)
	assert.WithinDuration(t, expected.Timestamp, actual.Timestamp, time.Second)
	assert.True(t, expected.NAV.Equal(actual.NAV))
	assert.True(t, expected.PnL.Equal(actual.PnL))
	
	if expected.Drawdown != nil && actual.Drawdown != nil {
		assert.True(t, expected.Drawdown.Equal(*actual.Drawdown))
	} else {
		assert.Equal(t, expected.Drawdown, actual.Drawdown)
	}
	
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Second)
}

// CreateTestNAVHistorySequence creates a sequence of NAV history entries for testing performance metrics
func CreateTestNAVHistorySequence(t *testing.T, db *sqlx.DB, portfolioID uuid.UUID, initialInvestment decimal.Decimal, days int) []*models.NAVHistory {
	var history []*models.NAVHistory
	
	baseTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	
	for i := 0; i <= days; i++ {
		timestamp := baseTime.Add(time.Duration(i) * 24 * time.Hour)
		
		// Simulate some price movement
		multiplier := 1.0 + (float64(i)*0.01) - (float64(i%7)*0.005) // Some volatility
		nav := initialInvestment.Mul(decimal.NewFromFloat(multiplier))
		pnl := nav.Sub(initialInvestment)
		
		var drawdown *decimal.Decimal
		if i > 0 && nav.LessThan(initialInvestment.Mul(decimal.NewFromFloat(1.1))) {
			dd := decimal.NewFromFloat(-5.0) // -5% drawdown example
			drawdown = &dd
		}
		
		navEntry := CreateTestNAVHistory(t, db, portfolioID, timestamp, nav, pnl, drawdown)
		history = append(history, navEntry)
	}
	
	return history
}

// AssertPerformanceMetricsValid validates that performance metrics are reasonable
func AssertPerformanceMetricsValid(t *testing.T, metrics *models.PerformanceMetrics, initialInvestment decimal.Decimal) {
	require.NotNil(t, metrics)
	
	// Basic validations
	assert.GreaterOrEqual(t, metrics.DaysActive, 0)
	assert.True(t, metrics.HighWaterMark.GreaterThanOrEqual(initialInvestment))
	
	// If we have drawdown, it should be negative or zero
	if metrics.MaxDrawdown != nil {
		assert.LessOrEqual(t, metrics.MaxDrawdown.InexactFloat64(), 0.0)
	}
	
	if metrics.CurrentDrawdown != nil {
		assert.LessOrEqual(t, metrics.CurrentDrawdown.InexactFloat64(), 0.0)
	}
	
	// Total return percentage should be consistent with total return
	if initialInvestment.GreaterThan(decimal.Zero) {
		expectedReturnPct := metrics.TotalReturn.Div(initialInvestment).Mul(decimal.NewFromInt(100))
		assert.True(t, expectedReturnPct.Sub(metrics.TotalReturnPct).Abs().LessThan(decimal.NewFromFloat(0.01)))
	}
}