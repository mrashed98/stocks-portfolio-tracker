package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"portfolio-app/internal/models"
)

// Mock implementations
type MockPortfolioRepository struct {
	mock.Mock
}

func (m *MockPortfolioRepository) Create(ctx context.Context, portfolio *models.Portfolio) error {
	args := m.Called(ctx, portfolio)
	return args.Error(0)
}

func (m *MockPortfolioRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Portfolio, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Portfolio, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioRepository) GetAllPortfolioIDs(ctx context.Context) ([]uuid.UUID, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockPortfolioRepository) Update(ctx context.Context, portfolio *models.Portfolio) error {
	args := m.Called(ctx, portfolio)
	return args.Error(0)
}

func (m *MockPortfolioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPortfolioRepository) CreatePosition(ctx context.Context, position *models.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

func (m *MockPortfolioRepository) GetPositions(ctx context.Context, portfolioID uuid.UUID) ([]*models.Position, error) {
	args := m.Called(ctx, portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Position), args.Error(1)
}

func (m *MockPortfolioRepository) UpdatePosition(ctx context.Context, position *models.Position) error {
	args := m.Called(ctx, position)
	return args.Error(0)
}

func (m *MockPortfolioRepository) DeletePosition(ctx context.Context, portfolioID, stockID uuid.UUID) error {
	args := m.Called(ctx, portfolioID, stockID)
	return args.Error(0)
}

func (m *MockPortfolioRepository) CreateNAVHistory(ctx context.Context, navHistory *models.NAVHistory) error {
	args := m.Called(ctx, navHistory)
	return args.Error(0)
}

func (m *MockPortfolioRepository) GetNAVHistory(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]*models.NAVHistory, error) {
	args := m.Called(ctx, portfolioID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NAVHistory), args.Error(1)
}

func (m *MockPortfolioRepository) GetLatestNAV(ctx context.Context, portfolioID uuid.UUID) (*models.NAVHistory, error) {
	args := m.Called(ctx, portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NAVHistory), args.Error(1)
}

func (m *MockPortfolioRepository) CreatePortfolioWithPositions(ctx context.Context, portfolio *models.Portfolio, positions []*models.Position) error {
	args := m.Called(ctx, portfolio, positions)
	return args.Error(0)
}

type MockAllocationEngine struct {
	mock.Mock
}

func (m *MockAllocationEngine) CalculateAllocations(ctx context.Context, req *models.AllocationRequest) (*models.AllocationPreview, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockAllocationEngine) RebalanceAllocations(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.AllocationPreview, error) {
	args := m.Called(ctx, portfolioID, newTotalInvestment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockAllocationEngine) RecalculateWithExclusions(ctx context.Context, originalReq *models.AllocationRequest, excludedStocks []uuid.UUID) (*models.AllocationPreview, error) {
	args := m.Called(ctx, originalReq, excludedStocks)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockAllocationEngine) ValidateConstraints(allocations []models.StockAllocation, constraints models.AllocationConstraints) error {
	args := m.Called(allocations, constraints)
	return args.Error(0)
}

func (m *MockAllocationEngine) ValidateConstraintsDetailed(allocations []models.StockAllocation, constraints models.AllocationConstraints, totalInvestment decimal.Decimal) *ValidationResult {
	args := m.Called(allocations, constraints, totalInvestment)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*ValidationResult)
}

func (m *MockAllocationEngine) ValidateConstraintsConfig(constraints models.AllocationConstraints, totalInvestment decimal.Decimal) *ValidationResult {
	args := m.Called(constraints, totalInvestment)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*ValidationResult)
}

type MockTestStrategyRepository struct {
	mock.Mock
}

func (m *MockTestStrategyRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Strategy, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockTestStrategyRepository) GetStrategyStocks(ctx context.Context, strategyID uuid.UUID) ([]*models.StrategyStock, error) {
	args := m.Called(ctx, strategyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.StrategyStock), args.Error(1)
}

type MockTestMarketDataService struct {
	mock.Mock
}

func (m *MockTestMarketDataService) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Quote), args.Error(1)
}

func (m *MockTestMarketDataService) GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*Quote, error) {
	args := m.Called(ctx, symbols)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*Quote), args.Error(1)
}

func (m *MockTestMarketDataService) GetQuotesByStockIDs(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*Quote, error) {
	args := m.Called(ctx, stockIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*Quote), args.Error(1)
}

func (m *MockTestMarketDataService) GetOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error) {
	args := m.Called(ctx, symbol, from, to, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*OHLCV), args.Error(1)
}

func TestPortfolioService_CreatePortfolio(t *testing.T) {
	// Setup mocks
	mockRepo := &MockPortfolioRepository{}
	mockAllocationEngine := &MockAllocationEngine{}
	mockStrategyRepo := &MockTestStrategyRepository{}
	mockMarketDataService := &MockTestMarketDataService{}

	service := NewPortfolioService(mockAllocationEngine, mockStrategyRepo, mockRepo, mockMarketDataService)

	ctx := context.Background()
	userID := uuid.New()
	stockID := uuid.New()

	// Test data
	req := &models.CreatePortfolioRequest{
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		Positions: []models.CreatePositionRequest{
			{
				StockID:         stockID,
				Quantity:        100,
				EntryPrice:      decimal.NewFromFloat(100.00),
				AllocationValue: decimal.NewFromFloat(10000.00),
				StrategyContrib: map[string]decimal.Decimal{
					"strategy1": decimal.NewFromFloat(10000.00),
				},
			},
		},
	}

	expectedPortfolio := &models.Portfolio{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Setup expectations
	mockRepo.On("CreatePortfolioWithPositions", ctx, mock.AnythingOfType("*models.Portfolio"), mock.AnythingOfType("[]*models.Position")).Return(nil)
	mockRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(expectedPortfolio, nil)

	// Execute
	result, err := service.CreatePortfolio(ctx, req, userID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPortfolio.Name, result.Name)
	assert.True(t, expectedPortfolio.TotalInvestment.Equal(result.TotalInvestment))

	mockRepo.AssertExpectations(t)
}

func TestPortfolioService_CreatePortfolio_ValidationErrors(t *testing.T) {
	service := NewPortfolioService(nil, nil, nil, nil)
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		req     *models.CreatePortfolioRequest
		wantErr string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: "create portfolio request cannot be nil",
		},
		{
			name: "empty name",
			req: &models.CreatePortfolioRequest{
				Name:            "",
				TotalInvestment: decimal.NewFromFloat(10000.00),
				Positions:       []models.CreatePositionRequest{},
			},
			wantErr: "portfolio name is required",
		},
		{
			name: "zero investment",
			req: &models.CreatePortfolioRequest{
				Name:            "Test",
				TotalInvestment: decimal.Zero,
				Positions:       []models.CreatePositionRequest{},
			},
			wantErr: "total investment must be greater than zero",
		},
		{
			name: "no positions",
			req: &models.CreatePortfolioRequest{
				Name:            "Test",
				TotalInvestment: decimal.NewFromFloat(10000.00),
				Positions:       []models.CreatePositionRequest{},
			},
			wantErr: "at least one position is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreatePortfolio(ctx, tt.req, userID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestPortfolioService_GetPortfolio(t *testing.T) {
	// Setup mocks
	mockRepo := &MockPortfolioRepository{}
	mockAllocationEngine := &MockAllocationEngine{}
	mockStrategyRepo := &MockTestStrategyRepository{}
	mockMarketDataService := &MockTestMarketDataService{}

	service := NewPortfolioService(mockAllocationEngine, mockStrategyRepo, mockRepo, mockMarketDataService)

	ctx := context.Background()
	portfolioID := uuid.New()
	stockID := uuid.New()

	expectedPortfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Positions: []models.Position{
			{
				PortfolioID:     portfolioID,
				StockID:         stockID,
				Quantity:        100,
				EntryPrice:      decimal.NewFromFloat(100.00),
				AllocationValue: decimal.NewFromFloat(10000.00),
				Stock: &models.Stock{
					ID:     stockID,
					Ticker: "AAPL",
					Name:   "Apple Inc.",
				},
			},
		},
	}

	quotes := map[string]*Quote{
		"AAPL": {
			Symbol: "AAPL",
			Price:  decimal.NewFromFloat(150.00),
		},
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, portfolioID).Return(expectedPortfolio, nil)
	mockMarketDataService.On("GetMultipleQuotes", ctx, []string{"AAPL"}).Return(quotes, nil)

	// Execute
	result, err := service.GetPortfolio(ctx, portfolioID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPortfolio.ID, result.ID)
	assert.Len(t, result.Positions, 1)

	// Verify position was enriched with market data
	position := result.Positions[0]
	assert.NotNil(t, position.CurrentPrice)
	assert.True(t, decimal.NewFromFloat(150.00).Equal(*position.CurrentPrice))

	mockRepo.AssertExpectations(t)
	mockMarketDataService.AssertExpectations(t)
}

func TestPortfolioService_UpdatePortfolioNAV(t *testing.T) {
	// Setup mocks
	mockRepo := &MockPortfolioRepository{}
	mockAllocationEngine := &MockAllocationEngine{}
	mockStrategyRepo := &MockTestStrategyRepository{}
	mockMarketDataService := &MockTestMarketDataService{}

	service := NewPortfolioService(mockAllocationEngine, mockStrategyRepo, mockRepo, mockMarketDataService)

	ctx := context.Background()
	portfolioID := uuid.New()
	stockID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Positions: []models.Position{
			{
				PortfolioID:     portfolioID,
				StockID:         stockID,
				Quantity:        100,
				EntryPrice:      decimal.NewFromFloat(100.00),
				AllocationValue: decimal.NewFromFloat(10000.00),
				Stock: &models.Stock{
					ID:     stockID,
					Ticker: "AAPL",
					Name:   "Apple Inc.",
				},
			},
		},
	}

	quotes := map[string]*Quote{
		"AAPL": {
			Symbol: "AAPL",
			Price:  decimal.NewFromFloat(150.00),
		},
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, portfolioID).Return(portfolio, nil)
	mockMarketDataService.On("GetMultipleQuotes", ctx, []string{"AAPL"}).Return(quotes, nil)
	mockRepo.On("GetNAVHistory", ctx, portfolioID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return([]*models.NAVHistory{}, nil)
	mockRepo.On("CreateNAVHistory", ctx, mock.AnythingOfType("*models.NAVHistory")).Return(nil)

	// Execute
	result, err := service.UpdatePortfolioNAV(ctx, portfolioID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, portfolioID, result.PortfolioID)
	assert.True(t, decimal.NewFromFloat(15000.00).Equal(result.NAV)) // 100 shares * $150
	assert.True(t, decimal.NewFromFloat(5000.00).Equal(result.PnL))  // $15000 - $10000

	mockRepo.AssertExpectations(t)
	mockMarketDataService.AssertExpectations(t)
}

func TestPortfolioService_GenerateRebalancePreview(t *testing.T) {
	// Setup mocks
	mockRepo := &MockPortfolioRepository{}
	mockAllocationEngine := &MockAllocationEngine{}
	mockStrategyRepo := &MockTestStrategyRepository{}
	mockMarketDataService := &MockTestMarketDataService{}

	service := NewPortfolioService(mockAllocationEngine, mockStrategyRepo, mockRepo, mockMarketDataService)

	ctx := context.Background()
	portfolioID := uuid.New()
	strategyID := uuid.New()
	stockID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Positions: []models.Position{
			{
				PortfolioID:     portfolioID,
				StockID:         stockID,
				Quantity:        100,
				EntryPrice:      decimal.NewFromFloat(100.00),
				AllocationValue: decimal.NewFromFloat(10000.00),
				StrategyContribMap: map[string]decimal.Decimal{
					strategyID.String(): decimal.NewFromFloat(10000.00),
				},
			},
		},
	}

	expectedPreview := &models.AllocationPreview{
		TotalInvestment: decimal.NewFromFloat(20000.00),
		Allocations: []models.StockAllocation{
			{
				StockID:         stockID,
				Ticker:          "AAPL",
				AllocationValue: decimal.NewFromFloat(20000.00),
				Price:           decimal.NewFromFloat(150.00),
				Quantity:        133,
				ActualValue:     decimal.NewFromFloat(19950.00),
			},
		},
		UnallocatedCash: decimal.NewFromFloat(50.00),
		TotalAllocated:  decimal.NewFromFloat(19950.00),
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, portfolioID).Return(portfolio, nil)
	mockAllocationEngine.On("CalculateAllocations", ctx, mock.AnythingOfType("*models.AllocationRequest")).Return(expectedPreview, nil)

	// Execute
	result, err := service.GenerateRebalancePreview(ctx, portfolioID, decimal.NewFromFloat(20000.00))

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, decimal.NewFromFloat(20000.00).Equal(result.TotalInvestment))
	assert.Len(t, result.Allocations, 1)

	mockRepo.AssertExpectations(t)
	mockAllocationEngine.AssertExpectations(t)
}

func TestPortfolioService_GetPortfolioPerformanceMetrics(t *testing.T) {
	// Setup mocks
	mockRepo := &MockPortfolioRepository{}
	mockAllocationEngine := &MockAllocationEngine{}
	mockStrategyRepo := &MockTestStrategyRepository{}
	mockMarketDataService := &MockTestMarketDataService{}

	service := NewPortfolioService(mockAllocationEngine, mockStrategyRepo, mockRepo, mockMarketDataService)

	ctx := context.Background()
	portfolioID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	navHistory := []*models.NAVHistory{
		{
			PortfolioID: portfolioID,
			Timestamp:   time.Now().Add(-30 * 24 * time.Hour),
			NAV:         decimal.NewFromFloat(10000.00),
			PnL:         decimal.Zero,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			PortfolioID: portfolioID,
			Timestamp:   time.Now().Add(-15 * 24 * time.Hour),
			NAV:         decimal.NewFromFloat(12000.00),
			PnL:         decimal.NewFromFloat(2000.00),
			Drawdown:    nil,
			CreatedAt:   time.Now().Add(-15 * 24 * time.Hour),
		},
		{
			PortfolioID: portfolioID,
			Timestamp:   time.Now(),
			NAV:         decimal.NewFromFloat(11500.00),
			PnL:         decimal.NewFromFloat(1500.00),
			Drawdown:    &[]decimal.Decimal{decimal.NewFromFloat(-4.17)}[0], // -4.17% drawdown
			CreatedAt:   time.Now(),
		},
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, portfolioID).Return(portfolio, nil)
	mockRepo.On("GetNAVHistory", ctx, portfolioID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(navHistory, nil)

	// Execute
	result, err := service.GetPortfolioPerformanceMetrics(ctx, portfolioID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, decimal.NewFromFloat(1500.00).Equal(result.TotalReturn))
	assert.True(t, decimal.NewFromFloat(15.00).Equal(result.TotalReturnPct))
	assert.Equal(t, 30, result.DaysActive)
	assert.True(t, decimal.NewFromFloat(12000.00).Equal(result.HighWaterMark))
	assert.NotNil(t, result.MaxDrawdown)
	assert.True(t, decimal.NewFromFloat(-4.17).Equal(*result.MaxDrawdown))

	mockRepo.AssertExpectations(t)
}

func TestPortfolioService_UpdatePortfolioNAV_EmptyPortfolio(t *testing.T) {
	// Setup mocks
	mockRepo := &MockPortfolioRepository{}
	mockAllocationEngine := &MockAllocationEngine{}
	mockStrategyRepo := &MockTestStrategyRepository{}
	mockMarketDataService := &MockTestMarketDataService{}

	service := NewPortfolioService(mockAllocationEngine, mockStrategyRepo, mockRepo, mockMarketDataService)

	ctx := context.Background()
	portfolioID := uuid.New()

	// Portfolio with no positions
	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Empty Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Positions:       []models.Position{}, // Empty positions
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, portfolioID).Return(portfolio, nil)
	mockRepo.On("GetNAVHistory", ctx, portfolioID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return([]*models.NAVHistory{}, nil)
	mockRepo.On("CreateNAVHistory", ctx, mock.AnythingOfType("*models.NAVHistory")).Return(nil)

	// Execute
	result, err := service.UpdatePortfolioNAV(ctx, portfolioID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, portfolioID, result.PortfolioID)
	assert.True(t, portfolio.TotalInvestment.Equal(result.NAV)) // NAV equals cash for empty portfolio
	assert.True(t, decimal.Zero.Equal(result.PnL))

	mockRepo.AssertExpectations(t)
}

func TestPortfolioService_UpdatePortfolioNAV_WithDrawdownCalculation(t *testing.T) {
	// Setup mocks
	mockRepo := &MockPortfolioRepository{}
	mockAllocationEngine := &MockAllocationEngine{}
	mockStrategyRepo := &MockTestStrategyRepository{}
	mockMarketDataService := &MockTestMarketDataService{}

	service := NewPortfolioService(mockAllocationEngine, mockStrategyRepo, mockRepo, mockMarketDataService)

	ctx := context.Background()
	portfolioID := uuid.New()
	stockID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Positions: []models.Position{
			{
				PortfolioID:     portfolioID,
				StockID:         stockID,
				Quantity:        100,
				EntryPrice:      decimal.NewFromFloat(100.00),
				AllocationValue: decimal.NewFromFloat(10000.00),
				Stock: &models.Stock{
					ID:     stockID,
					Ticker: "AAPL",
					Name:   "Apple Inc.",
				},
			},
		},
	}

	quotes := map[string]*Quote{
		"AAPL": {
			Symbol: "AAPL",
			Price:  decimal.NewFromFloat(90.00), // Price dropped from $100 to $90
		},
	}

	// Historical NAV showing a previous high
	previousNAVHistory := []*models.NAVHistory{
		{
			PortfolioID: portfolioID,
			Timestamp:   time.Now().Add(-24 * time.Hour),
			NAV:         decimal.NewFromFloat(12000.00), // Previous high
			PnL:         decimal.NewFromFloat(2000.00),
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
	}

	// Setup expectations
	mockRepo.On("GetByID", ctx, portfolioID).Return(portfolio, nil)
	mockMarketDataService.On("GetMultipleQuotes", ctx, []string{"AAPL"}).Return(quotes, nil)
	mockRepo.On("GetNAVHistory", ctx, portfolioID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(previousNAVHistory, nil)
	mockRepo.On("CreateNAVHistory", ctx, mock.AnythingOfType("*models.NAVHistory")).Return(nil)

	// Execute
	result, err := service.UpdatePortfolioNAV(ctx, portfolioID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, portfolioID, result.PortfolioID)
	assert.True(t, decimal.NewFromFloat(9000.00).Equal(result.NAV)) // 100 shares * $90
	assert.True(t, decimal.NewFromFloat(-1000.00).Equal(result.PnL)) // $9000 - $10000
	
	// Should have drawdown calculated from high water mark of $12000
	assert.NotNil(t, result.Drawdown)
	expectedDrawdown := decimal.NewFromFloat(-25.00) // (9000 - 12000) / 12000 * 100
	assert.True(t, expectedDrawdown.Equal(*result.Drawdown))

	mockRepo.AssertExpectations(t)
	mockMarketDataService.AssertExpectations(t)
}

func TestPortfolioService_RebalancePortfolio_Success(t *testing.T) {
	// Setup mocks
	mockRepo := &MockPortfolioRepository{}
	mockAllocationEngine := &MockAllocationEngine{}
	mockStrategyRepo := &MockTestStrategyRepository{}
	mockMarketDataService := &MockTestMarketDataService{}

	service := NewPortfolioService(mockAllocationEngine, mockStrategyRepo, mockRepo, mockMarketDataService)

	ctx := context.Background()
	portfolioID := uuid.New()
	strategyID := uuid.New()
	stockID := uuid.New()

	portfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Positions: []models.Position{
			{
				PortfolioID:     portfolioID,
				StockID:         stockID,
				Quantity:        100,
				EntryPrice:      decimal.NewFromFloat(100.00),
				AllocationValue: decimal.NewFromFloat(10000.00),
				StrategyContribMap: map[string]decimal.Decimal{
					strategyID.String(): decimal.NewFromFloat(10000.00),
				},
			},
		},
	}

	rebalancePreview := &models.AllocationPreview{
		TotalInvestment: decimal.NewFromFloat(20000.00),
		Allocations: []models.StockAllocation{
			{
				StockID:         stockID,
				Ticker:          "AAPL",
				AllocationValue: decimal.NewFromFloat(20000.00),
				Price:           decimal.NewFromFloat(150.00),
				Quantity:        133,
				ActualValue:     decimal.NewFromFloat(19950.00),
			},
		},
		UnallocatedCash: decimal.NewFromFloat(50.00),
		TotalAllocated:  decimal.NewFromFloat(19950.00),
	}



	// Setup expectations
	mockRepo.On("GetByID", ctx, portfolioID).Return(portfolio, nil).Times(4) // Called multiple times: GenerateRebalancePreview, RebalancePortfolio, UpdatePortfolioNAV, GetPortfolio
	mockAllocationEngine.On("CalculateAllocations", ctx, mock.AnythingOfType("*models.AllocationRequest")).Return(rebalancePreview, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Portfolio")).Return(nil)
	mockRepo.On("UpdatePosition", ctx, mock.AnythingOfType("*models.Position")).Return(nil)
	mockMarketDataService.On("GetMultipleQuotes", ctx, mock.AnythingOfType("[]string")).Return(map[string]*Quote{}, nil)
	mockRepo.On("GetNAVHistory", ctx, portfolioID, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return([]*models.NAVHistory{}, nil)
	mockRepo.On("CreateNAVHistory", ctx, mock.AnythingOfType("*models.NAVHistory")).Return(nil)

	// Execute
	result, err := service.RebalancePortfolio(ctx, portfolioID, decimal.NewFromFloat(20000.00))

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, portfolioID, result.ID)

	mockRepo.AssertExpectations(t)
	mockAllocationEngine.AssertExpectations(t)
}

func TestPortfolioService_ValidateAllocationRequest(t *testing.T) {
	service := NewPortfolioService(nil, nil, nil, nil)

	tests := []struct {
		name    string
		req     *models.AllocationRequest
		wantErr string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: "allocation request cannot be nil",
		},
		{
			name: "no strategies",
			req: &models.AllocationRequest{
				StrategyIDs:     []uuid.UUID{},
				TotalInvestment: decimal.NewFromFloat(10000.00),
				Constraints: models.AllocationConstraints{
					MaxAllocationPerStock: decimal.NewFromFloat(20.0),
					MinAllocationAmount:   decimal.NewFromFloat(100.0),
				},
			},
			wantErr: "at least one strategy must be specified",
		},
		{
			name: "zero investment",
			req: &models.AllocationRequest{
				StrategyIDs:     []uuid.UUID{uuid.New()},
				TotalInvestment: decimal.Zero,
				Constraints: models.AllocationConstraints{
					MaxAllocationPerStock: decimal.NewFromFloat(20.0),
					MinAllocationAmount:   decimal.NewFromFloat(100.0),
				},
			},
			wantErr: "total investment must be greater than zero",
		},
		{
			name: "invalid max allocation",
			req: &models.AllocationRequest{
				StrategyIDs:     []uuid.UUID{uuid.New()},
				TotalInvestment: decimal.NewFromFloat(10000.00),
				Constraints: models.AllocationConstraints{
					MaxAllocationPerStock: decimal.NewFromFloat(150.0), // > 100%
					MinAllocationAmount:   decimal.NewFromFloat(100.0),
				},
			},
			wantErr: "maximum allocation per stock cannot exceed 100%",
		},
		{
			name: "negative min allocation",
			req: &models.AllocationRequest{
				StrategyIDs:     []uuid.UUID{uuid.New()},
				TotalInvestment: decimal.NewFromFloat(10000.00),
				Constraints: models.AllocationConstraints{
					MaxAllocationPerStock: decimal.NewFromFloat(20.0),
					MinAllocationAmount:   decimal.NewFromFloat(-100.0),
				},
			},
			wantErr: "minimum allocation amount cannot be negative",
		},
		{
			name: "valid request",
			req: &models.AllocationRequest{
				StrategyIDs:     []uuid.UUID{uuid.New()},
				TotalInvestment: decimal.NewFromFloat(10000.00),
				Constraints: models.AllocationConstraints{
					MaxAllocationPerStock: decimal.NewFromFloat(20.0),
					MinAllocationAmount:   decimal.NewFromFloat(100.0),
				},
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateAllocationRequest(tt.req)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}