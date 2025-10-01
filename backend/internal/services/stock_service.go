package services

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"portfolio-app/internal/models"
	"portfolio-app/internal/repositories"
)

// StockService defines the interface for stock operations
type StockService interface {
	CreateStock(ctx context.Context, req *models.CreateStockRequest) (*models.Stock, error)
	UpdateStock(ctx context.Context, id uuid.UUID, req *models.UpdateStockRequest) (*models.Stock, error)
	GetStock(ctx context.Context, id uuid.UUID) (*models.Stock, error)
	GetStockByTicker(ctx context.Context, ticker string) (*models.Stock, error)
	GetStocks(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error)
	GetStocksWithSignals(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error)
	DeleteStock(ctx context.Context, id uuid.UUID) error
	UpdateStockSignal(ctx context.Context, stockID uuid.UUID, signal models.SignalType) (*models.Signal, error)
	GetStockSignalHistory(ctx context.Context, stockID uuid.UUID, from, to time.Time) ([]*models.Signal, error)
	ValidateTickerSymbol(ticker string) error
	AddStockToStrategy(ctx context.Context, strategyID, stockID uuid.UUID, userID uuid.UUID) error
	RemoveStockFromStrategy(ctx context.Context, strategyID, stockID uuid.UUID, userID uuid.UUID) error
}

// stockService implements the StockService interface
type stockService struct {
	stockRepo    repositories.StockRepository
	signalRepo   repositories.SignalRepository
	strategyRepo repositories.StrategyRepository
	db           *sql.DB
}

// NewStockService creates a new stock service instance
func NewStockService(
	stockRepo repositories.StockRepository,
	signalRepo repositories.SignalRepository,
	strategyRepo repositories.StrategyRepository,
	db *sql.DB,
) StockService {
	return &stockService{
		stockRepo:    stockRepo,
		signalRepo:   signalRepo,
		strategyRepo: strategyRepo,
		db:           db,
	}
}

// CreateStock creates a new stock with ticker validation
func (s *stockService) CreateStock(ctx context.Context, req *models.CreateStockRequest) (*models.Stock, error) {
	// Validate ticker symbol format
	if err := s.ValidateTickerSymbol(req.Ticker); err != nil {
		return nil, err
	}

	// Check if stock already exists
	existingStock, err := s.stockRepo.GetByTicker(ctx, req.Ticker)
	if err != nil && !isNotFoundError(err) {
		return nil, fmt.Errorf("failed to check existing stock: %w", err)
	}
	if existingStock != nil {
		return nil, &models.ValidationError{
			Field:   "ticker",
			Message: fmt.Sprintf("Stock with ticker %s already exists", req.Ticker),
		}
	}

	stock := &models.Stock{}
	stock.FromCreateRequest(req)

	// TODO: In a real implementation, fetch basic stock info from external API
	// For now, we'll use the provided name or generate a placeholder
	if req.Name == "" {
		stock.Name = fmt.Sprintf("%s Corporation", req.Ticker)
	}

	createdStock, err := s.stockRepo.Create(ctx, stock)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock: %w", err)
	}

	return createdStock, nil
}

// UpdateStock updates an existing stock
func (s *stockService) UpdateStock(ctx context.Context, id uuid.UUID, req *models.UpdateStockRequest) (*models.Stock, error) {
	// Get existing stock
	existingStock, err := s.stockRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	// Apply updates
	existingStock.ApplyUpdate(req)

	updatedStock, err := s.stockRepo.Update(ctx, existingStock)
	if err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	return updatedStock, nil
}

// GetStock retrieves a stock by ID with current signal
func (s *stockService) GetStock(ctx context.Context, id uuid.UUID) (*models.Stock, error) {
	stock, err := s.stockRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	// Load current signal
	currentSignal, err := s.signalRepo.GetCurrentSignal(ctx, id)
	if err != nil && !isNotFoundError(err) {
		return nil, fmt.Errorf("failed to get current signal: %w", err)
	}
	if currentSignal != nil {
		stock.CurrentSignal = currentSignal
	}

	return stock, nil
}

// GetStockByTicker retrieves a stock by ticker symbol with current signal
func (s *stockService) GetStockByTicker(ctx context.Context, ticker string) (*models.Stock, error) {
	stock, err := s.stockRepo.GetByTicker(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock by ticker: %w", err)
	}

	// Load current signal
	currentSignal, err := s.signalRepo.GetCurrentSignal(ctx, stock.ID)
	if err != nil && !isNotFoundError(err) {
		return nil, fmt.Errorf("failed to get current signal: %w", err)
	}
	if currentSignal != nil {
		stock.CurrentSignal = currentSignal
	}

	return stock, nil
}

// GetStocks retrieves stocks with optional search and pagination
func (s *stockService) GetStocks(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error) {
	// Set default pagination limits
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	stocks, err := s.stockRepo.GetAll(ctx, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get stocks: %w", err)
	}

	return stocks, nil
}

// GetStocksWithSignals retrieves stocks with their current signals
func (s *stockService) GetStocksWithSignals(ctx context.Context, search string, limit, offset int) ([]*models.Stock, error) {
	// Set default pagination limits
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	stocks, err := s.stockRepo.GetStocksWithCurrentSignals(ctx, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get stocks with signals: %w", err)
	}

	return stocks, nil
}

// DeleteStock deletes a stock
func (s *stockService) DeleteStock(ctx context.Context, id uuid.UUID) error {
	if err := s.stockRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete stock: %w", err)
	}

	return nil
}

// UpdateStockSignal updates the signal for a stock
func (s *stockService) UpdateStockSignal(ctx context.Context, stockID uuid.UUID, signal models.SignalType) (*models.Signal, error) {
	// Verify stock exists
	_, err := s.stockRepo.GetByID(ctx, stockID)
	if err != nil {
		return nil, fmt.Errorf("stock not found: %w", err)
	}

	updatedSignal, err := s.signalRepo.Update(ctx, stockID, signal)
	if err != nil {
		return nil, fmt.Errorf("failed to update stock signal: %w", err)
	}

	return updatedSignal, nil
}

// GetStockSignalHistory retrieves signal history for a stock
func (s *stockService) GetStockSignalHistory(ctx context.Context, stockID uuid.UUID, from, to time.Time) ([]*models.Signal, error) {
	// Verify stock exists
	_, err := s.stockRepo.GetByID(ctx, stockID)
	if err != nil {
		return nil, fmt.Errorf("stock not found: %w", err)
	}

	signals, err := s.signalRepo.GetSignalHistory(ctx, stockID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get signal history: %w", err)
	}

	return signals, nil
}

// ValidateTickerSymbol validates ticker symbol format
func (s *stockService) ValidateTickerSymbol(ticker string) error {
	if ticker == "" {
		return &models.ValidationError{
			Field:   "ticker",
			Message: "Ticker symbol is required",
		}
	}

	// Convert to uppercase for validation
	ticker = strings.ToUpper(ticker)

	// Basic ticker validation: 1-20 characters, letters and numbers only
	tickerRegex := regexp.MustCompile(`^[A-Z0-9]{1,20}$`)
	if !tickerRegex.MatchString(ticker) {
		return &models.ValidationError{
			Field:   "ticker",
			Message: "Ticker symbol must be 1-20 characters, letters and numbers only",
		}
	}

	// TODO: In a real implementation, validate against external market data API
	// For now, we'll accept any valid format

	return nil
}

// AddStockToStrategy adds a stock to a strategy with user authorization
func (s *stockService) AddStockToStrategy(ctx context.Context, strategyID, stockID uuid.UUID, userID uuid.UUID) error {
	// Verify strategy belongs to user
	_, err := s.strategyRepo.GetByID(ctx, strategyID, userID)
	if err != nil {
		return fmt.Errorf("strategy not found or access denied: %w", err)
	}

	// Verify stock exists
	_, err = s.stockRepo.GetByID(ctx, stockID)
	if err != nil {
		return fmt.Errorf("stock not found: %w", err)
	}

	if err := s.strategyRepo.AddStockToStrategy(ctx, strategyID, stockID); err != nil {
		return fmt.Errorf("failed to add stock to strategy: %w", err)
	}

	return nil
}

// RemoveStockFromStrategy removes a stock from a strategy with user authorization
func (s *stockService) RemoveStockFromStrategy(ctx context.Context, strategyID, stockID uuid.UUID, userID uuid.UUID) error {
	// Verify strategy belongs to user
	_, err := s.strategyRepo.GetByID(ctx, strategyID, userID)
	if err != nil {
		return fmt.Errorf("strategy not found or access denied: %w", err)
	}

	if err := s.strategyRepo.RemoveStockFromStrategy(ctx, strategyID, stockID); err != nil {
		return fmt.Errorf("failed to remove stock from strategy: %w", err)
	}

	return nil
}

// isNotFoundError checks if an error is a NotFoundError
func isNotFoundError(err error) bool {
	_, ok := err.(*models.NotFoundError)
	return ok
}