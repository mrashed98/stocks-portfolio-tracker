package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"portfolio-app/internal/services"
)

// MockMarketDataService is a mock implementation of MarketDataService
type MockMarketDataService struct {
	mock.Mock
}

func (m *MockMarketDataService) GetQuote(ctx context.Context, symbol string) (*services.Quote, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.Quote), args.Error(1)
}

func (m *MockMarketDataService) GetMultipleQuotes(ctx context.Context, symbols []string) (map[string]*services.Quote, error) {
	args := m.Called(ctx, symbols)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*services.Quote), args.Error(1)
}

func (m *MockMarketDataService) GetQuotesByStockIDs(ctx context.Context, stockIDs []uuid.UUID) (map[uuid.UUID]*services.Quote, error) {
	args := m.Called(ctx, stockIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*services.Quote), args.Error(1)
}

func (m *MockMarketDataService) GetOHLCV(ctx context.Context, symbol string, from, to time.Time, interval string) ([]*services.OHLCV, error) {
	args := m.Called(ctx, symbol, from, to, interval)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*services.OHLCV), args.Error(1)
}

func TestMarketDataHandler_GetQuote(t *testing.T) {
	tests := []struct {
		name           string
		ticker         string
		mockQuote      *services.Quote
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful quote retrieval",
			ticker: "AAPL",
			mockQuote: &services.Quote{
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
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "service error",
			ticker:         "INVALID",
			mockQuote:      nil,
			mockError:      assert.AnError,
			expectedStatus: 500,
			expectedBody:   "Failed to get quote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockMarketDataService)
			handler := NewMarketDataHandler(mockService)
			app := fiber.New()

			// Mock expectations
			if tt.ticker != "" {
				mockService.On("GetQuote", mock.Anything, tt.ticker).Return(tt.mockQuote, tt.mockError)
			}

			// Setup route
			app.Get("/quotes/:ticker", handler.GetQuote)

			// Create request
			req := httptest.NewRequest("GET", "/quotes/"+tt.ticker, nil)
			resp, err := app.Test(req)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMarketDataHandler_GetMultipleQuotes(t *testing.T) {
	tests := []struct {
		name           string
		symbols        string
		mockQuotes     map[string]*services.Quote
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "successful multiple quotes",
			symbols: "AAPL,GOOGL",
			mockQuotes: map[string]*services.Quote{
				"AAPL": {
					Symbol: "AAPL",
					Price:  decimal.NewFromFloat(150.25),
				},
				"GOOGL": {
					Symbol: "GOOGL",
					Price:  decimal.NewFromFloat(2750.80),
				},
			},
			mockError:      nil,
			expectedStatus: 200,
		},
		{
			name:           "empty symbols parameter",
			symbols:        "",
			mockQuotes:     nil,
			mockError:      nil,
			expectedStatus: 400,
			expectedBody:   "symbols parameter is required",
		},
		{
			name:           "too many symbols",
			symbols:        generateLongSymbolList(51),
			mockQuotes:     nil,
			mockError:      nil,
			expectedStatus: 400,
			expectedBody:   "Maximum 50 symbols allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockMarketDataService)
			handler := NewMarketDataHandler(mockService)
			app := fiber.New()

			// Mock expectations
			if tt.symbols != "" && len(tt.symbols) <= 200 { // Reasonable length check
				expectedSymbols := parseSymbols(tt.symbols)
				if len(expectedSymbols) <= 50 {
					mockService.On("GetMultipleQuotes", mock.Anything, expectedSymbols).Return(tt.mockQuotes, tt.mockError)
				}
			}

			// Setup route
			app.Get("/quotes", handler.GetMultipleQuotes)

			// Create request
			req := httptest.NewRequest("GET", "/quotes?symbols="+tt.symbols, nil)
			resp, err := app.Test(req)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMarketDataHandler_GetOHLCV(t *testing.T) {
	now := time.Now()
	from := now.AddDate(0, 0, -30)

	tests := []struct {
		name           string
		ticker         string
		from           string
		to             string
		interval       string
		mockOHLCV      []*services.OHLCV
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful OHLCV retrieval",
			ticker:   "AAPL",
			from:     from.Format("2006-01-02"),
			to:       now.Format("2006-01-02"),
			interval: "1day",
			mockOHLCV: []*services.OHLCV{
				{
					Timestamp: now,
					Open:      decimal.NewFromFloat(149.00),
					High:      decimal.NewFromFloat(152.00),
					Low:       decimal.NewFromFloat(148.50),
					Close:     decimal.NewFromFloat(150.25),
					Volume:    1000000,
				},
			},
			mockError:      nil,
			expectedStatus: 200,
		},

		{
			name:           "invalid from date",
			ticker:         "AAPL",
			from:           "invalid-date",
			expectedStatus: 400,
			expectedBody:   "Invalid 'from' date format",
		},
		{
			name:           "invalid interval",
			ticker:         "AAPL",
			interval:       "invalid",
			expectedStatus: 400,
			expectedBody:   "Invalid interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockMarketDataService)
			handler := NewMarketDataHandler(mockService)
			app := fiber.New()

			// Mock expectations
			if tt.ticker != "" && tt.from != "invalid-date" && tt.interval != "invalid" {
				mockService.On("GetOHLCV", mock.Anything, tt.ticker, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).Return(tt.mockOHLCV, tt.mockError)
			}

			// Setup route
			app.Get("/ohlcv/:ticker", handler.GetOHLCV)

			// Create request URL
			url := "/ohlcv/" + tt.ticker
			params := []string{}
			if tt.from != "" {
				params = append(params, "from="+tt.from)
			}
			if tt.to != "" {
				params = append(params, "to="+tt.to)
			}
			if tt.interval != "" {
				params = append(params, "interval="+tt.interval)
			}
			if len(params) > 0 {
				url += "?" + joinParams(params)
			}

			req := httptest.NewRequest("GET", url, nil)
			resp, err := app.Test(req)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMarketDataHandler_TradingViewSymbolSearch(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "search for AAPL",
			query:          "AAPL",
			expectedStatus: 200,
			expectedCount:  1,
		},
		{
			name:           "empty query returns all",
			query:          "",
			expectedStatus: 200,
			expectedCount:  5, // All mock symbols
		},
		{
			name:           "search for Apple",
			query:          "Apple",
			expectedStatus: 200,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockMarketDataService)
			handler := NewMarketDataHandler(mockService)
			app := fiber.New()

			// Setup route
			app.Get("/symbols", handler.TradingViewSymbolSearch)

			// Create request
			url := "/symbols"
			if tt.query != "" {
				url += "?query=" + tt.query
			}
			req := httptest.NewRequest("GET", url, nil)
			resp, err := app.Test(req)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var symbols []map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&symbols)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(symbols))
		})
	}
}

// Helper functions
func generateLongSymbolList(count int) string {
	symbols := make([]string, count)
	for i := 0; i < count; i++ {
		symbols[i] = "SYM" + string(rune('A'+i%26))
	}
	return joinSymbols(symbols)
}

func parseSymbols(symbolsStr string) []string {
	if symbolsStr == "" {
		return []string{}
	}
	symbols := []string{}
	for _, symbol := range splitSymbols(symbolsStr) {
		symbols = append(symbols, trimAndUpper(symbol))
	}
	return symbols
}

func joinSymbols(symbols []string) string {
	result := ""
	for i, symbol := range symbols {
		if i > 0 {
			result += ","
		}
		result += symbol
	}
	return result
}

func splitSymbols(symbolsStr string) []string {
	symbols := []string{}
	current := ""
	for _, char := range symbolsStr {
		if char == ',' {
			if current != "" {
				symbols = append(symbols, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		symbols = append(symbols, current)
	}
	return symbols
}

func trimAndUpper(s string) string {
	// Simple trim and upper implementation
	result := ""
	start := 0
	end := len(s) - 1
	
	// Trim leading spaces
	for start < len(s) && s[start] == ' ' {
		start++
	}
	
	// Trim trailing spaces
	for end >= start && s[end] == ' ' {
		end--
	}
	
	// Convert to upper
	for i := start; i <= end; i++ {
		char := s[i]
		if char >= 'a' && char <= 'z' {
			char = char - 'a' + 'A'
		}
		result += string(char)
	}
	
	return result
}

func joinParams(params []string) string {
	result := ""
	for i, param := range params {
		if i > 0 {
			result += "&"
		}
		result += param
	}
	return result
}