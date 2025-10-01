package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"portfolio-app/internal/models"
	"portfolio-app/internal/testutils"
)

func TestPortfolioRepository_Create(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test portfolio
	portfolio := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Test creation
	err := repo.Create(ctx, portfolio)
	require.NoError(t, err)

	// Verify portfolio was created
	retrieved, err := repo.GetByID(ctx, portfolio.ID)
	require.NoError(t, err)
	assert.Equal(t, portfolio.ID, retrieved.ID)
	assert.Equal(t, portfolio.UserID, retrieved.UserID)
	assert.Equal(t, portfolio.Name, retrieved.Name)
	assert.True(t, portfolio.TotalInvestment.Equal(retrieved.TotalInvestment))
}

func TestPortfolioRepository_GetByID(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Test getting non-existent portfolio
	_, err := repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "portfolio not found")
}

func TestPortfolioRepository_GetByUserID(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	userID := uuid.New()

	// Create multiple portfolios for the user
	portfolio1 := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            "Portfolio 1",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	portfolio2 := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            "Portfolio 2",
		TotalInvestment: decimal.NewFromFloat(20000.00),
		CreatedAt:       time.Now().Add(time.Hour),
		UpdatedAt:       time.Now().Add(time.Hour),
	}

	// Create portfolios
	require.NoError(t, repo.Create(ctx, portfolio1))
	require.NoError(t, repo.Create(ctx, portfolio2))

	// Get portfolios by user ID
	portfolios, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, portfolios, 2)

	// Verify portfolios are ordered by created_at DESC
	assert.Equal(t, portfolio2.ID, portfolios[0].ID)
	assert.Equal(t, portfolio1.ID, portfolios[1].ID)
}

func TestPortfolioRepository_Update(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test portfolio
	portfolio := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Name:            "Original Name",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, repo.Create(ctx, portfolio))

	// Update portfolio
	portfolio.Name = "Updated Name"
	portfolio.TotalInvestment = decimal.NewFromFloat(15000.00)
	portfolio.UpdatedAt = time.Now().Add(time.Hour)

	err := repo.Update(ctx, portfolio)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, portfolio.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.True(t, decimal.NewFromFloat(15000.00).Equal(retrieved.TotalInvestment))
}

func TestPortfolioRepository_Delete(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test portfolio
	portfolio := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, repo.Create(ctx, portfolio))

	// Delete portfolio
	err := repo.Delete(ctx, portfolio.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, portfolio.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "portfolio not found")
}

func TestPortfolioRepository_CreatePosition(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test portfolio and stock
	portfolioID := uuid.New()
	stockID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, repo.Create(ctx, portfolio))

	// Create test stock
	testutils.CreateTestStock(t, db, stockID, "AAPL", "Apple Inc.", "Technology")

	// Create position
	position := &models.Position{
		PortfolioID:     portfolioID,
		StockID:         stockID,
		Quantity:        100,
		EntryPrice:      decimal.NewFromFloat(150.00),
		AllocationValue: decimal.NewFromFloat(15000.00),
		StrategyContrib: []byte(`{"strategy1": 15000.00}`),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := repo.CreatePosition(ctx, position)
	require.NoError(t, err)

	// Verify position was created
	positions, err := repo.GetPositions(ctx, portfolioID)
	require.NoError(t, err)
	assert.Len(t, positions, 1)
	assert.Equal(t, stockID, positions[0].StockID)
	assert.Equal(t, 100, positions[0].Quantity)
	assert.True(t, decimal.NewFromFloat(150.00).Equal(positions[0].EntryPrice))
}

func TestPortfolioRepository_GetPositions(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test portfolio
	portfolioID := uuid.New()
	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, repo.Create(ctx, portfolio))

	// Test getting positions for empty portfolio
	positions, err := repo.GetPositions(ctx, portfolioID)
	require.NoError(t, err)
	assert.Len(t, positions, 0)
}

func TestPortfolioRepository_CreateNAVHistory(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test portfolio
	portfolioID := uuid.New()
	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, repo.Create(ctx, portfolio))

	// Create NAV history entry
	drawdown := decimal.NewFromFloat(-5.0)
	navHistory := &models.NAVHistory{
		PortfolioID: portfolioID,
		Timestamp:   time.Now(),
		NAV:         decimal.NewFromFloat(9500.00),
		PnL:         decimal.NewFromFloat(-500.00),
		Drawdown:    &drawdown,
		CreatedAt:   time.Now(),
	}

	err := repo.CreateNAVHistory(ctx, navHistory)
	require.NoError(t, err)

	// Verify NAV history was created
	latest, err := repo.GetLatestNAV(ctx, portfolioID)
	require.NoError(t, err)
	assert.Equal(t, portfolioID, latest.PortfolioID)
	assert.True(t, decimal.NewFromFloat(9500.00).Equal(latest.NAV))
	assert.True(t, decimal.NewFromFloat(-500.00).Equal(latest.PnL))
	assert.NotNil(t, latest.Drawdown)
	assert.True(t, decimal.NewFromFloat(-5.0).Equal(*latest.Drawdown))
}

func TestPortfolioRepository_GetNAVHistory(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test portfolio
	portfolioID := uuid.New()
	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, repo.Create(ctx, portfolio))

	// Create multiple NAV history entries
	now := time.Now()
	entries := []*models.NAVHistory{
		{
			PortfolioID: portfolioID,
			Timestamp:   now.Add(-2 * time.Hour),
			NAV:         decimal.NewFromFloat(10000.00),
			PnL:         decimal.Zero,
			CreatedAt:   now.Add(-2 * time.Hour),
		},
		{
			PortfolioID: portfolioID,
			Timestamp:   now.Add(-1 * time.Hour),
			NAV:         decimal.NewFromFloat(10500.00),
			PnL:         decimal.NewFromFloat(500.00),
			CreatedAt:   now.Add(-1 * time.Hour),
		},
		{
			PortfolioID: portfolioID,
			Timestamp:   now,
			NAV:         decimal.NewFromFloat(9800.00),
			PnL:         decimal.NewFromFloat(-200.00),
			CreatedAt:   now,
		},
	}

	for _, entry := range entries {
		require.NoError(t, repo.CreateNAVHistory(ctx, entry))
	}

	// Get NAV history within date range
	from := now.Add(-3 * time.Hour)
	to := now.Add(time.Hour)

	history, err := repo.GetNAVHistory(ctx, portfolioID, from, to)
	require.NoError(t, err)
	assert.Len(t, history, 3)

	// Verify entries are ordered by timestamp ASC
	assert.True(t, history[0].Timestamp.Before(history[1].Timestamp))
	assert.True(t, history[1].Timestamp.Before(history[2].Timestamp))
}

func TestPortfolioRepository_GetLatestNAV(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test portfolio
	portfolioID := uuid.New()
	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, repo.Create(ctx, portfolio))

	// Test getting latest NAV when none exists
	latest, err := repo.GetLatestNAV(ctx, portfolioID)
	require.NoError(t, err)
	assert.Nil(t, latest)
}

func TestPortfolioRepository_GetAllPortfolioIDs(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Test getting portfolio IDs when none exist
	portfolioIDs, err := repo.GetAllPortfolioIDs(ctx)
	require.NoError(t, err)
	assert.Len(t, portfolioIDs, 0)

	// Create multiple portfolios
	userID := uuid.New()
	portfolio1 := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            "Portfolio 1",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now().Add(-2 * time.Hour),
		UpdatedAt:       time.Now().Add(-2 * time.Hour),
	}

	portfolio2 := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            "Portfolio 2",
		TotalInvestment: decimal.NewFromFloat(20000.00),
		CreatedAt:       time.Now().Add(-1 * time.Hour),
		UpdatedAt:       time.Now().Add(-1 * time.Hour),
	}

	portfolio3 := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          uuid.New(), // Different user
		Name:            "Portfolio 3",
		TotalInvestment: decimal.NewFromFloat(15000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create portfolios
	require.NoError(t, repo.Create(ctx, portfolio1))
	require.NoError(t, repo.Create(ctx, portfolio2))
	require.NoError(t, repo.Create(ctx, portfolio3))

	// Get all portfolio IDs
	portfolioIDs, err = repo.GetAllPortfolioIDs(ctx)
	require.NoError(t, err)
	assert.Len(t, portfolioIDs, 3)

	// Verify all portfolio IDs are returned (order should be by created_at)
	expectedIDs := []uuid.UUID{portfolio1.ID, portfolio2.ID, portfolio3.ID}
	assert.ElementsMatch(t, expectedIDs, portfolioIDs)

	// Verify order is by created_at
	assert.Equal(t, portfolio1.ID, portfolioIDs[0])
	assert.Equal(t, portfolio2.ID, portfolioIDs[1])
	assert.Equal(t, portfolio3.ID, portfolioIDs[2])
}

func TestPortfolioRepository_CreatePortfolioWithPositions(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, db)

	repo := NewPortfolioRepository(db.DB)
	ctx := context.Background()

	// Create test stocks
	stockID1 := uuid.New()
	stockID2 := uuid.New()
	testutils.CreateTestStock(t, db, stockID1, "AAPL", "Apple Inc.", "Technology")
	testutils.CreateTestStock(t, db, stockID2, "GOOGL", "Alphabet Inc.", "Technology")

	// Create portfolio with positions
	portfolioID := uuid.New()
	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(20000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	positions := []*models.Position{
		{
			PortfolioID:     portfolioID,
			StockID:         stockID1,
			Quantity:        100,
			EntryPrice:      decimal.NewFromFloat(150.00),
			AllocationValue: decimal.NewFromFloat(15000.00),
			StrategyContrib: []byte(`{"strategy1": 15000.00}`),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			PortfolioID:     portfolioID,
			StockID:         stockID2,
			Quantity:        20,
			EntryPrice:      decimal.NewFromFloat(250.00),
			AllocationValue: decimal.NewFromFloat(5000.00),
			StrategyContrib: []byte(`{"strategy1": 5000.00}`),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Create portfolio with positions in transaction
	err := repo.CreatePortfolioWithPositions(ctx, portfolio, positions)
	require.NoError(t, err)

	// Verify portfolio was created
	retrievedPortfolio, err := repo.GetByID(ctx, portfolioID)
	require.NoError(t, err)
	assert.Equal(t, portfolio.Name, retrievedPortfolio.Name)
	assert.Len(t, retrievedPortfolio.Positions, 2)

	// Verify initial NAV history was created
	assert.Len(t, retrievedPortfolio.NAVHistory, 1)
	assert.True(t, portfolio.TotalInvestment.Equal(retrievedPortfolio.NAVHistory[0].NAV))
	assert.True(t, decimal.Zero.Equal(retrievedPortfolio.NAVHistory[0].PnL))
}