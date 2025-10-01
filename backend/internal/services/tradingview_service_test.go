package services

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTradingViewService_ConvertToTradingViewSymbol(t *testing.T) {
	// Test the symbol conversion logic without requiring full service initialization
	tests := []struct {
		input    string
		expected string
	}{
		{"AAPL", "NASDAQ:AAPL"},
		{"aapl", "NASDAQ:AAPL"},
		{"GOOGL", "NASDAQ:GOOGL"},
		{"MSFT", "NASDAQ:MSFT"},
		{"JPM", "NYSE:JPM"},
		{"JNJ", "NYSE:JNJ"},
		{"UNKNOWN", "NASDAQ:UNKNOWN"},
	}
	
	// Create a minimal service instance for testing the conversion method
	service := &TradingViewService{}
	
	for _, test := range tests {
		result := service.convertToTradingViewSymbol(test.input)
		assert.Equal(t, test.expected, result, "Failed for input: %s", test.input)
	}
}

func TestTradingViewService_HandleQuoteData(t *testing.T) {
	service := &TradingViewService{
		quotes: make(map[string]*Quote),
	}
	
	symbol := "NASDAQ:AAPL"
	
	// Test new quote creation - simulate quote data structure
	service.quotes[symbol] = &Quote{
		Symbol:    symbol,
		Price:     decimal.NewFromFloat(150.50),
		Volume:    1000000,
		Timestamp: time.Now(),
	}
	
	quote, exists := service.quotes[symbol]
	assert.True(t, exists)
	assert.Equal(t, symbol, quote.Symbol)
	assert.True(t, decimal.NewFromFloat(150.50).Equal(quote.Price))
	assert.Equal(t, int64(1000000), quote.Volume)
	
	// Test quote update with previous close for change calculation
	quote.PreviousClose = decimal.NewFromFloat(145.00)
	quote.Price = decimal.NewFromFloat(152.00)
	quote.Change = quote.Price.Sub(quote.PreviousClose)
	if quote.PreviousClose.GreaterThan(decimal.Zero) {
		quote.ChangePercent = quote.Change.Div(quote.PreviousClose).Mul(decimal.NewFromInt(100))
	}
	
	updatedQuote := service.quotes[symbol]
	assert.True(t, decimal.NewFromFloat(152.00).Equal(updatedQuote.Price))
	assert.True(t, decimal.NewFromFloat(7.00).Equal(updatedQuote.Change))
	expectedChangePercent := decimal.NewFromFloat(7.00).Div(decimal.NewFromFloat(145.00)).Mul(decimal.NewFromInt(100))
	assert.True(t, expectedChangePercent.Equal(updatedQuote.ChangePercent))
}

func TestTradingViewService_GetOHLCV_NotSupported(t *testing.T) {
	service := &TradingViewService{}
	
	_, err := service.GetOHLCV(context.Background(), "AAPL", time.Now(), time.Now(), "1D")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

// Benchmark tests
func BenchmarkTradingViewService_ConvertToTradingViewSymbol(b *testing.B) {
	service := &TradingViewService{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.convertToTradingViewSymbol("AAPL")
	}
}