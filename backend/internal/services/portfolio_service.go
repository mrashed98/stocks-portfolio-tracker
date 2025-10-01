package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"portfolio-app/internal/models"
)

// PortfolioRepository interface for portfolio data access
type PortfolioRepository interface {
	Create(ctx context.Context, portfolio *models.Portfolio) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Portfolio, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Portfolio, error)
	GetAllPortfolioIDs(ctx context.Context) ([]uuid.UUID, error)
	Update(ctx context.Context, portfolio *models.Portfolio) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	CreatePosition(ctx context.Context, position *models.Position) error
	GetPositions(ctx context.Context, portfolioID uuid.UUID) ([]*models.Position, error)
	UpdatePosition(ctx context.Context, position *models.Position) error
	DeletePosition(ctx context.Context, portfolioID, stockID uuid.UUID) error
	
	CreateNAVHistory(ctx context.Context, navHistory *models.NAVHistory) error
	GetNAVHistory(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]*models.NAVHistory, error)
	GetLatestNAV(ctx context.Context, portfolioID uuid.UUID) (*models.NAVHistory, error)
	
	CreatePortfolioWithPositions(ctx context.Context, portfolio *models.Portfolio, positions []*models.Position) error
}

// PortfolioService handles portfolio-related operations
type PortfolioService struct {
	allocationEngine AllocationEngineInterface
	strategyRepo     StrategyRepository
	portfolioRepo    PortfolioRepository
	marketDataService MarketDataService
}

// PortfolioServiceInterface defines the portfolio service contract
type PortfolioServiceInterface interface {
	// Allocation preview operations
	GenerateAllocationPreview(ctx context.Context, req *models.AllocationRequest) (*models.AllocationPreview, error)
	GenerateAllocationPreviewWithExclusions(ctx context.Context, req *models.AllocationRequest, excludedStocks []uuid.UUID) (*models.AllocationPreview, error)
	ValidateAllocationRequest(req *models.AllocationRequest) error
	
	// Portfolio CRUD operations
	CreatePortfolio(ctx context.Context, req *models.CreatePortfolioRequest, userID uuid.UUID) (*models.Portfolio, error)
	GetPortfolio(ctx context.Context, id uuid.UUID) (*models.Portfolio, error)
	GetUserPortfolios(ctx context.Context, userID uuid.UUID) ([]*models.Portfolio, error)
	UpdatePortfolio(ctx context.Context, id uuid.UUID, req *models.UpdatePortfolioRequest) (*models.Portfolio, error)
	DeletePortfolio(ctx context.Context, id uuid.UUID) error
	
	// Portfolio performance operations
	UpdatePortfolioNAV(ctx context.Context, portfolioID uuid.UUID) (*models.NAVHistory, error)
	GetPortfolioHistory(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]*models.NAVHistory, error)
	GetPortfolioPerformanceMetrics(ctx context.Context, portfolioID uuid.UUID) (*models.PerformanceMetrics, error)
	
	// Portfolio rebalancing operations
	GenerateRebalancePreview(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.AllocationPreview, error)
	RebalancePortfolio(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.Portfolio, error)
}

// NewPortfolioService creates a new portfolio service
func NewPortfolioService(
	allocationEngine AllocationEngineInterface,
	strategyRepo StrategyRepository,
	portfolioRepo PortfolioRepository,
	marketDataService MarketDataService,
) *PortfolioService {
	return &PortfolioService{
		allocationEngine:  allocationEngine,
		strategyRepo:      strategyRepo,
		portfolioRepo:     portfolioRepo,
		marketDataService: marketDataService,
	}
}

// GenerateAllocationPreview generates an allocation preview based on strategies and constraints
func (s *PortfolioService) GenerateAllocationPreview(ctx context.Context, req *models.AllocationRequest) (*models.AllocationPreview, error) {
	// Validate the request
	if err := s.ValidateAllocationRequest(req); err != nil {
		return nil, fmt.Errorf("invalid allocation request: %w", err)
	}

	// Validate that all strategies exist and belong to the user (if user context is available)
	strategies, err := s.strategyRepo.GetByIDs(ctx, req.StrategyIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to validate strategies: %w", err)
	}

	if len(strategies) != len(req.StrategyIDs) {
		return nil, fmt.Errorf("some strategies not found or not accessible")
	}

	// Generate allocation preview
	preview, err := s.allocationEngine.CalculateAllocations(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate allocations: %w", err)
	}

	return preview, nil
}

// GenerateAllocationPreviewWithExclusions generates an allocation preview with specific stocks excluded
func (s *PortfolioService) GenerateAllocationPreviewWithExclusions(ctx context.Context, req *models.AllocationRequest, excludedStocks []uuid.UUID) (*models.AllocationPreview, error) {
	// Validate the request
	if err := s.ValidateAllocationRequest(req); err != nil {
		return nil, fmt.Errorf("invalid allocation request: %w", err)
	}

	// Use the allocation engine's recalculation method
	preview, err := s.allocationEngine.RecalculateWithExclusions(ctx, req, excludedStocks)
	if err != nil {
		return nil, fmt.Errorf("failed to recalculate allocations with exclusions: %w", err)
	}

	return preview, nil
}

// ValidateAllocationRequest validates an allocation request
func (s *PortfolioService) ValidateAllocationRequest(req *models.AllocationRequest) error {
	if req == nil {
		return fmt.Errorf("allocation request cannot be nil")
	}

	if len(req.StrategyIDs) == 0 {
		return fmt.Errorf("at least one strategy must be specified")
	}

	if req.TotalInvestment.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("total investment must be greater than zero")
	}

	// Validate constraints
	if req.Constraints.MaxAllocationPerStock.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("maximum allocation per stock must be greater than zero")
	}

	if req.Constraints.MaxAllocationPerStock.GreaterThan(decimal.NewFromInt(100)) {
		return fmt.Errorf("maximum allocation per stock cannot exceed 100%%")
	}

	if req.Constraints.MinAllocationAmount.LessThan(decimal.Zero) {
		return fmt.Errorf("minimum allocation amount cannot be negative")
	}

	return nil
}

// CreatePortfolio creates a new portfolio with positions based on allocation preview
func (s *PortfolioService) CreatePortfolio(ctx context.Context, req *models.CreatePortfolioRequest, userID uuid.UUID) (*models.Portfolio, error) {
	// Validate the request
	if req == nil {
		return nil, fmt.Errorf("create portfolio request cannot be nil")
	}
	
	if req.Name == "" {
		return nil, fmt.Errorf("portfolio name is required")
	}
	
	if req.TotalInvestment.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("total investment must be greater than zero")
	}
	
	if len(req.Positions) == 0 {
		return nil, fmt.Errorf("at least one position is required")
	}
	
	// Create portfolio entity
	portfolio := &models.Portfolio{}
	portfolio.FromCreateRequest(req, userID)
	
	// Create position entities
	positions := make([]*models.Position, len(req.Positions))
	for i, posReq := range req.Positions {
		position := &models.Position{}
		if err := position.FromCreateRequest(&posReq, portfolio.ID); err != nil {
			return nil, fmt.Errorf("failed to create position %d: %w", i, err)
		}
		positions[i] = position
	}
	
	// Create portfolio with positions in transaction
	if err := s.portfolioRepo.CreatePortfolioWithPositions(ctx, portfolio, positions); err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}
	
	// Load the created portfolio with all related data
	createdPortfolio, err := s.portfolioRepo.GetByID(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created portfolio: %w", err)
	}
	
	return createdPortfolio, nil
}

// GetPortfolio retrieves a portfolio by ID with current market data
func (s *PortfolioService) GetPortfolio(ctx context.Context, id uuid.UUID) (*models.Portfolio, error) {
	portfolio, err := s.portfolioRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	
	// Enrich positions with current market data
	if err := s.enrichPositionsWithMarketData(ctx, portfolio.Positions); err != nil {
		// Log error but don't fail - return portfolio with stale data
		fmt.Printf("Warning: failed to enrich positions with market data: %v\n", err)
	}
	
	return portfolio, nil
}

// GetUserPortfolios retrieves all portfolios for a user
func (s *PortfolioService) GetUserPortfolios(ctx context.Context, userID uuid.UUID) ([]*models.Portfolio, error) {
	portfolios, err := s.portfolioRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user portfolios: %w", err)
	}
	
	// Enrich all portfolios with current market data
	for _, portfolio := range portfolios {
		if err := s.enrichPositionsWithMarketData(ctx, portfolio.Positions); err != nil {
			// Log error but continue with other portfolios
			fmt.Printf("Warning: failed to enrich portfolio %s with market data: %v\n", portfolio.ID, err)
		}
	}
	
	return portfolios, nil
}

// UpdatePortfolio updates a portfolio
func (s *PortfolioService) UpdatePortfolio(ctx context.Context, id uuid.UUID, req *models.UpdatePortfolioRequest) (*models.Portfolio, error) {
	// Get existing portfolio
	portfolio, err := s.portfolioRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	
	// Apply updates
	portfolio.ApplyUpdate(req)
	
	// Update in database
	if err := s.portfolioRepo.Update(ctx, portfolio); err != nil {
		return nil, fmt.Errorf("failed to update portfolio: %w", err)
	}
	
	// Return updated portfolio
	return s.GetPortfolio(ctx, id)
}

// DeletePortfolio deletes a portfolio
func (s *PortfolioService) DeletePortfolio(ctx context.Context, id uuid.UUID) error {
	if err := s.portfolioRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete portfolio: %w", err)
	}
	
	return nil
}

// UpdatePortfolioNAV calculates and updates the current NAV for a portfolio
func (s *PortfolioService) UpdatePortfolioNAV(ctx context.Context, portfolioID uuid.UUID) (*models.NAVHistory, error) {
	// Get portfolio with positions
	portfolio, err := s.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	
	if len(portfolio.Positions) == 0 {
		// Portfolio has no positions, NAV equals cash (total investment)
		navHistory := &models.NAVHistory{
			PortfolioID: portfolioID,
			Timestamp:   time.Now(),
			NAV:         portfolio.TotalInvestment,
			PnL:         decimal.Zero,
			CreatedAt:   time.Now(),
		}
		
		// Calculate drawdown from high water mark
		if err := s.calculateDrawdown(ctx, navHistory); err != nil {
			return nil, fmt.Errorf("failed to calculate drawdown: %w", err)
		}
		
		if err := s.portfolioRepo.CreateNAVHistory(ctx, navHistory); err != nil {
			return nil, fmt.Errorf("failed to create NAV history: %w", err)
		}
		
		return navHistory, nil
	}
	
	// Get current market prices for all positions
	if err := s.enrichPositionsWithMarketData(ctx, portfolio.Positions); err != nil {
		return nil, fmt.Errorf("failed to get current market prices: %w", err)
	}
	
	// Calculate current NAV
	currentNAV := decimal.Zero
	totalPnL := decimal.Zero
	
	for _, position := range portfolio.Positions {
		if position.CurrentValue != nil {
			currentNAV = currentNAV.Add(*position.CurrentValue)
		} else {
			// Fallback to entry value if no current price available
			currentNAV = currentNAV.Add(position.AllocationValue)
		}
		
		if position.PnL != nil {
			totalPnL = totalPnL.Add(*position.PnL)
		}
	}
	
	// Create NAV history entry
	navHistory := &models.NAVHistory{
		PortfolioID: portfolioID,
		Timestamp:   time.Now(),
		NAV:         currentNAV,
		PnL:         totalPnL,
		CreatedAt:   time.Now(),
	}
	
	// Calculate drawdown from high water mark
	if err := s.calculateDrawdown(ctx, navHistory); err != nil {
		return nil, fmt.Errorf("failed to calculate drawdown: %w", err)
	}
	
	// Save NAV history
	if err := s.portfolioRepo.CreateNAVHistory(ctx, navHistory); err != nil {
		return nil, fmt.Errorf("failed to create NAV history: %w", err)
	}
	
	return navHistory, nil
}

// GetPortfolioHistory retrieves NAV history for a portfolio
func (s *PortfolioService) GetPortfolioHistory(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]*models.NAVHistory, error) {
	history, err := s.portfolioRepo.GetNAVHistory(ctx, portfolioID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio history: %w", err)
	}
	
	return history, nil
}

// GetPortfolioPerformanceMetrics calculates performance metrics for a portfolio
func (s *PortfolioService) GetPortfolioPerformanceMetrics(ctx context.Context, portfolioID uuid.UUID) (*models.PerformanceMetrics, error) {
	// Get portfolio for initial investment
	portfolio, err := s.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	
	// Get all NAV history
	history, err := s.portfolioRepo.GetNAVHistory(ctx, portfolioID, time.Time{}, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get NAV history: %w", err)
	}
	
	if len(history) == 0 {
		return &models.PerformanceMetrics{}, nil
	}
	
	// Convert to slice of values for calculation
	historyValues := make([]models.NAVHistory, len(history))
	for i, h := range history {
		historyValues[i] = *h
	}
	
	// Calculate performance metrics
	metrics := models.CalculatePerformanceMetrics(historyValues, portfolio.TotalInvestment)
	
	return metrics, nil
}

// GenerateRebalancePreview generates a preview of how a portfolio would be rebalanced
func (s *PortfolioService) GenerateRebalancePreview(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.AllocationPreview, error) {
	// Get portfolio to extract original strategy configuration
	portfolio, err := s.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	
	if len(portfolio.Positions) == 0 {
		return nil, fmt.Errorf("portfolio has no positions to rebalance")
	}
	
	// Extract strategy IDs from positions (this is a simplified approach)
	// In a real implementation, you might want to store the original allocation request
	strategyIDs := make(map[uuid.UUID]bool)
	for _, position := range portfolio.Positions {
		if position.StrategyContribMap != nil {
			for strategyIDStr := range position.StrategyContribMap {
				if strategyID, err := uuid.Parse(strategyIDStr); err == nil {
					strategyIDs[strategyID] = true
				}
			}
		}
	}
	
	// Convert to slice
	strategyIDSlice := make([]uuid.UUID, 0, len(strategyIDs))
	for strategyID := range strategyIDs {
		strategyIDSlice = append(strategyIDSlice, strategyID)
	}
	
	if len(strategyIDSlice) == 0 {
		return nil, fmt.Errorf("could not determine original strategies for rebalancing")
	}
	
	// Create allocation request with new total investment
	allocationReq := &models.AllocationRequest{
		StrategyIDs:     strategyIDSlice,
		TotalInvestment: newTotalInvestment,
		Constraints: models.AllocationConstraints{
			MaxAllocationPerStock: decimal.NewFromInt(20), // Default 20% max per stock
			MinAllocationAmount:   decimal.NewFromInt(100), // Default $100 minimum
		},
	}
	
	// Generate new allocation preview
	preview, err := s.allocationEngine.CalculateAllocations(ctx, allocationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate rebalance allocations: %w", err)
	}
	
	return preview, nil
}

// RebalancePortfolio rebalances a portfolio to match new allocation targets
func (s *PortfolioService) RebalancePortfolio(ctx context.Context, portfolioID uuid.UUID, newTotalInvestment decimal.Decimal) (*models.Portfolio, error) {
	// Generate rebalance preview
	preview, err := s.GenerateRebalancePreview(ctx, portfolioID, newTotalInvestment)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rebalance preview: %w", err)
	}
	
	// Get current portfolio
	portfolio, err := s.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}
	
	// Update portfolio total investment
	portfolio.TotalInvestment = newTotalInvestment
	portfolio.UpdatedAt = time.Now()
	
	if err := s.portfolioRepo.Update(ctx, portfolio); err != nil {
		return nil, fmt.Errorf("failed to update portfolio: %w", err)
	}
	
	// Update positions based on new allocations
	for _, allocation := range preview.Allocations {
		// Find existing position
		var existingPosition *models.Position
		for i := range portfolio.Positions {
			if portfolio.Positions[i].StockID == allocation.StockID {
				existingPosition = &portfolio.Positions[i]
				break
			}
		}
		
		if existingPosition != nil {
			// Update existing position
			existingPosition.Quantity = allocation.Quantity
			existingPosition.AllocationValue = allocation.ActualValue
			existingPosition.UpdatedAt = time.Now()
			
			if err := s.portfolioRepo.UpdatePosition(ctx, existingPosition); err != nil {
				return nil, fmt.Errorf("failed to update position for stock %s: %w", allocation.Ticker, err)
			}
		} else {
			// Create new position
			newPosition := &models.Position{
				PortfolioID:     portfolioID,
				StockID:         allocation.StockID,
				Quantity:        allocation.Quantity,
				EntryPrice:      allocation.Price,
				AllocationValue: allocation.ActualValue,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			
			// Convert strategy contribution to JSON
			if allocation.StrategyContrib != nil {
				contribJSON, err := json.Marshal(allocation.StrategyContrib)
				if err == nil {
					newPosition.StrategyContrib = contribJSON
					newPosition.StrategyContribMap = allocation.StrategyContrib
				}
			}
			
			if err := s.portfolioRepo.CreatePosition(ctx, newPosition); err != nil {
				return nil, fmt.Errorf("failed to create position for stock %s: %w", allocation.Ticker, err)
			}
		}
	}
	
	// Update NAV after rebalancing
	if _, err := s.UpdatePortfolioNAV(ctx, portfolioID); err != nil {
		// Log error but don't fail the rebalancing
		fmt.Printf("Warning: failed to update NAV after rebalancing: %v\n", err)
	}
	
	// Return updated portfolio
	return s.GetPortfolio(ctx, portfolioID)
}

// enrichPositionsWithMarketData fetches current market prices and calculates position metrics
func (s *PortfolioService) enrichPositionsWithMarketData(ctx context.Context, positions []models.Position) error {
	if len(positions) == 0 {
		return nil
	}
	
	// Extract unique tickers
	tickers := make([]string, 0, len(positions))
	tickerSet := make(map[string]bool)
	
	for _, position := range positions {
		if position.Stock != nil && !tickerSet[position.Stock.Ticker] {
			tickers = append(tickers, position.Stock.Ticker)
			tickerSet[position.Stock.Ticker] = true
		}
	}
	
	if len(tickers) == 0 {
		return nil
	}
	
	// Fetch current quotes
	quotes, err := s.marketDataService.GetMultipleQuotes(ctx, tickers)
	if err != nil {
		return fmt.Errorf("failed to fetch market quotes: %w", err)
	}
	
	// Update positions with current market data
	for i := range positions {
		if positions[i].Stock == nil {
			continue
		}
		
		quote, hasQuote := quotes[positions[i].Stock.Ticker]
		if hasQuote {
			positions[i].CalculateMetrics(quote.Price)
		}
	}
	
	return nil
}

// calculateDrawdown calculates drawdown from high water mark
func (s *PortfolioService) calculateDrawdown(ctx context.Context, navHistory *models.NAVHistory) error {
	// Get all previous NAV history to find high water mark
	allHistory, err := s.portfolioRepo.GetNAVHistory(ctx, navHistory.PortfolioID, time.Time{}, navHistory.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to get NAV history for drawdown calculation: %w", err)
	}
	
	// Find high water mark
	highWaterMark := navHistory.NAV
	for _, entry := range allHistory {
		if entry.NAV.GreaterThan(highWaterMark) {
			highWaterMark = entry.NAV
		}
	}
	
	// Calculate drawdown
	navHistory.CalculateDrawdown(highWaterMark)
	
	return nil
}

// AllocationPreviewCache provides caching for allocation previews
type AllocationPreviewCache struct {
	cache map[string]*CachedPreview
}

// CachedPreview represents a cached allocation preview
type CachedPreview struct {
	Preview   *models.AllocationPreview
	Timestamp time.Time
	TTL       time.Duration
}

// NewAllocationPreviewCache creates a new allocation preview cache
func NewAllocationPreviewCache() *AllocationPreviewCache {
	return &AllocationPreviewCache{
		cache: make(map[string]*CachedPreview),
	}
}

// Get retrieves a cached preview if it exists and is not expired
func (c *AllocationPreviewCache) Get(key string) (*models.AllocationPreview, bool) {
	cached, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(cached.Timestamp) > cached.TTL {
		delete(c.cache, key)
		return nil, false
	}

	return cached.Preview, true
}

// Set stores a preview in the cache
func (c *AllocationPreviewCache) Set(key string, preview *models.AllocationPreview, ttl time.Duration) {
	c.cache[key] = &CachedPreview{
		Preview:   preview,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// Clear removes all cached previews
func (c *AllocationPreviewCache) Clear() {
	c.cache = make(map[string]*CachedPreview)
}

// GenerateCacheKey generates a cache key for an allocation request
func GenerateCacheKey(req *models.AllocationRequest) string {
	// Create a simple hash-like key based on request parameters
	// In a real implementation, you might want to use a proper hash function
	key := fmt.Sprintf("alloc_%s_%s_%s_%s",
		req.TotalInvestment.String(),
		req.Constraints.MaxAllocationPerStock.String(),
		req.Constraints.MinAllocationAmount.String(),
		fmt.Sprintf("%v", req.StrategyIDs))
	
	if len(req.ExcludedStocks) > 0 {
		key += fmt.Sprintf("_excl_%v", req.ExcludedStocks)
	}
	
	return key
}