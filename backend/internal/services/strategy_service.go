package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"portfolio-app/internal/models"
	"portfolio-app/internal/repositories"
)

// StrategyService defines the interface for strategy operations
type StrategyService interface {
	CreateStrategy(ctx context.Context, req *models.CreateStrategyRequest, userID uuid.UUID) (*models.Strategy, error)
	UpdateStrategy(ctx context.Context, id uuid.UUID, req *models.UpdateStrategyRequest, userID uuid.UUID) (*models.Strategy, error)
	GetStrategy(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Strategy, error)
	GetUserStrategies(ctx context.Context, userID uuid.UUID) ([]*models.Strategy, error)
	DeleteStrategy(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	UpdateStockEligibility(ctx context.Context, strategyID, stockID uuid.UUID, eligible bool, userID uuid.UUID) error
	ValidateStrategyWeights(ctx context.Context, userID uuid.UUID, excludeStrategyID *uuid.UUID) error
}

// strategyService implements the StrategyService interface
type strategyService struct {
	strategyRepo repositories.StrategyRepository
	db           *sql.DB
}

// NewStrategyService creates a new strategy service instance
func NewStrategyService(strategyRepo repositories.StrategyRepository, db *sql.DB) StrategyService {
	return &strategyService{
		strategyRepo: strategyRepo,
		db:           db,
	}
}

// CreateStrategy creates a new strategy with weight validation
func (s *strategyService) CreateStrategy(ctx context.Context, req *models.CreateStrategyRequest, userID uuid.UUID) (*models.Strategy, error) {
	// Validate percentage mode weight constraints
	if req.WeightMode == models.WeightModePercent {
		if err := s.validatePercentageWeights(ctx, userID, nil, req.WeightValue); err != nil {
			return nil, err
		}
	}

	strategy := &models.Strategy{}
	strategy.FromCreateRequest(req, userID)

	createdStrategy, err := s.strategyRepo.Create(ctx, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to create strategy: %w", err)
	}

	return createdStrategy, nil
}

// UpdateStrategy updates an existing strategy with weight validation
func (s *strategyService) UpdateStrategy(ctx context.Context, id uuid.UUID, req *models.UpdateStrategyRequest, userID uuid.UUID) (*models.Strategy, error) {
	// Get existing strategy
	existingStrategy, err := s.strategyRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy: %w", err)
	}

	// Validate percentage mode weight constraints if weight is being updated
	if req.WeightMode != nil && *req.WeightMode == models.WeightModePercent && req.WeightValue != nil {
		if err := s.validatePercentageWeights(ctx, userID, &id, *req.WeightValue); err != nil {
			return nil, err
		}
	} else if req.WeightValue != nil && existingStrategy.WeightMode == models.WeightModePercent {
		if err := s.validatePercentageWeights(ctx, userID, &id, *req.WeightValue); err != nil {
			return nil, err
		}
	}

	// Apply updates
	existingStrategy.ApplyUpdate(req)

	updatedStrategy, err := s.strategyRepo.Update(ctx, existingStrategy)
	if err != nil {
		return nil, fmt.Errorf("failed to update strategy: %w", err)
	}

	return updatedStrategy, nil
}

// GetStrategy retrieves a strategy by ID
func (s *strategyService) GetStrategy(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Strategy, error) {
	strategy, err := s.strategyRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy: %w", err)
	}

	return strategy, nil
}

// GetUserStrategies retrieves all strategies for a user
func (s *strategyService) GetUserStrategies(ctx context.Context, userID uuid.UUID) ([]*models.Strategy, error) {
	strategies, err := s.strategyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user strategies: %w", err)
	}

	return strategies, nil
}

// DeleteStrategy deletes a strategy
func (s *strategyService) DeleteStrategy(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if err := s.strategyRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete strategy: %w", err)
	}

	return nil
}

// UpdateStockEligibility updates the eligibility of a stock within a strategy
func (s *strategyService) UpdateStockEligibility(ctx context.Context, strategyID, stockID uuid.UUID, eligible bool, userID uuid.UUID) error {
	// Verify strategy belongs to user
	_, err := s.strategyRepo.GetByID(ctx, strategyID, userID)
	if err != nil {
		return fmt.Errorf("strategy not found or access denied: %w", err)
	}

	if err := s.strategyRepo.UpdateStockEligibility(ctx, strategyID, stockID, eligible); err != nil {
		return fmt.Errorf("failed to update stock eligibility: %w", err)
	}

	return nil
}

// ValidateStrategyWeights validates that percentage mode strategies don't exceed 100%
func (s *strategyService) ValidateStrategyWeights(ctx context.Context, userID uuid.UUID, excludeStrategyID *uuid.UUID) error {
	return s.validatePercentageWeights(ctx, userID, excludeStrategyID, decimal.Zero)
}

// validatePercentageWeights validates that total percentage weights don't exceed 100%
func (s *strategyService) validatePercentageWeights(ctx context.Context, userID uuid.UUID, excludeStrategyID *uuid.UUID, additionalWeight decimal.Decimal) error {
	strategies, err := s.strategyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user strategies for validation: %w", err)
	}

	totalWeight := decimal.Zero
	for _, strategy := range strategies {
		// Skip the strategy being excluded (for updates)
		if excludeStrategyID != nil && strategy.ID == *excludeStrategyID {
			continue
		}

		// Only count percentage mode strategies
		if strategy.WeightMode == models.WeightModePercent {
			totalWeight = totalWeight.Add(strategy.WeightValue)
		}
	}

	// Add the additional weight (for new strategy or update)
	totalWeight = totalWeight.Add(additionalWeight)

	// Check if total exceeds 100%
	if totalWeight.GreaterThan(decimal.NewFromInt(100)) {
		return &models.ValidationError{
			Field:   "weight_value",
			Message: fmt.Sprintf("Total percentage weights cannot exceed 100%%. Current total would be %s%%", totalWeight.String()),
		}
	}

	return nil
}