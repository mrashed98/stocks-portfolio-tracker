package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"portfolio-app/internal/models"
)

// MockStockRepository is a mock implementation of StockRepository
type MockStockRepository struct {
	mock.Mock
}

func (m *MockStockRepository) Create(ctx context.Context, stock *models.Stock) (*models.Stock, error) {
	args := m.Called(ctx, stock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockStockRepository) Update(ctx context.Context, stock *models.Stock) (*models.Stock, error) {
	args := m.Called(ctx, stock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockStockRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Stock, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockStockRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Stock, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Stock), args.Error(1)
}

func (m *MockStockRepository) GetByTicker(ctx context.Context, ticker string) (*models.Stock, error) {
	args := m.Called(ctx, ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockStockRepository) GetAll(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error) {
	args := m.Called(ctx, search, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Stock), args.Error(1)
}

func (m *MockStockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStockRepository) GetStocksWithCurrentSignals(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error) {
	args := m.Called(ctx, search, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Stock), args.Error(1)
}

// MockSignalRepository is a mock implementation of SignalRepository
type MockSignalRepository struct {
	mock.Mock
}

func (m *MockSignalRepository) Create(ctx context.Context, signal *models.Signal) (*models.Signal, error) {
	args := m.Called(ctx, signal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Signal), args.Error(1)
}

func (m *MockSignalRepository) Update(ctx context.Context, stockID uuid.UUID, signal models.SignalType) (*models.Signal, error) {
	args := m.Called(ctx, stockID, signal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Signal), args.Error(1)
}

func (m *MockSignalRepository) GetCurrentSignal(ctx context.Context, stockID uuid.UUID) (*models.Signal, error) {
	args := m.Called(ctx, stockID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Signal), args.Error(1)
}

func (m *MockSignalRepository) GetSignalHistory(ctx context.Context, stockID uuid.UUID, from, to time.Time) ([]*models.Signal, error) {
	args := m.Called(ctx, stockID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Signal), args.Error(1)
}

func (m *MockSignalRepository) Delete(ctx context.Context, stockID uuid.UUID, date time.Time) error {
	args := m.Called(ctx, stockID, date)
	return args.Error(0)
}

func (m *MockSignalRepository) GetLatestSignals(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*models.Signal, error) {
	args := m.Called(ctx, stockIDs)
	return args.Get(0).(map[uuid.UUID]*models.Signal), args.Error(1)
}

func TestStockService_CreateStock(t *testing.T) {

	t.Run("successful stock creation", func(t *testing.T) {
		// Create fresh mocks for this test
		mockStockRepo1 := new(MockStockRepository)
		mockSignalRepo1 := new(MockSignalRepository)
		mockStrategyRepo1 := new(MockStrategyRepository)
		service1 := NewStockService(mockStockRepo1, mockSignalRepo1, mockStrategyRepo1, &sql.DB{})

		req := &models.CreateStockRequest{
			Ticker: "AAPL",
			Name:   "Apple Inc.",
			Sector: stockStringPtr("Technology"),
		}

		expectedStock := &models.Stock{
			ID:     uuid.New(),
			Ticker: "AAPL",
			Name:   "Apple Inc.",
			Sector: stockStringPtr("Technology"),
		}

		mockStockRepo1.On("GetByTicker", mock.Anything, "AAPL").Return(nil, &models.NotFoundError{Resource: "stock"})
		mockStockRepo1.On("Create", mock.Anything, mock.AnythingOfType("*models.Stock")).Return(expectedStock, nil)

		result, err := service1.CreateStock(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expectedStock.Ticker, result.Ticker)
		assert.Equal(t, expectedStock.Name, result.Name)
		mockStockRepo1.AssertExpectations(t)
	})

	t.Run("invalid ticker symbol", func(t *testing.T) {
		// Create fresh mocks for this test
		mockStockRepo3 := new(MockStockRepository)
		mockSignalRepo3 := new(MockSignalRepository)
		mockStrategyRepo3 := new(MockStrategyRepository)
		service3 := NewStockService(mockStockRepo3, mockSignalRepo3, mockStrategyRepo3, &sql.DB{})

		req := &models.CreateStockRequest{
			Ticker: "INVALID_TICKER_TOO_LONG",
			Name:   "Invalid Company",
		}

		result, err := service3.CreateStock(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Ticker symbol must be 1-20 characters")
	})

	t.Run("duplicate ticker", func(t *testing.T) {
		// Create a fresh mock for this test
		mockStockRepo2 := new(MockStockRepository)
		mockSignalRepo2 := new(MockSignalRepository)
		mockStrategyRepo2 := new(MockStrategyRepository)
		service2 := NewStockService(mockStockRepo2, mockSignalRepo2, mockStrategyRepo2, &sql.DB{})

		req := &models.CreateStockRequest{
			Ticker: "AAPL",
			Name:   "Apple Inc.",
		}

		existingStock := &models.Stock{
			ID:     uuid.New(),
			Ticker: "AAPL",
			Name:   "Apple Inc.",
		}

		mockStockRepo2.On("GetByTicker", mock.Anything, "AAPL").Return(existingStock, nil)

		result, err := service2.CreateStock(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already exists")
		mockStockRepo2.AssertExpectations(t)
	})
}

func TestStockService_ValidateTickerSymbol(t *testing.T) {
	service := &stockService{}

	testCases := []struct {
		name        string
		ticker      string
		expectError bool
	}{
		{"valid ticker", "AAPL", false},
		{"empty ticker", "", true},
		{"too long ticker", "VERYLONGTICKERSYMBOLEXTRA", true},
		{"invalid characters", "AAPL@", true},
		{"valid short ticker", "A", false},
		{"valid with numbers", "BRK1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidateTickerSymbol(tc.ticker)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStockService_UpdateStockSignal(t *testing.T) {
	mockStockRepo := new(MockStockRepository)
	mockSignalRepo := new(MockSignalRepository)
	mockStrategyRepo := new(MockStrategyRepository)
	
	service := NewStockService(mockStockRepo, mockSignalRepo, mockStrategyRepo, &sql.DB{})

	t.Run("successful signal update", func(t *testing.T) {
		stockID := uuid.New()
		signalType := models.SignalBuy

		stock := &models.Stock{
			ID:     stockID,
			Ticker: "AAPL",
			Name:   "Apple Inc.",
		}

		expectedSignal := &models.Signal{
			StockID: stockID,
			Signal:  signalType,
			Date:    time.Now(),
		}

		mockStockRepo.On("GetByID", mock.Anything, stockID).Return(stock, nil)
		mockSignalRepo.On("Update", mock.Anything, stockID, signalType).Return(expectedSignal, nil)

		result, err := service.UpdateStockSignal(context.Background(), stockID, signalType)

		assert.NoError(t, err)
		assert.Equal(t, expectedSignal.Signal, result.Signal)
		assert.Equal(t, expectedSignal.StockID, result.StockID)
		mockStockRepo.AssertExpectations(t)
		mockSignalRepo.AssertExpectations(t)
	})

	t.Run("stock not found", func(t *testing.T) {
		stockID := uuid.New()
		signalType := models.SignalBuy

		mockStockRepo.On("GetByID", mock.Anything, stockID).Return(nil, &models.NotFoundError{Resource: "stock"})

		result, err := service.UpdateStockSignal(context.Background(), stockID, signalType)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "stock not found")
		mockStockRepo.AssertExpectations(t)
	})
}

func TestStockService_AddStockToStrategy(t *testing.T) {
	mockStockRepo := new(MockStockRepository)
	mockSignalRepo := new(MockSignalRepository)
	mockStrategyRepo := new(MockStrategyRepository)
	
	service := NewStockService(mockStockRepo, mockSignalRepo, mockStrategyRepo, &sql.DB{})

	t.Run("successful stock assignment", func(t *testing.T) {
		strategyID := uuid.New()
		stockID := uuid.New()
		userID := uuid.New()

		strategy := &models.Strategy{
			ID:     strategyID,
			UserID: userID,
			Name:   "Growth Strategy",
		}

		stock := &models.Stock{
			ID:     stockID,
			Ticker: "AAPL",
			Name:   "Apple Inc.",
		}

		mockStrategyRepo.On("GetByID", mock.Anything, strategyID, userID).Return(strategy, nil)
		mockStockRepo.On("GetByID", mock.Anything, stockID).Return(stock, nil)
		mockStrategyRepo.On("AddStockToStrategy", mock.Anything, strategyID, stockID).Return(nil)

		err := service.AddStockToStrategy(context.Background(), strategyID, stockID, userID)

		assert.NoError(t, err)
		mockStrategyRepo.AssertExpectations(t)
		mockStockRepo.AssertExpectations(t)
	})

	t.Run("strategy not found or access denied", func(t *testing.T) {
		strategyID := uuid.New()
		stockID := uuid.New()
		userID := uuid.New()

		mockStrategyRepo.On("GetByID", mock.Anything, strategyID, userID).Return(nil, &models.NotFoundError{Resource: "strategy"})

		err := service.AddStockToStrategy(context.Background(), strategyID, stockID, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "strategy not found or access denied")
		mockStrategyRepo.AssertExpectations(t)
	})
}

// Helper function to create string pointers for stock tests
func stockStringPtr(s string) *string {
	return &s
}