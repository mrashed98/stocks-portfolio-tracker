package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"portfolio-app/internal/models"
)

// AllocationEngine handles portfolio allocation calculations
type AllocationEngine struct {
	strategyRepo        StrategyRepository
	stockRepo           StockRepository
	signalRepo          SignalRepository
	marketDataService   MarketDataService
	constraintValidator *ConstraintValidator
}

// AllocationEngineInterface defines the allocation engine contract
type AllocationEngineInterface interface {
	CalculateAllocations(ctx context.Context, req *models.AllocationRequest) (*models.AllocationPreview, error)
	RebalanceAllocations(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.AllocationPreview, error)
	RecalculateWithExclusions(ctx context.Context, originalReq *models.AllocationRequest, excludedStocks []uuid.UUID) (*models.AllocationPreview, error)
	ValidateConstraints(allocations []models.StockAllocation, constraints models.AllocationConstraints) error
	ValidateConstraintsDetailed(allocations []models.StockAllocation, constraints models.AllocationConstraints, totalInvestment decimal.Decimal) *ValidationResult
	ValidateConstraintsConfig(constraints models.AllocationConstraints, totalInvestment decimal.Decimal) *ValidationResult
}

// StrategyRepository interface for strategy data access
type StrategyRepository interface {
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Strategy, error)
	GetStrategyStocks(ctx context.Context, strategyID uuid.UUID) ([]*models.StrategyStock, error)
}

// StockRepository interface for stock data access
type StockRepository interface {
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Stock, error)
}

// SignalRepository interface for signal data access
type SignalRepository interface {
	GetLatestSignals(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*models.Signal, error)
}

// NewAllocationEngine creates a new allocation engine
func NewAllocationEngine(
	strategyRepo StrategyRepository,
	stockRepo StockRepository,
	signalRepo SignalRepository,
	marketDataService MarketDataService,
) *AllocationEngine {
	return &AllocationEngine{
		strategyRepo:        strategyRepo,
		stockRepo:           stockRepo,
		signalRepo:          signalRepo,
		marketDataService:   marketDataService,
		constraintValidator: NewConstraintValidator(),
	}
}

// CalculateAllocations calculates portfolio allocations based on strategies and constraints
func (e *AllocationEngine) CalculateAllocations(ctx context.Context, req *models.AllocationRequest) (*models.AllocationPreview, error) {
	// 1. Get strategies
	strategies, err := e.strategyRepo.GetByIDs(ctx, req.StrategyIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategies: %w", err)
	}

	if len(strategies) == 0 {
		return nil, fmt.Errorf("no strategies found")
	}

	// 2. Calculate strategy weights
	strategyAllocations, err := e.calculateStrategyWeights(strategies, req.TotalInvestment)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate strategy weights: %w", err)
	}

	// 3. Get all stocks from strategies and their signals
	stockAllocations, err := e.distributeToStocks(ctx, strategies, strategyAllocations, req.ExcludedStocks)
	if err != nil {
		return nil, fmt.Errorf("failed to distribute to stocks: %w", err)
	}

	// 4. Validate constraints configuration first
	configValidation := e.ValidateConstraintsConfig(req.Constraints, req.TotalInvestment)
	if !configValidation.IsValid {
		return nil, fmt.Errorf("invalid constraints configuration: %s", configValidation.Violations[0].Message)
	}

	// 5. Apply constraints
	constrainedAllocations, err := e.applyConstraints(stockAllocations, req.Constraints, req.TotalInvestment)
	if err != nil {
		return nil, fmt.Errorf("failed to apply constraints: %w", err)
	}

	// 6. Normalize allocations to match total investment
	finalAllocations := e.normalizeAllocations(constrainedAllocations, req.TotalInvestment)

	// 7. Fetch real-time prices and calculate quantities
	finalAllocationsWithPrices, err := e.addPricesAndQuantities(ctx, finalAllocations)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prices and calculate quantities: %w", err)
	}

	// 8. Perform detailed validation and provide warnings if needed
	validation := e.ValidateConstraintsDetailed(finalAllocationsWithPrices, req.Constraints, req.TotalInvestment)
	if !validation.IsValid {
		// Log warnings but don't fail - let the user decide
		// In a real implementation, you might want to return warnings in the response
		fmt.Printf("Allocation warnings: %+v\n", validation.Violations)
	}

	// 9. Calculate unallocated cash based on actual quantities
	totalAllocated := decimal.Zero
	for _, allocation := range finalAllocationsWithPrices {
		totalAllocated = totalAllocated.Add(allocation.ActualValue)
	}
	unallocatedCash := req.TotalInvestment.Sub(totalAllocated)

	return &models.AllocationPreview{
		TotalInvestment: req.TotalInvestment,
		Allocations:     finalAllocationsWithPrices,
		UnallocatedCash: unallocatedCash,
		TotalAllocated:  totalAllocated,
		Constraints:     req.Constraints,
	}, nil
}

// calculateStrategyWeights calculates the allocation amount for each strategy
func (e *AllocationEngine) calculateStrategyWeights(strategies []*models.Strategy, totalInvestment decimal.Decimal) (map[uuid.UUID]decimal.Decimal, error) {
	strategyAllocations := make(map[uuid.UUID]decimal.Decimal)
	
	// Separate percentage and budget strategies
	var percentStrategies []*models.Strategy
	var budgetStrategies []*models.Strategy
	totalBudget := decimal.Zero
	
	for _, strategy := range strategies {
		if strategy.WeightMode == models.WeightModePercent {
			percentStrategies = append(percentStrategies, strategy)
		} else {
			budgetStrategies = append(budgetStrategies, strategy)
			totalBudget = totalBudget.Add(strategy.WeightValue)
		}
	}
	
	// Validate that budget strategies don't exceed total investment
	if totalBudget.GreaterThan(totalInvestment) {
		return nil, ErrBudgetExceedsInvestment(totalBudget, totalInvestment)
	}
	
	// Calculate remaining amount for percentage strategies
	remainingForPercent := totalInvestment.Sub(totalBudget)
	
	// Validate percentage strategies don't exceed 100%
	totalPercentage := decimal.Zero
	for _, strategy := range percentStrategies {
		totalPercentage = totalPercentage.Add(strategy.WeightValue)
	}
	
	if totalPercentage.GreaterThan(decimal.NewFromInt(100)) {
		return nil, ErrInvalidStrategyWeights(totalPercentage)
	}
	
	// Allocate budget strategies
	for _, strategy := range budgetStrategies {
		strategyAllocations[strategy.ID] = strategy.WeightValue
	}
	
	// Allocate percentage strategies
	for _, strategy := range percentStrategies {
		allocation := remainingForPercent.Mul(strategy.WeightValue).Div(decimal.NewFromInt(100))
		strategyAllocations[strategy.ID] = allocation
	}
	
	return strategyAllocations, nil
}

// distributeToStocks distributes strategy allocations to eligible stocks with "Buy" signals
func (e *AllocationEngine) distributeToStocks(ctx context.Context, strategies []*models.Strategy, strategyAllocations map[uuid.UUID]decimal.Decimal, excludedStocks []uuid.UUID) ([]models.StockAllocation, error) {
	stockAllocations := make(map[uuid.UUID]*models.StockAllocation)
	excludedSet := make(map[uuid.UUID]bool)
	
	// Create excluded stocks set for quick lookup
	for _, stockID := range excludedStocks {
		excludedSet[stockID] = true
	}
	
	// Get all unique stock IDs from all strategies
	allStockIDs := make(map[uuid.UUID]bool)
	strategyStocks := make(map[uuid.UUID][]*models.StrategyStock)
	
	for _, strategy := range strategies {
		stocks, err := e.strategyRepo.GetStrategyStocks(ctx, strategy.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get stocks for strategy %s: %w", strategy.ID, err)
		}
		
		strategyStocks[strategy.ID] = stocks
		for _, stock := range stocks {
			if stock.Eligible && !excludedSet[stock.StockID] {
				allStockIDs[stock.StockID] = true
			}
		}
	}
	
	// Convert to slice
	stockIDSlice := make([]uuid.UUID, 0, len(allStockIDs))
	for stockID := range allStockIDs {
		stockIDSlice = append(stockIDSlice, stockID)
	}
	
	// Get stock details
	stocks, err := e.stockRepo.GetByIDs(ctx, stockIDSlice)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock details: %w", err)
	}
	
	stockMap := make(map[uuid.UUID]*models.Stock)
	for _, stock := range stocks {
		stockMap[stock.ID] = stock
	}
	
	// Get latest signals for all stocks
	signals, err := e.signalRepo.GetLatestSignals(ctx, stockIDSlice)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock signals: %w", err)
	}
	
	// Process each strategy
	for _, strategy := range strategies {
		strategyAllocation := strategyAllocations[strategy.ID]
		if strategyAllocation.IsZero() {
			continue
		}
		
		// Get eligible stocks with "Buy" signals for this strategy
		eligibleStocks := make([]uuid.UUID, 0)
		for _, strategyStock := range strategyStocks[strategy.ID] {
			if !strategyStock.Eligible || excludedSet[strategyStock.StockID] {
				continue
			}
			
			// Check if stock has "Buy" signal
			signal, hasSignal := signals[strategyStock.StockID]
			if !hasSignal || signal.Signal != models.SignalBuy {
				continue
			}
			
			eligibleStocks = append(eligibleStocks, strategyStock.StockID)
		}
		
		if len(eligibleStocks) == 0 {
			continue
		}
		
		// Distribute strategy allocation equally among eligible stocks
		allocationPerStock := strategyAllocation.Div(decimal.NewFromInt(int64(len(eligibleStocks))))
		
		for _, stockID := range eligibleStocks {
			stock := stockMap[stockID]
			if stock == nil {
				continue
			}
			
			// Initialize or update stock allocation
			if existing, exists := stockAllocations[stockID]; exists {
				existing.AllocationValue = existing.AllocationValue.Add(allocationPerStock)
				if existing.StrategyContrib == nil {
					existing.StrategyContrib = make(map[string]decimal.Decimal)
				}
				existing.StrategyContrib[strategy.ID.String()] = allocationPerStock
			} else {
				stockAllocations[stockID] = &models.StockAllocation{
					StockID:         stockID,
					Ticker:          stock.Ticker,
					Name:            stock.Name,
					AllocationValue: allocationPerStock,
					StrategyContrib: map[string]decimal.Decimal{
						strategy.ID.String(): allocationPerStock,
					},
				}
			}
		}
	}
	
	// Convert map to slice
	result := make([]models.StockAllocation, 0, len(stockAllocations))
	for _, allocation := range stockAllocations {
		result = append(result, *allocation)
	}
	
	return result, nil
}

// applyConstraints applies min/max allocation constraints
func (e *AllocationEngine) applyConstraints(allocations []models.StockAllocation, constraints models.AllocationConstraints, totalInvestment decimal.Decimal) ([]models.StockAllocation, error) {
	result := make([]models.StockAllocation, 0, len(allocations))
	
	maxAllocationAmount := totalInvestment.Mul(constraints.MaxAllocationPerStock).Div(decimal.NewFromInt(100))
	
	for _, allocation := range allocations {
		// Apply maximum allocation constraint
		if allocation.AllocationValue.GreaterThan(maxAllocationAmount) {
			allocation.AllocationValue = maxAllocationAmount
		}
		
		// Apply minimum allocation constraint
		if allocation.AllocationValue.LessThan(constraints.MinAllocationAmount) {
			// Skip stocks that don't meet minimum allocation
			continue
		}
		
		result = append(result, allocation)
	}
	
	return result, nil
}

// normalizeAllocations normalizes allocations to match total investment amount
func (e *AllocationEngine) normalizeAllocations(allocations []models.StockAllocation, totalInvestment decimal.Decimal) []models.StockAllocation {
	if len(allocations) == 0 {
		return allocations
	}
	
	// Calculate current total
	currentTotal := decimal.Zero
	for _, allocation := range allocations {
		currentTotal = currentTotal.Add(allocation.AllocationValue)
	}
	
	if currentTotal.IsZero() {
		return allocations
	}
	
	// Calculate normalization factor
	normalizationFactor := totalInvestment.Div(currentTotal)
	
	// Apply normalization
	for i := range allocations {
		allocations[i].AllocationValue = allocations[i].AllocationValue.Mul(normalizationFactor)
		allocations[i].Weight = allocations[i].AllocationValue.Div(totalInvestment).Mul(decimal.NewFromInt(100))
		
		// Normalize strategy contributions
		for strategyID, contrib := range allocations[i].StrategyContrib {
			allocations[i].StrategyContrib[strategyID] = contrib.Mul(normalizationFactor)
		}
	}
	
	return allocations
}

// RebalanceAllocations recalculates allocations for an existing portfolio with new investment amount
func (e *AllocationEngine) RebalanceAllocations(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.AllocationPreview, error) {
	// This would typically:
	// 1. Get the existing portfolio and its positions
	// 2. Extract the original strategy configuration
	// 3. Recalculate allocations with the new total investment
	// 4. Return the new allocation preview
	
	// For now, return an error indicating this needs portfolio repository integration
	return nil, fmt.Errorf("rebalancing requires portfolio repository integration - to be implemented in task 8")
}

// ValidateConstraints validates that allocations meet the specified constraints (simple error return)
func (e *AllocationEngine) ValidateConstraints(allocations []models.StockAllocation, constraints models.AllocationConstraints) error {
	for _, allocation := range allocations {
		// Check minimum allocation
		if allocation.AllocationValue.LessThan(constraints.MinAllocationAmount) {
			return fmt.Errorf("stock %s allocation (%s) is below minimum (%s)", 
				allocation.Ticker, 
				allocation.AllocationValue.String(), 
				constraints.MinAllocationAmount.String())
		}
		
		// Check maximum allocation percentage
		if allocation.Weight.GreaterThan(constraints.MaxAllocationPerStock) {
			return fmt.Errorf("stock %s allocation (%s%%) exceeds maximum (%s%%)", 
				allocation.Ticker, 
				allocation.Weight.String(), 
				constraints.MaxAllocationPerStock.String())
		}
	}
	
	return nil
}

// ValidateConstraintsDetailed validates allocations with detailed feedback and suggestions
func (e *AllocationEngine) ValidateConstraintsDetailed(allocations []models.StockAllocation, constraints models.AllocationConstraints, totalInvestment decimal.Decimal) *ValidationResult {
	return e.constraintValidator.ValidateAllocations(allocations, constraints, totalInvestment)
}

// ValidateConstraintsConfig validates the constraint configuration itself
func (e *AllocationEngine) ValidateConstraintsConfig(constraints models.AllocationConstraints, totalInvestment decimal.Decimal) *ValidationResult {
	return e.constraintValidator.ValidateConstraintsConfig(constraints, totalInvestment)
}

// addPricesAndQuantities fetches real-time prices and calculates quantities for allocations
func (e *AllocationEngine) addPricesAndQuantities(ctx context.Context, allocations []models.StockAllocation) ([]models.StockAllocation, error) {
	if len(allocations) == 0 {
		return allocations, nil
	}

	// Extract symbols for price fetching
	symbols := make([]string, len(allocations))
	for i, allocation := range allocations {
		symbols[i] = allocation.Ticker
	}

	// Fetch quotes for all symbols
	quotes, err := e.marketDataService.GetMultipleQuotes(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market quotes: %w", err)
	}

	// Update allocations with prices and calculate quantities
	result := make([]models.StockAllocation, len(allocations))
	for i, allocation := range allocations {
		result[i] = allocation // Copy the allocation

		quote, hasQuote := quotes[allocation.Ticker]
		if !hasQuote {
			return nil, fmt.Errorf("no quote available for symbol %s", allocation.Ticker)
		}

		// Set the price
		result[i].Price = quote.Price

		// Calculate quantity using floor logic (can only buy whole shares)
		if quote.Price.GreaterThan(decimal.Zero) {
			quantity := allocation.AllocationValue.Div(quote.Price).IntPart()
			result[i].Quantity = int(quantity)

			// Calculate actual value based on whole shares
			result[i].ActualValue = quote.Price.Mul(decimal.NewFromInt(quantity))
		} else {
			result[i].Quantity = 0
			result[i].ActualValue = decimal.Zero
		}
	}

	return result, nil
}

// RecalculateWithExclusions recalculates allocations when stocks are removed
func (e *AllocationEngine) RecalculateWithExclusions(ctx context.Context, originalReq *models.AllocationRequest, excludedStocks []uuid.UUID) (*models.AllocationPreview, error) {
	// Create a new request with excluded stocks
	newReq := *originalReq
	
	// Merge excluded stocks
	excludedSet := make(map[uuid.UUID]bool)
	for _, stockID := range originalReq.ExcludedStocks {
		excludedSet[stockID] = true
	}
	for _, stockID := range excludedStocks {
		excludedSet[stockID] = true
	}
	
	// Convert back to slice
	newExcluded := make([]uuid.UUID, 0, len(excludedSet))
	for stockID := range excludedSet {
		newExcluded = append(newExcluded, stockID)
	}
	newReq.ExcludedStocks = newExcluded

	// Recalculate allocations
	return e.CalculateAllocations(ctx, &newReq)
}