package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

// Quote represents a stock quote with price information
type Quote struct {
	Symbol        string          `json:"symbol"`
	Price         decimal.Decimal `json:"price"`
	Change        decimal.Decimal `json:"change"`
	ChangePercent decimal.Decimal `json:"change_percent"`
	Volume        int64           `json:"volume"`
	High          decimal.Decimal `json:"high"`
	Low           decimal.Decimal `json:"low"`
	Open          decimal.Decimal `json:"open"`
	PreviousClose decimal.Decimal `json:"previous_close"`
	Timestamp     time.Time       `json:"timestamp"`
}

// OHLCV represents historical price data
type OHLCV struct {
	Timestamp time.Time       `json:"timestamp"`
	Open      decimal.Decimal `json:"open"`
	High      decimal.Decimal `json:"high"`
	Low       decimal.Decimal `json:"low"`
	Close     decimal.Decimal `json:"close"`
	Volume    int64           `json:"volume"`
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for external API calls
type CircuitBreaker struct {
	maxFailures   int
	timeout       time.Duration
	state         CircuitState
	failures      int
	lastFailure   time.Time
	nextRetry     time.Time
	mutex         sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       CircuitClosed,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == CircuitOpen {
		if time.Now().Before(cb.nextRetry) {
			return fmt.Errorf("circuit breaker is open, next retry at %v", cb.nextRetry)
		}
		cb.state = CircuitHalfOpen
	}

	err := fn()
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = CircuitOpen
		cb.nextRetry = time.Now().Add(cb.timeout)
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.failures = 0
	cb.state = CircuitClosed
}

// MarketDataService defines the interface for market data operations
type MarketDataService interface {
	GetQuote(ctx context.Context, symbol string) (*Quote, error)
	GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*Quote, error)
	GetQuotesByStockIDs(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*Quote, error)
	GetOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error)
}

// ExternalMarketDataService implements MarketDataService with external API integration
type ExternalMarketDataService struct {
	redisClient    *redis.Client
	httpClient     *http.Client
	circuitBreaker *CircuitBreaker
	apiKey         string
	baseURL        string
	cacheTTL       time.Duration
	provider       MarketDataProvider
}

// MarketDataProvider represents different market data providers
type MarketDataProvider string

const (
	ProviderTwelveData   MarketDataProvider = "twelvedata"
	ProviderAlphaVantage MarketDataProvider = "alphavantage"
	ProviderYahooFinance MarketDataProvider = "yahoo"
)

// NewExternalMarketDataService creates a new external market data service
func NewExternalMarketDataService(redisClient *redis.Client, apiKey string) *ExternalMarketDataService {
	// Determine provider based on API key format or configuration
	provider := ProviderTwelveData
	baseURL := "https://api.twelvedata.com/v1"
	
	// Alpha Vantage keys are typically longer and alphanumeric
	if len(apiKey) > 20 && strings.Contains(apiKey, "ALPHA") {
		provider = ProviderAlphaVantage
		baseURL = "https://www.alphavantage.co/query"
	}
	
	return &ExternalMarketDataService{
		redisClient: redisClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
		apiKey:         apiKey,
		baseURL:        baseURL,
		cacheTTL:       5 * time.Minute, // Cache quotes for 5 minutes
		provider:       provider,
	}
}

// NewYahooFinanceService creates a service that uses Yahoo Finance (no API key required)
func NewYahooFinanceService(redisClient *redis.Client) *ExternalMarketDataService {
	return &ExternalMarketDataService{
		redisClient: redisClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
		apiKey:         "", // Yahoo Finance doesn't require API key
		baseURL:        "https://query1.finance.yahoo.com/v8/finance/chart",
		cacheTTL:       5 * time.Minute,
		provider:       ProviderYahooFinance,
	}
}

// MarketDataServiceFactory creates the appropriate market data service based on configuration
type MarketDataServiceFactory struct {
	redisClient *redis.Client
}

// NewMarketDataServiceFactory creates a new factory
func NewMarketDataServiceFactory(redisClient *redis.Client) *MarketDataServiceFactory {
	return &MarketDataServiceFactory{
		redisClient: redisClient,
	}
}

// CreateService creates a market data service based on the provider type and API key
func (f *MarketDataServiceFactory) CreateService(provider string, apiKey string) MarketDataService {
	switch strings.ToLower(provider) {
	case "tradingview", "tv":
		return NewTradingViewService(f.redisClient)
	case "twelvedata", "12data":
		return NewExternalMarketDataService(f.redisClient, apiKey)
	case "alphavantage", "alpha":
		return NewExternalMarketDataService(f.redisClient, apiKey)
	case "yahoo", "yahoofinance":
		return NewYahooFinanceService(f.redisClient)
	case "mock", "":
		return NewMockMarketDataService()
	default:
		// Default to mock service for unknown providers
		return NewMockMarketDataService()
	}
}

// TwelveDataQuoteResponse represents the response from Twelve Data API
type TwelveDataQuoteResponse struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Exchange      string `json:"exchange"`
	Currency      string `json:"currency"`
	Datetime      string `json:"datetime"`
	Timestamp     int64  `json:"timestamp"`
	Open          string `json:"open"`
	High          string `json:"high"`
	Low           string `json:"low"`
	Close         string `json:"close"`
	Volume        string `json:"volume"`
	PreviousClose string `json:"previous_close"`
	Change        string `json:"change"`
	PercentChange string `json:"percent_change"`
}

// TwelveDataTimeSeriesResponse represents historical data response
type TwelveDataTimeSeriesResponse struct {
	Meta   TwelveDataMeta    `json:"meta"`
	Values []TwelveDataValue `json:"values"`
	Status string            `json:"status"`
}

type TwelveDataMeta struct {
	Symbol   string `json:"symbol"`
	Interval string `json:"interval"`
	Currency string `json:"currency"`
	Exchange string `json:"exchange"`
}

type TwelveDataValue struct {
	Datetime string `json:"datetime"`
	Open     string `json:"open"`
	High     string `json:"high"`
	Low      string `json:"low"`
	Close    string `json:"close"`
	Volume   string `json:"volume"`
}

// Yahoo Finance API response structures
type YahooFinanceResponse struct {
	Chart YahooChart `json:"chart"`
}

type YahooChart struct {
	Result []YahooResult `json:"result"`
	Error  interface{}   `json:"error"`
}

type YahooResult struct {
	Meta       YahooMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators YahooIndicators `json:"indicators"`
}

type YahooMeta struct {
	Currency             string  `json:"currency"`
	Symbol               string  `json:"symbol"`
	ExchangeName         string  `json:"exchangeName"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	PreviousClose        float64 `json:"previousClose"`
	RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
	RegularMarketVolume  int64   `json:"regularMarketVolume"`
}

type YahooIndicators struct {
	Quote []YahooQuote `json:"quote"`
}

type YahooQuote struct {
	Open   []float64 `json:"open"`
	High   []float64 `json:"high"`
	Low    []float64 `json:"low"`
	Close  []float64 `json:"close"`
	Volume []int64   `json:"volume"`
}

// Alpha Vantage API response structures
type AlphaVantageQuoteResponse struct {
	GlobalQuote AlphaVantageGlobalQuote `json:"Global Quote"`
}

type AlphaVantageGlobalQuote struct {
	Symbol           string `json:"01. symbol"`
	Open             string `json:"02. open"`
	High             string `json:"03. high"`
	Low              string `json:"04. low"`
	Price            string `json:"05. price"`
	Volume           string `json:"06. volume"`
	LatestTradingDay string `json:"07. latest trading day"`
	PreviousClose    string `json:"08. previous close"`
	Change           string `json:"09. change"`
	ChangePercent    string `json:"10. change percent"`
}

type AlphaVantageTimeSeriesResponse struct {
	MetaData   AlphaVantageMetaData            `json:"Meta Data"`
	TimeSeries map[string]AlphaVantageOHLCData `json:"Time Series (Daily)"`
}

type AlphaVantageMetaData struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

type AlphaVantageOHLCData struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

// GetQuote retrieves a quote for a single symbol with caching
func (s *ExternalMarketDataService) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("quote:%s", symbol)
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var quote Quote
		if json.Unmarshal([]byte(cached), &quote) == nil {
			return &quote, nil
		}
	}

	// Fetch from external API with circuit breaker
	var quote *Quote
	err = s.circuitBreaker.Call(func() error {
		var fetchErr error
		quote, fetchErr = s.fetchQuoteFromAPI(ctx, symbol)
		return fetchErr
	})

	if err != nil {
		// Return cached data if available, even if stale
		if cached != "" {
			var staleQuote Quote
			if json.Unmarshal([]byte(cached), &staleQuote) == nil {
				return &staleQuote, nil
			}
		}
		return nil, fmt.Errorf("failed to fetch quote for %s: %w", symbol, err)
	}

	// Cache the result
	if quote != nil {
		if data, marshalErr := json.Marshal(quote); marshalErr == nil {
			s.redisClient.Set(ctx, cacheKey, data, s.cacheTTL)
		}
	}

	return quote, nil
}

// GetMultipleQuotes retrieves quotes for multiple symbols with batch processing
func (s *ExternalMarketDataService) GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*Quote, error) {
	result := make(map[string]*Quote)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Process in batches to avoid overwhelming the API
	batchSize := 10
	for i := 0; i < len(symbols); i += batchSize {
		end := i + batchSize
		if end > len(symbols) {
			end = len(symbols)
		}

		batch := symbols[i:end]
		wg.Add(1)

		go func(batch []string) {
			defer wg.Done()
			for _, symbol := range batch {
				quote, err := s.GetQuote(ctx, symbol)
				mu.Lock()
				if err != nil {
					errors = append(errors, fmt.Errorf("failed to get quote for %s: %w", symbol, err))
				} else {
					result[symbol] = quote
				}
				mu.Unlock()
			}
		}(batch)
	}

	wg.Wait()

	if len(errors) > 0 && len(result) == 0 {
		return nil, fmt.Errorf("failed to fetch any quotes: %v", errors)
	}

	return result, nil
}

// GetQuotesByStockIDs retrieves quotes for stocks by their IDs
func (s *ExternalMarketDataService) GetQuotesByStockIDs(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*Quote, error) {
	// This would require integration with stock repository to get tickers
	// For now, return an error indicating this needs stock repository integration
	return nil, fmt.Errorf("GetQuotesByStockIDs requires stock repository integration")
}

// GetOHLCV retrieves historical OHLCV data
func (s *ExternalMarketDataService) GetOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("ohlcv:%s:%s:%s:%s", symbol, from.Format("2006-01-02"), to.Format("2006-01-02"), interval)
	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var ohlcv []*OHLCV
		if json.Unmarshal([]byte(cached), &ohlcv) == nil {
			return ohlcv, nil
		}
	}

	// Fetch from API
	var ohlcv []*OHLCV
	err = s.circuitBreaker.Call(func() error {
		var fetchErr error
		ohlcv, fetchErr = s.fetchOHLCVFromAPI(ctx, symbol, from, to, interval)
		return fetchErr
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch OHLCV for %s: %w", symbol, err)
	}

	// Cache the result (longer TTL for historical data)
	if ohlcv != nil {
		if data, marshalErr := json.Marshal(ohlcv); marshalErr == nil {
			s.redisClient.Set(ctx, cacheKey, data, 1*time.Hour)
		}
	}

	return ohlcv, nil
}

// fetchQuoteFromAPI fetches a quote from the external API
func (s *ExternalMarketDataService) fetchQuoteFromAPI(ctx context.Context, symbol string) (*Quote, error) {
	switch s.provider {
	case ProviderTwelveData:
		return s.fetchTwelveDataQuote(ctx, symbol)
	case ProviderAlphaVantage:
		return s.fetchAlphaVantageQuote(ctx, symbol)
	case ProviderYahooFinance:
		return s.fetchYahooFinanceQuote(ctx, symbol)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", s.provider)
	}
}

// fetchTwelveDataQuote fetches quote from Twelve Data API
func (s *ExternalMarketDataService) fetchTwelveDataQuote(ctx context.Context, symbol string) (*Quote, error) {
	url := fmt.Sprintf("%s/quote?symbol=%s&apikey=%s", s.baseURL, symbol, s.apiKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp TwelveDataQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return s.convertTwelveDataToQuote(&apiResp)
}

// fetchAlphaVantageQuote fetches quote from Alpha Vantage API
func (s *ExternalMarketDataService) fetchAlphaVantageQuote(ctx context.Context, symbol string) (*Quote, error) {
	url := fmt.Sprintf("%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", s.baseURL, symbol, s.apiKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp AlphaVantageQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return s.convertAlphaVantageToQuote(&apiResp.GlobalQuote)
}

// fetchYahooFinanceQuote fetches quote from Yahoo Finance API
func (s *ExternalMarketDataService) fetchYahooFinanceQuote(ctx context.Context, symbol string) (*Quote, error) {
	url := fmt.Sprintf("%s/%s?interval=1d&range=1d", s.baseURL, symbol)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add user agent to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PortfolioApp/1.0)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp YahooFinanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data found for symbol %s", symbol)
	}

	return s.convertYahooFinanceToQuote(&apiResp.Chart.Result[0])
}

// fetchOHLCVFromAPI fetches historical data from the external API
func (s *ExternalMarketDataService) fetchOHLCVFromAPI(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error) {
	switch s.provider {
	case ProviderTwelveData:
		return s.fetchTwelveDataOHLCV(ctx, symbol, from, to, interval)
	case ProviderAlphaVantage:
		return s.fetchAlphaVantageOHLCV(ctx, symbol, from, to, interval)
	case ProviderYahooFinance:
		return s.fetchYahooFinanceOHLCV(ctx, symbol, from, to, interval)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", s.provider)
	}
}

// fetchTwelveDataOHLCV fetches historical data from Twelve Data API
func (s *ExternalMarketDataService) fetchTwelveDataOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error) {
	url := fmt.Sprintf("%s/time_series?symbol=%s&interval=%s&start_date=%s&end_date=%s&apikey=%s",
		s.baseURL, symbol, interval, from.Format("2006-01-02"), to.Format("2006-01-02"), s.apiKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp TwelveDataTimeSeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return s.convertTwelveDataToOHLCV(apiResp.Values)
}

// fetchAlphaVantageOHLCV fetches historical data from Alpha Vantage API
func (s *ExternalMarketDataService) fetchAlphaVantageOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error) {
	// Alpha Vantage uses different function names for different intervals
	function := "TIME_SERIES_DAILY"
	if interval == "1min" {
		function = "TIME_SERIES_INTRADAY"
	}
	
	url := fmt.Sprintf("%s?function=%s&symbol=%s&apikey=%s", s.baseURL, function, symbol, s.apiKey)
	if function == "TIME_SERIES_INTRADAY" {
		url += "&interval=1min"
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp AlphaVantageTimeSeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return s.convertAlphaVantageToOHLCV(apiResp.TimeSeries, from, to)
}

// fetchYahooFinanceOHLCV fetches historical data from Yahoo Finance API
func (s *ExternalMarketDataService) fetchYahooFinanceOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error) {
	// Convert interval to Yahoo Finance format
	yahooInterval := "1d"
	switch interval {
	case "1min":
		yahooInterval = "1m"
	case "5min":
		yahooInterval = "5m"
	case "1h":
		yahooInterval = "1h"
	case "1day":
		yahooInterval = "1d"
	}
	
	url := fmt.Sprintf("%s/%s?interval=%s&period1=%d&period2=%d",
		s.baseURL, symbol, yahooInterval, from.Unix(), to.Unix())
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PortfolioApp/1.0)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp YahooFinanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data found for symbol %s", symbol)
	}

	return s.convertYahooFinanceToOHLCV(&apiResp.Chart.Result[0])
}

// convertTwelveDataToQuote converts Twelve Data API response to internal Quote struct
func (s *ExternalMarketDataService) convertTwelveDataToQuote(apiResp *TwelveDataQuoteResponse) (*Quote, error) {
	price, err := decimal.NewFromString(apiResp.Close)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	change, _ := decimal.NewFromString(apiResp.Change)
	changePercent, _ := decimal.NewFromString(apiResp.PercentChange)
	volume, _ := strconv.ParseInt(apiResp.Volume, 10, 64)
	high, _ := decimal.NewFromString(apiResp.High)
	low, _ := decimal.NewFromString(apiResp.Low)
	open, _ := decimal.NewFromString(apiResp.Open)
	previousClose, _ := decimal.NewFromString(apiResp.PreviousClose)

	timestamp := time.Now()
	if apiResp.Timestamp > 0 {
		timestamp = time.Unix(apiResp.Timestamp, 0)
	}

	return &Quote{
		Symbol:        apiResp.Symbol,
		Price:         price,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        volume,
		High:          high,
		Low:           low,
		Open:          open,
		PreviousClose: previousClose,
		Timestamp:     timestamp,
	}, nil
}

// convertAlphaVantageToQuote converts Alpha Vantage API response to internal Quote struct
func (s *ExternalMarketDataService) convertAlphaVantageToQuote(apiResp *AlphaVantageGlobalQuote) (*Quote, error) {
	price, err := decimal.NewFromString(apiResp.Price)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	change, _ := decimal.NewFromString(apiResp.Change)
	
	// Parse change percent (format: "1.23%")
	changePercentStr := strings.TrimSuffix(apiResp.ChangePercent, "%")
	changePercent, _ := decimal.NewFromString(changePercentStr)
	
	volume, _ := strconv.ParseInt(apiResp.Volume, 10, 64)
	high, _ := decimal.NewFromString(apiResp.High)
	low, _ := decimal.NewFromString(apiResp.Low)
	open, _ := decimal.NewFromString(apiResp.Open)
	previousClose, _ := decimal.NewFromString(apiResp.PreviousClose)

	// Parse timestamp from latest trading day
	timestamp := time.Now()
	if parsedTime, err := time.Parse("2006-01-02", apiResp.LatestTradingDay); err == nil {
		timestamp = parsedTime
	}

	return &Quote{
		Symbol:        apiResp.Symbol,
		Price:         price,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        volume,
		High:          high,
		Low:           low,
		Open:          open,
		PreviousClose: previousClose,
		Timestamp:     timestamp,
	}, nil
}

// convertYahooFinanceToQuote converts Yahoo Finance API response to internal Quote struct
func (s *ExternalMarketDataService) convertYahooFinanceToQuote(result *YahooResult) (*Quote, error) {
	meta := result.Meta
	
	price := decimal.NewFromFloat(meta.RegularMarketPrice)
	previousClose := decimal.NewFromFloat(meta.PreviousClose)
	change := price.Sub(previousClose)
	changePercent := decimal.Zero
	if previousClose.GreaterThan(decimal.Zero) {
		changePercent = change.Div(previousClose).Mul(decimal.NewFromInt(100))
	}

	high := decimal.NewFromFloat(meta.RegularMarketDayHigh)
	low := decimal.NewFromFloat(meta.RegularMarketDayLow)
	volume := meta.RegularMarketVolume

	// Get open price from indicators if available
	open := previousClose
	if len(result.Indicators.Quote) > 0 && len(result.Indicators.Quote[0].Open) > 0 {
		lastIdx := len(result.Indicators.Quote[0].Open) - 1
		if result.Indicators.Quote[0].Open[lastIdx] != 0 {
			open = decimal.NewFromFloat(result.Indicators.Quote[0].Open[lastIdx])
		}
	}

	return &Quote{
		Symbol:        meta.Symbol,
		Price:         price,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        volume,
		High:          high,
		Low:           low,
		Open:          open,
		PreviousClose: previousClose,
		Timestamp:     time.Now(),
	}, nil
}

// convertTwelveDataToOHLCV converts Twelve Data API response to internal OHLCV structs
func (s *ExternalMarketDataService) convertTwelveDataToOHLCV(values []TwelveDataValue) ([]*OHLCV, error) {
	var result []*OHLCV

	for _, value := range values {
		timestamp, err := time.Parse("2006-01-02 15:04:05", value.Datetime)
		if err != nil {
			// Try alternative format
			timestamp, err = time.Parse("2006-01-02", value.Datetime)
			if err != nil {
				continue
			}
		}

		open, err := decimal.NewFromString(value.Open)
		if err != nil {
			continue
		}
		high, err := decimal.NewFromString(value.High)
		if err != nil {
			continue
		}
		low, err := decimal.NewFromString(value.Low)
		if err != nil {
			continue
		}
		close, err := decimal.NewFromString(value.Close)
		if err != nil {
			continue
		}
		volume, _ := strconv.ParseInt(value.Volume, 10, 64)

		result = append(result, &OHLCV{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	return result, nil
}

// convertAlphaVantageToOHLCV converts Alpha Vantage time series to OHLCV
func (s *ExternalMarketDataService) convertAlphaVantageToOHLCV(timeSeries map[string]AlphaVantageOHLCData, from, to time.Time) ([]*OHLCV, error) {
	var result []*OHLCV

	for dateStr, data := range timeSeries {
		timestamp, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// Filter by date range
		if timestamp.Before(from) || timestamp.After(to) {
			continue
		}

		open, err := decimal.NewFromString(data.Open)
		if err != nil {
			continue
		}
		high, err := decimal.NewFromString(data.High)
		if err != nil {
			continue
		}
		low, err := decimal.NewFromString(data.Low)
		if err != nil {
			continue
		}
		close, err := decimal.NewFromString(data.Close)
		if err != nil {
			continue
		}
		volume, _ := strconv.ParseInt(data.Volume, 10, 64)

		result = append(result, &OHLCV{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	return result, nil
}

// convertYahooFinanceToOHLCV converts Yahoo Finance response to OHLCV
func (s *ExternalMarketDataService) convertYahooFinanceToOHLCV(result *YahooResult) ([]*OHLCV, error) {
	var ohlcvData []*OHLCV

	if len(result.Indicators.Quote) == 0 {
		return ohlcvData, nil
	}

	quote := result.Indicators.Quote[0]
	timestamps := result.Timestamp

	// Ensure all arrays have the same length
	minLen := len(timestamps)
	if len(quote.Open) < minLen {
		minLen = len(quote.Open)
	}
	if len(quote.High) < minLen {
		minLen = len(quote.High)
	}
	if len(quote.Low) < minLen {
		minLen = len(quote.Low)
	}
	if len(quote.Close) < minLen {
		minLen = len(quote.Close)
	}
	if len(quote.Volume) < minLen {
		minLen = len(quote.Volume)
	}

	for i := 0; i < minLen; i++ {
		// Skip entries with zero values
		if quote.Open[i] == 0 || quote.High[i] == 0 || quote.Low[i] == 0 || quote.Close[i] == 0 {
			continue
		}

		ohlcvData = append(ohlcvData, &OHLCV{
			Timestamp: time.Unix(timestamps[i], 0),
			Open:      decimal.NewFromFloat(quote.Open[i]),
			High:      decimal.NewFromFloat(quote.High[i]),
			Low:       decimal.NewFromFloat(quote.Low[i]),
			Close:     decimal.NewFromFloat(quote.Close[i]),
			Volume:    quote.Volume[i],
		})
	}

	return ohlcvData, nil
}

// MockMarketDataService provides mock market data for testing and development
type MockMarketDataService struct {
	quotes map[string]*Quote
}

// NewMockMarketDataService creates a new mock market data service
func NewMockMarketDataService() *MockMarketDataService {
	// Initialize with some sample quotes
	quotes := map[string]*Quote{
		"AAPL": {
			Symbol:        "AAPL",
			Price:         decimal.NewFromFloat(150.25),
			Change:        decimal.NewFromFloat(2.15),
			ChangePercent: decimal.NewFromFloat(1.45),
			Volume:        1000000,
			High:          decimal.NewFromFloat(152.00),
			Low:           decimal.NewFromFloat(148.50),
			Open:          decimal.NewFromFloat(149.00),
			PreviousClose: decimal.NewFromFloat(148.10),
			Timestamp:     time.Now(),
		},
		"GOOGL": {
			Symbol:        "GOOGL",
			Price:         decimal.NewFromFloat(2750.80),
			Change:        decimal.NewFromFloat(-15.20),
			ChangePercent: decimal.NewFromFloat(-0.55),
			Volume:        500000,
			High:          decimal.NewFromFloat(2770.00),
			Low:           decimal.NewFromFloat(2745.00),
			Open:          decimal.NewFromFloat(2766.00),
			PreviousClose: decimal.NewFromFloat(2766.00),
			Timestamp:     time.Now(),
		},
		"MSFT": {
			Symbol:        "MSFT",
			Price:         decimal.NewFromFloat(305.45),
			Change:        decimal.NewFromFloat(5.30),
			ChangePercent: decimal.NewFromFloat(1.77),
			Volume:        800000,
			High:          decimal.NewFromFloat(307.00),
			Low:           decimal.NewFromFloat(300.15),
			Open:          decimal.NewFromFloat(301.00),
			PreviousClose: decimal.NewFromFloat(300.15),
			Timestamp:     time.Now(),
		},
		"TSLA": {
			Symbol:        "TSLA",
			Price:         decimal.NewFromFloat(245.67),
			Change:        decimal.NewFromFloat(-8.45),
			ChangePercent: decimal.NewFromFloat(-3.32),
			Volume:        1200000,
			High:          decimal.NewFromFloat(255.00),
			Low:           decimal.NewFromFloat(244.00),
			Open:          decimal.NewFromFloat(254.12),
			PreviousClose: decimal.NewFromFloat(254.12),
			Timestamp:     time.Now(),
		},
		"NVDA": {
			Symbol:        "NVDA",
			Price:         decimal.NewFromFloat(875.30),
			Change:        decimal.NewFromFloat(12.80),
			ChangePercent: decimal.NewFromFloat(1.48),
			Volume:        600000,
			High:          decimal.NewFromFloat(880.00),
			Low:           decimal.NewFromFloat(862.50),
			Open:          decimal.NewFromFloat(865.00),
			PreviousClose: decimal.NewFromFloat(862.50),
			Timestamp:     time.Now(),
		},
	}

	return &MockMarketDataService{
		quotes: quotes,
	}
}

// GetQuote retrieves a quote for a single symbol
func (m *MockMarketDataService) GetQuote(ctx context.Context, symbol string) (*Quote, error) {
	quote, exists := m.quotes[symbol]
	if !exists {
		// Return a default quote for unknown symbols
		return &Quote{
			Symbol:        symbol,
			Price:         decimal.NewFromFloat(100.00),
			Change:        decimal.Zero,
			ChangePercent: decimal.Zero,
			Volume:        0,
			High:          decimal.NewFromFloat(102.00),
			Low:           decimal.NewFromFloat(98.00),
			Open:          decimal.NewFromFloat(99.50),
			PreviousClose: decimal.NewFromFloat(100.00),
			Timestamp:     time.Now(),
		}, nil
	}

	// Return a copy with updated timestamp
	quoteCopy := *quote
	quoteCopy.Timestamp = time.Now()
	return &quoteCopy, nil
}

// GetMultipleQuotes retrieves quotes for multiple symbols
func (m *MockMarketDataService) GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*Quote, error) {
	result := make(map[string]*Quote)
	
	for _, symbol := range symbols {
		quote, err := m.GetQuote(ctx, symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get quote for %s: %w", symbol, err)
		}
		result[symbol] = quote
	}
	
	return result, nil
}

// GetQuotesByStockIDs retrieves quotes for stocks by their IDs (requires stock repository)
func (m *MockMarketDataService) GetQuotesByStockIDs(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*Quote, error) {
	// This would typically require a stock repository to get tickers from IDs
	// For now, return an error indicating this needs integration
	return nil, fmt.Errorf("GetQuotesByStockIDs requires stock repository integration - to be implemented in task 7")
}

// GetOHLCV retrieves historical OHLCV data (mock implementation)
func (m *MockMarketDataService) GetOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*OHLCV, error) {
	// Generate mock historical data
	var result []*OHLCV
	current := from
	basePrice := decimal.NewFromFloat(100.0)
	
	if quote, exists := m.quotes[symbol]; exists {
		basePrice = quote.Price
	}

	for current.Before(to) || current.Equal(to) {
		// Generate realistic OHLCV data with some randomness
		open := basePrice.Add(decimal.NewFromFloat(float64((current.Unix()%10)-5) * 0.5))
		
		// Ensure high is the highest value
		highVariation := decimal.NewFromFloat(float64(current.Unix()%5) * 0.3)
		high := open.Add(highVariation)
		
		// Ensure low is the lowest value
		lowVariation := decimal.NewFromFloat(float64(current.Unix()%3) * 0.2)
		low := open.Sub(lowVariation)
		
		// Close should be between low and high
		closeVariation := decimal.NewFromFloat(float64((current.Unix()%8)-4) * 0.1)
		close := open.Add(closeVariation)
		
		// Adjust high and low to ensure proper OHLC relationships
		if close.GreaterThan(high) {
			high = close
		}
		if close.LessThan(low) {
			low = close
		}
		if open.GreaterThan(high) {
			high = open
		}
		if open.LessThan(low) {
			low = open
		}
		
		volume := int64(100000 + (current.Unix()%500000))

		result = append(result, &OHLCV{
			Timestamp: current,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})

		// Move to next interval
		switch interval {
		case "1min":
			current = current.Add(time.Minute)
		case "5min":
			current = current.Add(5 * time.Minute)
		case "1h":
			current = current.Add(time.Hour)
		case "1day":
			current = current.Add(24 * time.Hour)
		default:
			current = current.Add(24 * time.Hour)
		}
	}

	return result, nil
}

// SetQuote allows setting custom quotes for testing
func (m *MockMarketDataService) SetQuote(symbol string, quote *Quote) {
	m.quotes[symbol] = quote
}

// AddQuote adds a new quote to the mock service
func (m *MockMarketDataService) AddQuote(symbol string, price float64) {
	m.quotes[symbol] = &Quote{
		Symbol:        symbol,
		Price:         decimal.NewFromFloat(price),
		Change:        decimal.Zero,
		ChangePercent: decimal.Zero,
		Volume:        0,
		High:          decimal.NewFromFloat(price * 1.02),
		Low:           decimal.NewFromFloat(price * 0.98),
		Open:          decimal.NewFromFloat(price * 0.995),
		PreviousClose: decimal.NewFromFloat(price),
		Timestamp:     time.Now(),
	}
}