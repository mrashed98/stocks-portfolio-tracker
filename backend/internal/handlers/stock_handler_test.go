package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"portfolio-app/internal/models"
)

// MockStockService is a mock implementation of StockService
type MockStockService struct {
	mock.Mock
}

func (m *MockStockService) CreateStock(ctx context.Context, req *models.CreateStockRequest) (*models.Stock, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockStockService) UpdateStock(ctx context.Context, id uuid.UUID, req *models.UpdateStockRequest) (*models.Stock, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockStockService) GetStock(ctx context.Context, id uuid.UUID) (*models.Stock, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockStockService) GetStockByTicker(ctx context.Context, ticker string) (*models.Stock, error) {
	args := m.Called(ctx, ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}

func (m *MockStockService) GetStocks(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error) {
	args := m.Called(ctx, search, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Stock), args.Error(1)
}

func (m *MockStockService) GetStocksWithSignals(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error) {
	args := m.Called(ctx, search, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Stock), args.Error(1)
}

func (m *MockStockService) DeleteStock(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStockService) UpdateStockSignal(ctx context.Context, stockID uuid.UUID, signal models.SignalType) (*models.Signal, error) {
	args := m.Called(ctx, stockID, signal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Signal), args.Error(1)
}

func (m *MockStockService) GetStockSignalHistory(ctx context.Context, stockID uuid.UUID, from, to time.Time) ([]*models.Signal, error) {
	args := m.Called(ctx, stockID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Signal), args.Error(1)
}

func (m *MockStockService) ValidateTickerSymbol(ticker string) error {
	args := m.Called(ticker)
	return args.Error(0)
}

func (m *MockStockService) AddStockToStrategy(ctx context.Context, strategyID, stockID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, strategyID, stockID, userID)
	return args.Error(0)
}

func (m *MockStockService) RemoveStockFromStrategy(ctx context.Context, strategyID, stockID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, strategyID, stockID, userID)
	return args.Error(0)
}

func TestStockHandler_CreateStock(t *testing.T) {
	mockService := new(MockStockService)
	handler := NewStockHandler(mockService)

	t.Run("successful stock creation", func(t *testing.T) {
		app := fiber.New()
		app.Post("/stocks", handler.CreateStock)

		req := models.CreateStockRequest{
			Ticker: "AAPL",
			Name:   "Apple Inc.",
			Sector: stockHandlerStringPtr("Technology"),
		}

		expectedStock := &models.Stock{
			ID:     uuid.New(),
			Ticker: "AAPL",
			Name:   "Apple Inc.",
			Sector: stockHandlerStringPtr("Technology"),
		}

		mockService.On("CreateStock", mock.Anything, mock.AnythingOfType("*models.CreateStockRequest")).Return(expectedStock, nil)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/stocks", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		app.Post("/stocks", handler.CreateStock)

		httpReq := httptest.NewRequest("POST", "/stocks", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation error from struct validation", func(t *testing.T) {
		app := fiber.New()
		app.Post("/stocks", handler.CreateStock)

		// This will fail struct validation before reaching the service
		req := models.CreateStockRequest{
			Ticker: "", // Empty ticker will fail required validation
			Name:   "Invalid Company",
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/stocks", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation error from service", func(t *testing.T) {
		// Create a fresh mock service for this test
		mockService2 := new(MockStockService)
		handler2 := NewStockHandler(mockService2)
		
		app := fiber.New()
		app.Post("/stocks", handler2.CreateStock)

		req := models.CreateStockRequest{
			Ticker: "AAPL",
			Name:   "Apple Inc.",
		}

		validationErr := &models.ValidationError{
			Field:   "ticker",
			Message: "Stock with ticker AAPL already exists",
		}

		mockService2.On("CreateStock", mock.Anything, mock.AnythingOfType("*models.CreateStockRequest")).Return(nil, validationErr)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/stocks", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

		mockService2.AssertExpectations(t)
	})
}

func TestStockHandler_GetStocks(t *testing.T) {
	mockService := new(MockStockService)
	handler := NewStockHandler(mockService)

	t.Run("successful get stocks", func(t *testing.T) {
		app := fiber.New()
		app.Get("/stocks", handler.GetStocks)

		expectedStocks := []*models.Stock{
			{
				ID:     uuid.New(),
				Ticker: "AAPL",
				Name:   "Apple Inc.",
			},
			{
				ID:     uuid.New(),
				Ticker: "GOOGL",
				Name:   "Alphabet Inc.",
			},
		}

		mockService.On("GetStocks", mock.Anything, "", 50, 0).Return(expectedStocks, nil)

		httpReq := httptest.NewRequest("GET", "/stocks", nil)
		resp, err := app.Test(httpReq)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("get stocks with search and pagination", func(t *testing.T) {
		app := fiber.New()
		app.Get("/stocks", handler.GetStocks)

		expectedStocks := []*models.Stock{
			{
				ID:     uuid.New(),
				Ticker: "AAPL",
				Name:   "Apple Inc.",
			},
		}

		mockService.On("GetStocks", mock.Anything, "apple", 10, 5).Return(expectedStocks, nil)

		httpReq := httptest.NewRequest("GET", "/stocks?search=apple&limit=10&offset=5", nil)
		resp, err := app.Test(httpReq)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("get stocks with signals", func(t *testing.T) {
		app := fiber.New()
		app.Get("/stocks", handler.GetStocks)

		expectedStocks := []*models.Stock{
			{
				ID:     uuid.New(),
				Ticker: "AAPL",
				Name:   "Apple Inc.",
				CurrentSignal: &models.Signal{
					Signal: models.SignalBuy,
					Date:   time.Now(),
				},
			},
		}

		mockService.On("GetStocksWithSignals", mock.Anything, "", 50, 0).Return(expectedStocks, nil)

		httpReq := httptest.NewRequest("GET", "/stocks?include_signals=true", nil)
		resp, err := app.Test(httpReq)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestStockHandler_UpdateStockSignal(t *testing.T) {
	mockService := new(MockStockService)
	handler := NewStockHandler(mockService)

	t.Run("successful signal update", func(t *testing.T) {
		app := fiber.New()
		app.Put("/stocks/:id/signal", handler.UpdateStockSignal)

		stockID := uuid.New()
		req := models.UpdateSignalRequest{
			Signal: models.SignalBuy,
		}

		expectedSignal := &models.Signal{
			StockID: stockID,
			Signal:  models.SignalBuy,
			Date:    time.Now(),
		}

		mockService.On("UpdateStockSignal", mock.Anything, stockID, models.SignalBuy).Return(expectedSignal, nil)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/stocks/"+stockID.String()+"/signal", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid stock ID", func(t *testing.T) {
		app := fiber.New()
		app.Put("/stocks/:id/signal", handler.UpdateStockSignal)

		req := models.UpdateSignalRequest{
			Signal: models.SignalBuy,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/stocks/invalid-id/signal", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("stock not found", func(t *testing.T) {
		app := fiber.New()
		app.Put("/stocks/:id/signal", handler.UpdateStockSignal)

		stockID := uuid.New()
		req := models.UpdateSignalRequest{
			Signal: models.SignalBuy,
		}

		notFoundErr := &models.NotFoundError{Resource: "stock"}
		mockService.On("UpdateStockSignal", mock.Anything, stockID, models.SignalBuy).Return(nil, notFoundErr)

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("PUT", "/stocks/"+stockID.String()+"/signal", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(httpReq)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		mockService.AssertExpectations(t)
	})
}

// Helper function to create string pointers for stock handler tests
func stockHandlerStringPtr(s string) *string {
	return &s
}