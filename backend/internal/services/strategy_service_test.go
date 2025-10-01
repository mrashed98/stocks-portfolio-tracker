package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"portfolio-app/internal/models"
)

// MockStrategyRepository is a mock implementation of StrategyRepository
type MockStrategyRepository struct {
	mock.Mock
}

func (m *MockStrategyRepository) Create(ctx context.Context, strategy *models.Strategy) (*models.Strategy, error) {
	args := m.Called(ctx, strategy)
	return args.Get(0).(*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) Update(ctx context.Context, strategy *models.Strategy) (*models.Strategy, error) {
	args := m.Called(ctx, strategy)
	return args.Get(0).(*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Strategy, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Strategy, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Strategy, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Strategy), args.Error(1)
}

func (m *MockStrategyRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockStrategyRepository) AddStockToStrategy(ctx context.Context, strategyID, stockID uuid.UUID) error {
	args := m.Called(ctx, strategyID, stockID)
	return args.Error(0)
}

func (m *MockStrategyRepository) RemoveStockFromStrategy(ctx context.Context, strategyID, stockID uuid.UUID) error {
	args := m.Called(ctx, strategyID, stockID)
	return args.Error(0)
}

func (m *MockStrategyRepository) UpdateStockEligibility(ctx context.Context, strategyID, stockID uuid.UUID, eligible bool) error {
	args := m.Called(ctx, strategyID, stockID, eligible)
	return args.Error(0)
}

func (m *MockStrategyRepository) GetStrategyStocks(ctx context.Context, strategyID uuid.UUID) ([]*models.StrategyStock, error) {
	args := m.Called(ctx, strategyID)
	return args.Get(0).([]*models.StrategyStock), args.Error(1)
}

func TestStrategyService_CreateStrategy(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	t.Run("successful creation with budget mode", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		
		req := &models.CreateStrategyRequest{
			Name:        "Test Strategy",
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(1000),
		}

		expectedStrategy := &models.Strategy{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        req.Name,
			WeightMode:  req.WeightMode,
			WeightValue: req.WeightValue,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Strategy")).Return(expectedStrategy, nil)

		result, err := service.CreateStrategy(ctx, req, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedStrategy.Name, result.Name)
		assert.Equal(t, expectedStrategy.WeightMode, result.WeightMode)
		assert.Equal(t, expectedStrategy.WeightValue, result.WeightValue)
		assert.Equal(t, expectedStrategy.UserID, result.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful creation with percentage mode", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository) // Create fresh mock for this test
		service := NewStrategyService(mockRepo, &sql.DB{})
		
		req := &models.CreateStrategyRequest{
			Name:        "Test Strategy",
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(50),
		}

		expectedStrategy := &models.Strategy{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        req.Name,
			WeightMode:  req.WeightMode,
			WeightValue: req.WeightValue,
		}

		// Mock the validation call
		mockRepo.On("GetByUserID", ctx, userID).Return([]*models.Strategy{}, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Strategy")).Return(expectedStrategy, nil)

		result, err := service.CreateStrategy(ctx, req, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedStrategy.Name, result.Name)
		assert.Equal(t, expectedStrategy.WeightMode, result.WeightMode)
		assert.Equal(t, expectedStrategy.WeightValue, result.WeightValue)
		assert.Equal(t, expectedStrategy.UserID, result.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("percentage mode validation failure", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository) // Create fresh mock for this test
		service := NewStrategyService(mockRepo, &sql.DB{})
		
		req := &models.CreateStrategyRequest{
			Name:        "Test Strategy",
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(60),
		}

		// Mock existing strategies that total 50%
		existingStrategies := []*models.Strategy{
			{
				ID:          uuid.New(),
				UserID:      userID,
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(50),
			},
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(existingStrategies, nil)

		result, err := service.CreateStrategy(ctx, req, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Total percentage weights cannot exceed 100%")
		mockRepo.AssertExpectations(t)
	})
}

func TestStrategyService_UpdateStrategy(t *testing.T) {
	userID := uuid.New()
	strategyID := uuid.New()
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		existingStrategy := &models.Strategy{
			ID:          strategyID,
			UserID:      userID,
			Name:        "Old Name",
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(1000),
		}

		newName := "New Name"
		req := &models.UpdateStrategyRequest{
			Name: &newName,
		}

		updatedStrategy := &models.Strategy{
			ID:          strategyID,
			UserID:      userID,
			Name:        newName,
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(1000),
		}

		mockRepo.On("GetByID", ctx, strategyID, userID).Return(existingStrategy, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Strategy")).Return(updatedStrategy, nil)

		result, err := service.UpdateStrategy(ctx, strategyID, req, userID)

		assert.NoError(t, err)
		assert.Equal(t, updatedStrategy, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("strategy not found", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		
		req := &models.UpdateStrategyRequest{
			Name: stringPtr("New Name"),
		}

		mockRepo.On("GetByID", ctx, strategyID, userID).Return(nil, &models.NotFoundError{Resource: "strategy"})

		result, err := service.UpdateStrategy(ctx, strategyID, req, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("percentage weight validation failure on update", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		
		existingStrategy := &models.Strategy{
			ID:          strategyID,
			UserID:      userID,
			Name:        "Test Strategy",
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(30),
		}

		newWeight := decimal.NewFromInt(80)
		req := &models.UpdateStrategyRequest{
			WeightValue: &newWeight,
		}

		// Mock existing strategies (excluding the one being updated)
		otherStrategies := []*models.Strategy{
			{
				ID:          uuid.New(),
				UserID:      userID,
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(50),
			},
		}

		mockRepo.On("GetByID", ctx, strategyID, userID).Return(existingStrategy, nil)
		mockRepo.On("GetByUserID", ctx, userID).Return(append(otherStrategies, existingStrategy), nil)

		result, err := service.UpdateStrategy(ctx, strategyID, req, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Total percentage weights cannot exceed 100%")
		mockRepo.AssertExpectations(t)
	})
}

func TestStrategyService_GetStrategy(t *testing.T) {
	userID := uuid.New()
	strategyID := uuid.New()
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		expectedStrategy := &models.Strategy{
			ID:          strategyID,
			UserID:      userID,
			Name:        "Test Strategy",
			WeightMode:  models.WeightModeBudget,
			WeightValue: decimal.NewFromInt(1000),
		}

		mockRepo.On("GetByID", ctx, strategyID, userID).Return(expectedStrategy, nil)

		result, err := service.GetStrategy(ctx, strategyID, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedStrategy, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("strategy not found", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		
		mockRepo.On("GetByID", ctx, strategyID, userID).Return(nil, &models.NotFoundError{Resource: "strategy"})

		result, err := service.GetStrategy(ctx, strategyID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestStrategyService_UpdateStockEligibility(t *testing.T) {
	userID := uuid.New()
	strategyID := uuid.New()
	stockID := uuid.New()
	ctx := context.Background()

	t.Run("successful eligibility update", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		strategy := &models.Strategy{
			ID:     strategyID,
			UserID: userID,
		}

		mockRepo.On("GetByID", ctx, strategyID, userID).Return(strategy, nil)
		mockRepo.On("UpdateStockEligibility", ctx, strategyID, stockID, true).Return(nil)

		err := service.UpdateStockEligibility(ctx, strategyID, stockID, true, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("strategy not found", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		
		mockRepo.On("GetByID", ctx, strategyID, userID).Return(nil, &models.NotFoundError{Resource: "strategy"})

		err := service.UpdateStockEligibility(ctx, strategyID, stockID, true, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "strategy not found or access denied")
		mockRepo.AssertExpectations(t)
	})
}

func TestStrategyService_ValidateStrategyWeights(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	t.Run("valid weights", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		strategies := []*models.Strategy{
			{
				ID:          uuid.New(),
				UserID:      userID,
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(50),
			},
			{
				ID:          uuid.New(),
				UserID:      userID,
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(30),
			},
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(strategies, nil)

		err := service.ValidateStrategyWeights(ctx, userID, nil)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid weights exceed 100%", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})
		
		strategies := []*models.Strategy{
			{
				ID:          uuid.New(),
				UserID:      userID,
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(60),
			},
			{
				ID:          uuid.New(),
				UserID:      userID,
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(50),
			},
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(strategies, nil)

		err := service.ValidateStrategyWeights(ctx, userID, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Total percentage weights cannot exceed 100%")
		mockRepo.AssertExpectations(t)
	})
}

func TestStrategyService_DeleteStrategy(t *testing.T) {
	userID := uuid.New()
	strategyID := uuid.New()
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})

		mockRepo.On("Delete", ctx, strategyID, userID).Return(nil)

		err := service.DeleteStrategy(ctx, strategyID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("strategy not found", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})

		mockRepo.On("Delete", ctx, strategyID, userID).Return(&models.NotFoundError{Resource: "strategy"})

		err := service.DeleteStrategy(ctx, strategyID, userID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestStrategyService_GetUserStrategies(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})

		expectedStrategies := []*models.Strategy{
			{
				ID:          uuid.New(),
				UserID:      userID,
				Name:        "Strategy 1",
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(60),
			},
			{
				ID:          uuid.New(),
				UserID:      userID,
				Name:        "Strategy 2",
				WeightMode:  models.WeightModeBudget,
				WeightValue: decimal.NewFromInt(5000),
			},
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(expectedStrategies, nil)

		result, err := service.GetUserStrategies(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Strategy 1", result[0].Name)
		assert.Equal(t, "Strategy 2", result[1].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no strategies found", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})

		mockRepo.On("GetByUserID", ctx, userID).Return([]*models.Strategy{}, nil)

		result, err := service.GetUserStrategies(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		mockRepo.AssertExpectations(t)
	})
}

// Note: AddStockToStrategy, RemoveStockFromStrategy, and GetStrategyStocks 
// are not part of the StrategyService interface. These operations are likely
// handled by the repository layer directly or through other services.

// Note: ValidateCreateRequest and ValidateUpdateRequest are not part of the 
// StrategyService interface. Validation is likely handled internally within
// the CreateStrategy and UpdateStrategy methods.

func TestStrategyService_MixedWeightModes(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	t.Run("mixed weight modes validation", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		service := NewStrategyService(mockRepo, &sql.DB{})

		// Existing strategies with mixed modes
		strategies := []*models.Strategy{
			{
				ID:          uuid.New(),
				UserID:      userID,
				WeightMode:  models.WeightModePercent,
				WeightValue: decimal.NewFromInt(40),
			},
			{
				ID:          uuid.New(),
				UserID:      userID,
				WeightMode:  models.WeightModeBudget,
				WeightValue: decimal.NewFromInt(3000),
			},
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(strategies, nil)

		// Try to add another percentage strategy that would exceed 100%
		req := &models.CreateStrategyRequest{
			Name:        "New Strategy",
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(70), // 40% + 70% = 110%
		}

		result, err := service.CreateStrategy(ctx, req, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Total percentage weights cannot exceed 100%")
		mockRepo.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkStrategyService_CreateStrategy(b *testing.B) {
	mockRepo := new(MockStrategyRepository)
	service := NewStrategyService(mockRepo, &sql.DB{})
	userID := uuid.New()
	ctx := context.Background()

	req := &models.CreateStrategyRequest{
		Name:        "Benchmark Strategy",
		WeightMode:  models.WeightModeBudget,
		WeightValue: decimal.NewFromInt(5000),
	}

	expectedStrategy := &models.Strategy{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		WeightMode:  req.WeightMode,
		WeightValue: req.WeightValue,
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Strategy")).Return(expectedStrategy, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateStrategy(ctx, req, userID)
	}
}

func BenchmarkStrategyService_ValidateStrategyWeights(b *testing.B) {
	mockRepo := new(MockStrategyRepository)
	service := NewStrategyService(mockRepo, &sql.DB{})
	userID := uuid.New()
	ctx := context.Background()

	strategies := []*models.Strategy{
		{
			ID:          uuid.New(),
			UserID:      userID,
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(50),
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			WeightMode:  models.WeightModePercent,
			WeightValue: decimal.NewFromInt(30),
		},
	}

	mockRepo.On("GetByUserID", mock.Anything, userID).Return(strategies, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ValidateStrategyWeights(ctx, userID, nil)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}