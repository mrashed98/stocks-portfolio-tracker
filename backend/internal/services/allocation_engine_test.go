package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"portfolio-app/internal/models"
)

// Mock repositories for allocation engine
type MockAllocationStrategyRepository struct {
	mock.Mock
}

func (m *MockAllocationStrategyRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Strategy, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockAllocationStrategyRepository) GetStrategyStocks(ctx context.Context, strategyID uuid.UUID) ([]*models.StrategyStock, error) {
	args := m.Called(ctx, strategyID)
	return args.Get(0).([]*models.StrategyStock), args.Error(1)
}

type MockAllocationStockRepository struct {
	mock.Mock
}

func (m *MockAllocationStockRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Stock, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]*models.Stock), args.Error(1)
}

type MockAllocationSignalRepository struct {
	mock.Mock
}

func (m *MockAllocationSignalRepository) GetLatestSignals(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*models.Signal, error) {
	args := m.Called(ctx, stockIDs)
	return args.Get(0).(map[uuid.UUID]*models.Signal), args.Error(1)
}

func TestAllocationEngine_CalculateAllocations(t *testing.T) {
	// Setup
	mockStrategyRepo := new(MockAllocationStrategyRepository)
	mockStockRepo := new(MockAllocationStockRepository)
	mockSignalRepo := new(MockAllocationSignalRepository)
	mockMarketData := NewMockMarketDataService()
	
	engine := NewAllocationEngine(mockStrategyRepo, mockStockRepo, mockSignalRepo, mockMarketData)
	
	// Test data
	strategyID1 := uuid.New()
	strategyID2 := uuid.New()
	stockID1 := uuid.New()
	stockID2 := uuid.New()
	stockID3 := uuid.New()
	
	strategies := []*models.Strategy{
		{
			ID:          strategyID1,
			Name:        "Growth Strategy",
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(60),
		},
		{
			ID:          strategyID2,
			Name:        "Value Strategy",
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(40),
		},
	}
	
	stocks := []*models.Stock{
		{
			ID:     stockID1,
			Ticker: "AAPL",
			Name:   "Apple Inc.",
		},
		{
			ID:     stockID2,
			Ticker: "GOOGL",
			Name:   "Alphabet Inc.",
		},
		{
			ID:     stockID3,
			Ticker: "MSFT",
			Name:   "Microsoft Corp.",
		},
	}
	
	strategyStocks1 := []*models.StrategyStock{
		{StrategyID: strategyID1, StockID: stockID1, Eligible: true},
		{StrategyID: strategyID1, StockID: stockID2, Eligible: true},
	}
	
	strategyStocks2 := []*models.StrategyStock{
		{StrategyID: strategyID2, StockID: stockID2, Eligible: true},
		{StrategyID: strategyID2, StockID: stockID3, Eligible: true},
	}
	
	signals := map[uuid.UUID]*models.Signal{
		stockID1: {StockID: stockID1, Signal: models.SignalBuy, Date: time.Now()},
		stockID2: {StockID: stockID2, Signal: models.SignalBuy, Date: time.Now()},
		stockID3: {StockID: stockID3, Signal: models.SignalBuy, Date: time.Now()},
	}
	
	// Setup mocks
	mockStrategyRepo.On("GetByIDs", mock.Anything, []uuid.UUID{strategyID1, strategyID2}).Return(strategies, nil)
	mockStrategyRepo.On("GetStrategyStocks", mock.Anything, strategyID1).Return(strategyStocks1, nil)
	mockStrategyRepo.On("GetStrategyStocks", mock.Anything, strategyID2).Return(strategyStocks2, nil)
	mockStockRepo.On("GetByIDs", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(stocks, nil)
	mockSignalRepo.On("GetLatestSignals", mock.Anything, mock.AnythingOfType("[]uuid.UUID")).Return(signals, nil)
	
	// Test request
	req := &models.AllocationRequest{
		StrategyIDs:     []uuid.UUID{strategyID1, strategyID2},
		TotalInvestment: decimal.NewFromInt(10000),
		Constraints: models.AllocationConstraints{
			MaxAllocationPerStock: decimal.NewFromInt(50), // 50%
			MinAllocationAmount:   decimal.NewFromInt(100),
		},
	}
	
	// Execute
	result, err := engine.CalculateAllocations(context.Background(), req)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, decimal.NewFromInt(10000), result.TotalInvestment)
	assert.Len(t, result.Allocations, 3) // All 3 stocks should be allocated
	
	// Verify total allocation
	totalAllocated := decimal.Zero
	for _, allocation := range result.Allocations {
		totalAllocated = totalAllocated.Add(allocation.AllocationValue)
		assert.True(t, allocation.AllocationValue.GreaterThan(decimal.Zero))
		assert.True(t, allocation.Weight.GreaterThan(decimal.Zero))
	}
	
	// Should be close to total investment (allowing for rounding)
	assert.True(t, totalAllocated.Sub(result.TotalInvestment).Abs().LessThan(decimal.NewFromFloat(0.01)))
	
	mockStrategyRepo.AssertExpectations(t)
	mockStockRepo.AssertExpectations(t)
	mockSignalRepo.AssertExpectations(t)
}

func TestAllocationEngine_CalculateStrategyWeights_PercentageMode(t *testing.T) {
	engine := &AllocationEngine{}
	
	strategies := []*models.Strategy{
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(60),
		},
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(40),
		},
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result, err := engine.calculateStrategyWeights(strategies, totalInvestment)
	
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.True(t, decimal.NewFromInt(6000).Equal(result[strategies[0].ID]))
	assert.True(t, decimal.NewFromInt(4000).Equal(result[strategies[1].ID]))
}

func TestAllocationEngine_CalculateStrategyWeights_BudgetMode(t *testing.T) {
	engine := &AllocationEngine{}
	
	strategies := []*models.Strategy{
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(3000),
		},
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(2000),
		},
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result, err := engine.calculateStrategyWeights(strategies, totalInvestment)
	
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.True(t, decimal.NewFromInt(3000).Equal(result[strategies[0].ID]))
	assert.True(t, decimal.NewFromInt(2000).Equal(result[strategies[1].ID]))
}

func TestAllocationEngine_CalculateStrategyWeights_MixedMode(t *testing.T) {
	engine := &AllocationEngine{}
	
	strategies := []*models.Strategy{
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(3000),
		},
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(50), // 50% of remaining 7000 = 3500
		},
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result, err := engine.calculateStrategyWeights(strategies, totalInvestment)
	
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.True(t, decimal.NewFromInt(3000).Equal(result[strategies[0].ID]))
	assert.True(t, decimal.NewFromInt(3500).Equal(result[strategies[1].ID]))
}

func TestAllocationEngine_CalculateStrategyWeights_ExceedsTotal(t *testing.T) {
	engine := &AllocationEngine{}
	
	strategies := []*models.Strategy{
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(8000),
		},
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(5000),
		},
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	_, err := engine.calculateStrategyWeights(strategies, totalInvestment)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceed total investment")
}

func TestAllocationEngine_CalculateStrategyWeights_ExceedsPercentage(t *testing.T) {
	engine := &AllocationEngine{}
	
	strategies := []*models.Strategy{
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(60),
		},
		{
			ID:          uuid.New(),
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(50),
		},
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	_, err := engine.calculateStrategyWeights(strategies, totalInvestment)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceed 100%")
}

func TestAllocationEngine_ApplyConstraints(t *testing.T) {
	engine := &AllocationEngine{}
	
	allocations := []models.StockAllocation{
		{
			StockID:         uuid.New(),
			Ticker:          "AAPL",
			AllocationValue: decimal.NewFromInt(6000), // 60% of 10000
		},
		{
			StockID:         uuid.New(),
			Ticker:          "GOOGL",
			AllocationValue: decimal.NewFromInt(50), // Below minimum
		},
		{
			StockID:         uuid.New(),
			Ticker:          "MSFT",
			AllocationValue: decimal.NewFromInt(2000), // 20% of 10000
		},
	}
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(50), // 50%
		MinAllocationAmount:   decimal.NewFromInt(100),
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result, err := engine.applyConstraints(allocations, constraints, totalInvestment)
	
	assert.NoError(t, err)
	assert.Len(t, result, 2) // GOOGL should be filtered out due to minimum constraint
	
	// AAPL should be capped at 50% = 5000
	found := false
	for _, allocation := range result {
		if allocation.Ticker == "AAPL" {
			assert.True(t, decimal.NewFromInt(5000).Equal(allocation.AllocationValue))
			found = true
		}
	}
	assert.True(t, found, "AAPL allocation should be present and capped")
}

func TestAllocationEngine_NormalizeAllocations(t *testing.T) {
	engine := &AllocationEngine{}
	
	allocations := []models.StockAllocation{
		{
			StockID:         uuid.New(),
			Ticker:          "AAPL",
			AllocationValue: decimal.NewFromInt(3000),
			StrategyContrib: map[string]decimal.Decimal{
				"strategy1": decimal.NewFromInt(3000),
			},
		},
		{
			StockID:         uuid.New(),
			Ticker:          "GOOGL",
			AllocationValue: decimal.NewFromInt(2000),
			StrategyContrib: map[string]decimal.Decimal{
				"strategy1": decimal.NewFromInt(2000),
			},
		},
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := engine.normalizeAllocations(allocations, totalInvestment)
	
	assert.Len(t, result, 2)
	
	// Total should be normalized to 10000 from 5000 (2x factor)
	totalAllocated := decimal.Zero
	for _, allocation := range result {
		totalAllocated = totalAllocated.Add(allocation.AllocationValue)
		assert.True(t, allocation.Weight.GreaterThan(decimal.Zero))
	}
	
	assert.True(t, totalInvestment.Equal(totalAllocated))
	
	// Check individual allocations
	for _, allocation := range result {
		if allocation.Ticker == "AAPL" {
			assert.True(t, decimal.NewFromInt(6000).Equal(allocation.AllocationValue))
			assert.True(t, decimal.NewFromInt(60).Equal(allocation.Weight))
		} else if allocation.Ticker == "GOOGL" {
			assert.True(t, decimal.NewFromInt(4000).Equal(allocation.AllocationValue))
			assert.True(t, decimal.NewFromInt(40).Equal(allocation.Weight))
		}
	}
}

func TestAllocationEngine_ValidateConstraints(t *testing.T) {
	engine := &AllocationEngine{}
	
	allocations := []models.StockAllocation{
		{
			Ticker:          "AAPL",
			AllocationValue: decimal.NewFromInt(5000),
			Weight:          decimal.NewFromInt(50),
		},
		{
			Ticker:          "GOOGL",
			AllocationValue: decimal.NewFromInt(200),
			Weight:          decimal.NewFromInt(20),
		},
	}
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(60), // 60%
		MinAllocationAmount:   decimal.NewFromInt(100),
	}
	
	err := engine.ValidateConstraints(allocations, constraints)
	assert.NoError(t, err)
	
	// Test violation of max constraint
	allocations[0].Weight = decimal.NewFromInt(70)
	err = engine.ValidateConstraints(allocations, constraints)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum")
	
	// Test violation of min constraint
	allocations[0].Weight = decimal.NewFromInt(50)
	allocations[1].AllocationValue = decimal.NewFromInt(50)
	err = engine.ValidateConstraints(allocations, constraints)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "below minimum")
}