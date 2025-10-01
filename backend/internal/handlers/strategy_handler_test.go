package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"portfolio-app/internal/models"
)

// MockStrategyService is a mock implementation of StrategyService
type MockStrategyService struct {
	mock.Mock
}

func (m *MockStrategyService) CreateStrategy(ctx context.Context, req *models.CreateStrategyRequest, userID uuid.UUID) (*models.Strategy, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Strategy), args.Error(1)
}

func (m *MockStrategyService) UpdateStrategy(ctx context.Context, id uuid.UUID, req *models.UpdateStrategyRequest, userID uuid.UUID) (*models.Strategy, error) {
	args := m.Called(ctx, id, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Strategy), args.Error(1)
}

func (m *MockStrategyService) GetStrategy(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Strategy, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Strategy), args.Error(1)
}

func (m *MockStrategyService) GetUserStrategies(ctx context.Context, userID uuid.UUID) ([]*models.Strategy, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockStrategyService) DeleteStrategy(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockStrategyService) UpdateStockEligibility(ctx context.Context, strategyID, stockID uuid.UUID, eligible bool, userID uuid.UUID) error {
	args := m.Called(ctx, strategyID, stockID, eligible, userID)
	return args.Error(0)
}

func (m *MockStrategyService) ValidateStrategyWeights(ctx context.Context, userID uuid.UUID, excludeStrategyID *uuid.UUID) error {
	args := m.Called(ctx, userID, excludeStrategyID)
	return args.Error(0)
}

func TestStrategyHandler_CreateStrategy(t *testing.T) {
	app := fiber.New()
	mockService := new(MockStrategyService)
	handler := NewStrategyHandler(mockService)

	app.Post("/strategies", handler.CreateStrategy)

	t.Run("successful creation", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		
		// Add middleware to set user context
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("userID", uuid.New().String())
			return c.Next()
		})
		
		app.Post("/strategies", handler.CreateStrategy)

		reqBody := map[string]interface{}{
			"name":         "Test Strategy",
			"weight_mode":  "percent",
			"weight_value": "50.0",
		}
		
		expectedStrategy := &models.Strategy{
			ID:          uuid.New(),
			Name:        "Test Strategy",
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(50),
		}

		mockService.On("CreateStrategy", mock.Anything, mock.AnythingOfType("*models.CreateStrategyRequest"), mock.AnythingOfType("uuid.UUID")).Return(expectedStrategy, nil)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/strategies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		app.Post("/strategies", handler.CreateStrategy)

		req := httptest.NewRequest("POST", "/strategies", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation error", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		app.Post("/strategies", handler.CreateStrategy)

		validationErr := &models.ValidationError{
			Field:   "weight_value",
			Message: "Total percentage weights cannot exceed 100%",
		}

		mockService.On("CreateStrategy", mock.Anything, mock.AnythingOfType("*models.CreateStrategyRequest"), mock.AnythingOfType("uuid.UUID")).Return(nil, validationErr)

		reqBody := map[string]interface{}{
			"name":         "Test Strategy",
			"weight_mode":  "percent",
			"weight_value": "60.0",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/strategies", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

		mockService.AssertExpectations(t)
	})
}

func TestStrategyHandler_GetStrategies(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		app.Get("/strategies", handler.GetStrategies)

		strategies := []*models.Strategy{
			{
				ID:          uuid.New(),
				Name:        "Strategy 1",
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(50),
			},
			{
				ID:          uuid.New(),
				Name:        "Strategy 2",
				WeightMode:  models.WeightModeBudget,
				WeightValue: decimal.NewFromInt(1000),
			},
		}

		mockService.On("GetUserStrategies", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(strategies, nil)

		req := httptest.NewRequest("GET", "/strategies", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestStrategyHandler_UpdateStrategyWeight(t *testing.T) {
	t.Run("successful weight update", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		app.Put("/strategies/:id/weight", handler.UpdateStrategyWeight)

		strategyID := uuid.New()
		updatedStrategy := &models.Strategy{
			ID:          strategyID,
			Name:        "Test Strategy",
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(75),
		}

		mockService.On("UpdateStrategy", mock.Anything, strategyID, mock.AnythingOfType("*models.UpdateStrategyRequest"), mock.AnythingOfType("uuid.UUID")).Return(updatedStrategy, nil)

		reqBody := map[string]interface{}{
			"weight_value": "75.0",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/strategies/"+strategyID.String()+"/weight", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid strategy ID", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		app.Put("/strategies/:id/weight", handler.UpdateStrategyWeight)

		reqBody := map[string]interface{}{
			"weight_value": "75.0",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/strategies/invalid-uuid/weight", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid weight value", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		app.Put("/strategies/:id/weight", handler.UpdateStrategyWeight)

		strategyID := uuid.New()
		reqBody := map[string]interface{}{
			"weight_value": "invalid",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/strategies/"+strategyID.String()+"/weight", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestStrategyHandler_UpdateStockEligibility(t *testing.T) {
	t.Run("successful eligibility update", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		app.Put("/strategies/:id/stocks/:stockId", handler.UpdateStockEligibility)

		strategyID := uuid.New()
		stockID := uuid.New()

		mockService.On("UpdateStockEligibility", mock.Anything, strategyID, stockID, true, mock.AnythingOfType("uuid.UUID")).Return(nil)

		reqBody := map[string]interface{}{
			"eligible": true,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/strategies/"+strategyID.String()+"/stocks/"+stockID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("strategy not found", func(t *testing.T) {
		mockService := new(MockStrategyService)
		handler := NewStrategyHandler(mockService)
		app := fiber.New()
		app.Put("/strategies/:id/stocks/:stockId", handler.UpdateStockEligibility)

		strategyID := uuid.New()
		stockID := uuid.New()

		mockService.On("UpdateStockEligibility", mock.Anything, strategyID, stockID, true, mock.AnythingOfType("uuid.UUID")).Return(&models.NotFoundError{Resource: "strategy"})

		reqBody := map[string]interface{}{
			"eligible": true,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/strategies/"+strategyID.String()+"/stocks/"+stockID.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		mockService.AssertExpectations(t)
	})
}