package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"portfolio-app/internal/models"
)

// Mock PortfolioService
type MockPortfolioService struct {
	mock.Mock
}

func (m *MockPortfolioService) GenerateAllocationPreview(ctx context.Context, req *models.AllocationRequest) (*models.AllocationPreview, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockPortfolioService) GenerateAllocationPreviewWithExclusions(ctx context.Context, req *models.AllocationRequest, excludedStocks []uuid.UUID) (*models.AllocationPreview, error) {
	args := m.Called(ctx, req, excludedStocks)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockPortfolioService) ValidateAllocationRequest(req *models.AllocationRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockPortfolioService) CreatePortfolio(ctx context.Context, req *models.CreatePortfolioRequest, userID uuid.UUID) (*models.Portfolio, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioService) GetPortfolio(ctx context.Context, id uuid.UUID) (*models.Portfolio, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioService) GetUserPortfolios(ctx context.Context, userID uuid.UUID) ([]*models.Portfolio, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioService) UpdatePortfolio(ctx context.Context, id uuid.UUID, req *models.UpdatePortfolioRequest) (*models.Portfolio, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func (m *MockPortfolioService) DeletePortfolio(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPortfolioService) UpdatePortfolioNAV(ctx context.Context, portfolioID uuid.UUID) (*models.NAVHistory, error) {
	args := m.Called(ctx, portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NAVHistory), args.Error(1)
}

func (m *MockPortfolioService) GetPortfolioHistory(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]*models.NAVHistory, error) {
	args := m.Called(ctx, portfolioID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NAVHistory), args.Error(1)
}

func (m *MockPortfolioService) GetPortfolioPerformanceMetrics(ctx context.Context, portfolioID uuid.UUID) (*models.PerformanceMetrics, error) {
	args := m.Called(ctx, portfolioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PerformanceMetrics), args.Error(1)
}

func (m *MockPortfolioService) GenerateRebalancePreview(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.AllocationPreview, error) {
	args := m.Called(ctx, portfolioID, newTotalInvestment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AllocationPreview), args.Error(1)
}

func (m *MockPortfolioService) RebalancePortfolio(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.Portfolio, error) {
	args := m.Called(ctx, portfolioID, newTotalInvestment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Portfolio), args.Error(1)
}

func setupTestApp(mockService *MockPortfolioService) *fiber.App {
	app := fiber.New()
	handler := NewPortfolioHandler(mockService)

	// Setup routes
	api := app.Group("/api")
	portfolios := api.Group("/portfolios")

	portfolios.Post("/preview", handler.GenerateAllocationPreview)
	portfolios.Post("/preview/exclude", handler.GenerateAllocationPreviewWithExclusions)
	portfolios.Post("/validate", handler.ValidateAllocationRequest)
	portfolios.Post("/", handler.CreatePortfolio)
	portfolios.Get("/", handler.GetUserPortfolios)
	portfolios.Get("/:id", handler.GetPortfolio)
	portfolios.Put("/:id", handler.UpdatePortfolio)
	portfolios.Delete("/:id", handler.DeletePortfolio)
	portfolios.Get("/:id/history", handler.GetPortfolioHistory)
	portfolios.Get("/:id/performance", handler.GetPortfolioPerformance)
	portfolios.Post("/:id/nav/update", handler.UpdatePortfolioNAV)
	portfolios.Post("/:id/rebalance/preview", handler.GenerateRebalancePreview)
	portfolios.Post("/:id/rebalance", handler.RebalancePortfolio)
	portfolios.Delete("/cache", handler.ClearAllocationCache)

	return app
}

func TestPortfolioHandler_GenerateAllocationPreview(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	// Test data
	req := models.AllocationRequest{
		StrategyIDs:     []uuid.UUID{uuid.New()},
		TotalInvestment: decimal.NewFromFloat(10000.00),
		Constraints: models.AllocationConstraints{
			MaxAllocationPerStock: decimal.NewFromFloat(20.0),
			MinAllocationAmount:   decimal.NewFromFloat(100.0),
		},
	}

	expectedPreview := &models.AllocationPreview{
		TotalInvestment: decimal.NewFromFloat(10000.00),
		Allocations: []models.StockAllocation{
			{
				StockID:         uuid.New(),
				Ticker:          "AAPL",
				Name:            "Apple Inc.",
				Weight:          decimal.NewFromFloat(100.0),
				AllocationValue: decimal.NewFromFloat(10000.00),
				Price:           decimal.NewFromFloat(150.00),
				Quantity:        66,
				ActualValue:     decimal.NewFromFloat(9900.00),
			},
		},
		UnallocatedCash: decimal.NewFromFloat(100.00),
		TotalAllocated:  decimal.NewFromFloat(9900.00),
	}

	// Setup expectations
	mockService.On("GenerateAllocationPreview", mock.Anything, mock.MatchedBy(func(r *models.AllocationRequest) bool {
		return len(r.StrategyIDs) == 1 && r.TotalInvestment.Equal(decimal.NewFromFloat(10000.00))
	})).Return(expectedPreview, nil)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/portfolios/preview", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "cached")

	mockService.AssertExpectations(t)
}

func TestPortfolioHandler_CreatePortfolio(t *testing.T) {
	mockService := &MockPortfolioService{}
	
	// Create app with middleware first
	app := fiber.New()
	
	// Add middleware to set user ID
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		return c.Next()
	})
	
	// Setup handler and routes
	handler := NewPortfolioHandler(mockService)
	api := app.Group("/api")
	portfolios := api.Group("/portfolios")
	portfolios.Post("/", handler.CreatePortfolio)

	// Test data
	req := models.CreatePortfolioRequest{
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		Positions: []models.CreatePositionRequest{
			{
				StockID:         uuid.New(),
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
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Setup expectations
	mockService.On("CreatePortfolio", mock.Anything, mock.MatchedBy(func(r *models.CreatePortfolioRequest) bool {
		return r.Name == "Test Portfolio" && r.TotalInvestment.Equal(decimal.NewFromFloat(10000.00))
	}), mock.AnythingOfType("uuid.UUID")).Return(expectedPortfolio, nil)

	// Create request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/portfolios", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")

	mockService.AssertExpectations(t)
}

func TestPortfolioHandler_GetPortfolio(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	portfolioID := uuid.New()
	expectedPortfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Setup expectations
	mockService.On("GetPortfolio", mock.Anything, portfolioID).Return(expectedPortfolio, nil)

	// Create request
	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/api/portfolios/%s", portfolioID.String()), nil)

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")

	mockService.AssertExpectations(t)
}

func TestPortfolioHandler_GetPortfolio_NotFound(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	portfolioID := uuid.New()

	// Setup expectations
	mockService.On("GetPortfolio", mock.Anything, portfolioID).Return(nil, fmt.Errorf("portfolio not found"))

	// Create request
	httpReq := httptest.NewRequest("GET", fmt.Sprintf("/api/portfolios/%s", portfolioID.String()), nil)

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Portfolio not found", response["error"])

	mockService.AssertExpectations(t)
}

func TestPortfolioHandler_UpdatePortfolioNAV(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	portfolioID := uuid.New()
	expectedNAV := &models.NAVHistory{
		PortfolioID: portfolioID,
		Timestamp:   time.Now(),
		NAV:         decimal.NewFromFloat(11000.00),
		PnL:         decimal.NewFromFloat(1000.00),
		CreatedAt:   time.Now(),
	}

	// Setup expectations
	mockService.On("UpdatePortfolioNAV", mock.Anything, portfolioID).Return(expectedNAV, nil)

	// Create request
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/portfolios/%s/nav/update", portfolioID.String()), nil)

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")

	mockService.AssertExpectations(t)
}

func TestPortfolioHandler_GenerateRebalancePreview(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	portfolioID := uuid.New()
	newInvestment := decimal.NewFromFloat(20000.00)

	reqBody := map[string]interface{}{
		"new_total_investment": 20000.00,
	}

	expectedPreview := &models.AllocationPreview{
		TotalInvestment: newInvestment,
		Allocations: []models.StockAllocation{
			{
				StockID:         uuid.New(),
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
	mockService.On("GenerateRebalancePreview", mock.Anything, portfolioID, mock.MatchedBy(func(d decimal.Decimal) bool {
		return d.Equal(decimal.NewFromFloat(20000.00))
	})).Return(expectedPreview, nil)

	// Create request
	reqBodyBytes, _ := json.Marshal(reqBody)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/portfolios/%s/rebalance/preview", portfolioID.String()), bytes.NewReader(reqBodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")

	mockService.AssertExpectations(t)
}

func TestPortfolioHandler_RebalancePortfolio(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	portfolioID := uuid.New()
	newInvestment := decimal.NewFromFloat(20000.00)

	reqBody := map[string]interface{}{
		"new_total_investment": 20000.00,
	}

	expectedPortfolio := &models.Portfolio{
		ID:              portfolioID,
		UserID:          uuid.New(),
		Name:            "Test Portfolio",
		TotalInvestment: newInvestment,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Setup expectations
	mockService.On("RebalancePortfolio", mock.Anything, portfolioID, mock.MatchedBy(func(d decimal.Decimal) bool {
		return d.Equal(decimal.NewFromFloat(20000.00))
	})).Return(expectedPortfolio, nil)

	// Create request
	reqBodyBytes, _ := json.Marshal(reqBody)
	httpReq := httptest.NewRequest("POST", fmt.Sprintf("/api/portfolios/%s/rebalance", portfolioID.String()), bytes.NewReader(reqBodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")

	mockService.AssertExpectations(t)
}

func TestPortfolioHandler_DeletePortfolio(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	portfolioID := uuid.New()

	// Setup expectations
	mockService.On("DeletePortfolio", mock.Anything, portfolioID).Return(nil)

	// Create request
	httpReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/portfolios/%s", portfolioID.String()), nil)

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestPortfolioHandler_InvalidPortfolioID(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	// Create request with invalid UUID
	httpReq := httptest.NewRequest("GET", "/api/portfolios/invalid-uuid", nil)

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid portfolio ID", response["error"])
}

func TestPortfolioHandler_CreatePortfolio_NoUserContext(t *testing.T) {
	mockService := &MockPortfolioService{}
	app := setupTestApp(mockService)

	// Test data
	req := models.CreatePortfolioRequest{
		Name:            "Test Portfolio",
		TotalInvestment: decimal.NewFromFloat(10000.00),
		Positions: []models.CreatePositionRequest{
			{
				StockID:         uuid.New(),
				Quantity:        100,
				EntryPrice:      decimal.NewFromFloat(100.00),
				AllocationValue: decimal.NewFromFloat(10000.00),
				StrategyContrib: map[string]decimal.Decimal{
					"strategy1": decimal.NewFromFloat(10000.00),
				},
			},
		},
	}

	// Create request without user context
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/portfolios", bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute
	resp, err := app.Test(httpReq)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "User authentication required", response["error"])
}