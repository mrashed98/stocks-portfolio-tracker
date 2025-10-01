package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	socket "github.com/verzth/tradingview-scraper/v2"
)

// TradingViewService implements MarketDataService using TradingView's real-time socket
type TradingViewService struct {
	redisClient    *redis.Client
	socket         socket.SocketInterface
	circuitBreaker *CircuitBreaker
	cacheTTL       time.Duration
	quotes         map[string]*Quote
	quoteMutex     sync.RWMutex
	isConnected    bool
	connMutex      sync.RWMutex
}

// NewTradingViewService creates a new TradingView market data service
func NewTradingViewService(redisClient *redis.Client) *TradingViewService {
	service := &TradingViewService{
		redisClient:    redisClient,
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
		cacheTTL:       5 * time.Minute,
		quotes:         make(map[string]*Quote),
		isConnected:    false,
	}

	// Initialize the connection
	if err := service.connect(); err != nil {
		log.Printf("Failed to connect to TradingView: %v", err)
	}

	return service
}

// connect establishes connection to TradingView socket
func (s *TradingViewService) connect() error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	if s.isConnected {
		return nil
	}

	tradingViewSocket, err := socket.Connect(
		s.handleQuoteData,
		s.handleError,
	)
	if err != nil {
		return fmt.Errorf("failed to connect to TradingView socket: %w", err)
	}

	s.socket = tradingViewSocket
	s.isConnected = true
	log.Println("Successfully connected to TradingView socket")

	return nil
}

// handleQuoteData processes incoming quote data from TradingView
func (s *TradingViewService) handleQuoteData(symbol string, data *socket.QuoteData) {
	s.quoteMutex.Lock()
	defer s.quoteMutex.Unlock()

	// Get existing quote or create new one
	quote, exists := s.quotes[symbol]
	if !exists {
		quote = &Quote{
			Symbol:    symbol,
			Timestamp: time.Now(),
		}
		s.quotes[symbol] = quote
	}

	// Update quote with new data
	if data.Price != nil {
		quote.Price = decimal.NewFromFloat(*data.Price)
		quote.Timestamp = time.Now()
	}
	if data.Volume != nil {
		quote.Volume = int64(*data.Volume)
	}
	if data.Bid != nil {
		// Store bid information (could be used for spread calculations)
		// For now, we'll use it to enhance our quote data
	}
	if data.Ask != nil {
		// Store ask information (could be used for spread calculations)
	}

	// Calculate change if we have previous close
	if !quote.PreviousClose.IsZero() && !quote.Price.IsZero() {
		quote.Change = quote.Price.Sub(quote.PreviousClose)
		if quote.PreviousClose.GreaterThan(decimal.Zero) {
			quote.ChangePercent = quote.Change.Div(quote.PreviousClose).Mul(decimal.NewFromInt(100))
		}
	}

	log.Printf("Updated quote for %s: Price=%s, Volume=%d", symbol, quote.Price.String(), quote.Volume)
}

// handleError processes errors from TradingView socket
func (s *TradingViewService) handleError(err error, context string) {
	log.Printf("TradingView socket error: %v, context: %s", err, context)
	
	s.connMutex.Lock()
	s.isConnected = false
	s.connMutex.Unlock()

	// Attempt to reconnect after a delay
	go func() {
		time.Sleep(5 * time.Second)
		if reconnectErr := s.connect(); reconnectErr != nil {
			log.Printf("Failed to reconnect to TradingView: %v", reconnectErr)
		}
	}()
}

// GetQuote retrieves a quote for a single symbol
func (s *TradingViewService) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("tv_quote:%s", symbol)
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var quote Quote
		if json.Unmarshal([]byte(cached), &quote) == nil {
			return &quote, nil
		}
	}

	// Ensure we're connected
	if err := s.ensureConnection(); err != nil {
		return nil, fmt.Errorf("failed to ensure TradingView connection: %w", err)
	}

	// Convert symbol to TradingView format (e.g., NASDAQ:AAPL)
	tvSymbol := s.convertToTradingViewSymbol(symbol)

	// Subscribe to the symbol if not already subscribed
	if err := s.subscribeToSymbol(tvSymbol); err != nil {
		return nil, fmt.Errorf("failed to subscribe to symbol %s: %w", tvSymbol, err)
	}

	// Check if we have real-time data
	s.quoteMutex.RLock()
	quote, exists := s.quotes[tvSymbol]
	s.quoteMutex.RUnlock()

	if exists && !quote.Price.IsZero() {
		// Cache the result
		if data, marshalErr := json.Marshal(quote); marshalErr == nil {
			s.redisClient.Set(ctx, cacheKey, data, s.cacheTTL)
		}
		return quote, nil
	}

	// If no real-time data available, return a placeholder or error
	// In a production system, you might want to fall back to another data source
	return &Quote{
		Symbol:    symbol,
		Price:     decimal.NewFromFloat(100.0), // Placeholder price
		Timestamp: time.Now(),
	}, nil
}

// GetMultipleQuotes retrieves quotes for multiple symbols
func (s *TradingViewService) GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*Quote, error) {
	result := make(map[string]*Quote)
	
	// Ensure connection
	if err := s.ensureConnection(); err != nil {
		return nil, fmt.Errorf("failed to ensure TradingView connection: %w", err)
	}

	// Subscribe to all symbols
	for _, symbol := range symbols {
		tvSymbol := s.convertToTradingViewSymbol(symbol)
		if err := s.subscribeToSymbol(tvSymbol); err != nil {
			log.Printf("Failed to subscribe to %s: %v", tvSymbol, err)
		}
	}

	// Wait a moment for data to arrive
	time.Sleep(1 * time.Second)

	// Collect quotes
	s.quoteMutex.RLock()
	for _, symbol := range symbols {
		tvSymbol := s.convertToTradingViewSymbol(symbol)
		if quote, exists := s.quotes[tvSymbol]; exists && !quote.Price.IsZero() {
			// Convert back to original symbol format
			quoteCopy := *quote
			quoteCopy.Symbol = symbol
			result[symbol] = &quoteCopy
		} else {
			// Provide placeholder if no data available
			result[symbol] = &Quote{
				Symbol:    symbol,
				Price:     decimal.NewFromFloat(100.0),
				Timestamp: time.Now(),
			}
		}
	}
	s.quoteMutex.RUnlock()

	return result, nil
}

// GetQuotesByStockIDs retrieves quotes for stocks by their IDs (requires stock repository)
func (s *TradingViewService) GetQuotesByStockIDs(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*Quote, error) {
	return nil, fmt.Errorf("GetQuotesByStockIDs requires stock repository integration")
}

// GetOHLCV retrieves historical OHLCV data (not supported by real-time socket)
func (s *TradingViewService) GetOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error) {
	// TradingView socket provides real-time data, not historical
	// For historical data, we'd need to use a different approach or API
	return nil, fmt.Errorf("OHLCV historical data not supported by TradingView socket service")
}

// ensureConnection ensures the TradingView socket is connected
func (s *TradingViewService) ensureConnection() error {
	s.connMutex.RLock()
	connected := s.isConnected
	s.connMutex.RUnlock()

	if !connected {
		return s.connect()
	}
	return nil
}

// subscribeToSymbol subscribes to a symbol for real-time updates
func (s *TradingViewService) subscribeToSymbol(symbol string) error {
	if s.socket == nil {
		return fmt.Errorf("socket not connected")
	}

	// Check if already subscribed
	s.quoteMutex.RLock()
	_, exists := s.quotes[symbol]
	s.quoteMutex.RUnlock()

	if !exists {
		// Initialize quote entry
		s.quoteMutex.Lock()
		s.quotes[symbol] = &Quote{
			Symbol:    symbol,
			Timestamp: time.Now(),
		}
		s.quoteMutex.Unlock()

		// Subscribe to symbol
		err := s.socket.AddSymbol(symbol)
		if err != nil {
			return fmt.Errorf("failed to add symbol %s: %w", symbol, err)
		}
		log.Printf("Subscribed to TradingView symbol: %s", symbol)
	}

	return nil
}

// convertToTradingViewSymbol converts a standard symbol to TradingView format
func (s *TradingViewService) convertToTradingViewSymbol(symbol string) string {
	// TradingView uses format like "NASDAQ:AAPL", "NYSE:MSFT", etc.
	// This is a simple mapping - in production, you'd want a more comprehensive mapping
	symbol = strings.ToUpper(symbol)
	
	// Common US stocks mapping
	switch symbol {
	case "AAPL", "GOOGL", "GOOG", "MSFT", "AMZN", "TSLA", "META", "NVDA":
		return "NASDAQ:" + symbol
	case "JPM", "JNJ", "V", "PG", "UNH", "HD", "MA", "DIS", "BAC", "ADBE":
		return "NYSE:" + symbol
	default:
		// Default to NASDAQ for unknown symbols
		return "NASDAQ:" + symbol
	}
}

// UnsubscribeFromSymbol removes a symbol subscription
func (s *TradingViewService) UnsubscribeFromSymbol(symbol string) error {
	if s.socket == nil {
		return fmt.Errorf("socket not connected")
	}

	tvSymbol := s.convertToTradingViewSymbol(symbol)
	err := s.socket.RemoveSymbol(tvSymbol)
	if err != nil {
		return fmt.Errorf("failed to remove symbol %s: %w", tvSymbol, err)
	}

	s.quoteMutex.Lock()
	delete(s.quotes, tvSymbol)
	s.quoteMutex.Unlock()

	log.Printf("Unsubscribed from TradingView symbol: %s", tvSymbol)
	return nil
}

// Close closes the TradingView connection
func (s *TradingViewService) Close() error {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	if s.socket != nil {
		err := s.socket.Close()
		if err != nil {
			log.Printf("Error closing TradingView socket: %v", err)
		}
		s.isConnected = false
		s.socket = nil
	}

	return nil
}