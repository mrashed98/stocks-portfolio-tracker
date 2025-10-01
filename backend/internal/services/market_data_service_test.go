package services

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockMarketDataService_GetQuote(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should return quote for existing symbol", func(t *testing.T) {
		quote, err := service.GetQuote(ctx, "AAPL")
		require.NoError(t, err)
		assert.Equal(t, "AAPL", quote.Symbol)
		assert.True(t, quote.Price.GreaterThan(decimal.Zero))
		assert.NotZero(t, quote.Volume)
		assert.True(t, quote.High.GreaterThanOrEqual(quote.Low))
	})

	t.Run("should return default quote for unknown symbol", func(t *testing.T) {
		quote, err := service.GetQuote(ctx, "UNKNOWN")
		require.NoError(t, err)
		assert.Equal(t, "UNKNOWN", quote.Symbol)
		assert.Equal(t, decimal.NewFromFloat(100.00), quote.Price)
	})

	t.Run("should update timestamp on each call", func(t *testing.T) {
		quote1, err := service.GetQuote(ctx, "AAPL")
		require.NoError(t, err)
		
		time.Sleep(10 * time.Millisecond)
		
		quote2, err := service.GetQuote(ctx, "AAPL")
		require.NoError(t, err)
		
		assert.True(t, quote2.Timestamp.After(quote1.Timestamp))
	})
}

func TestMockMarketDataService_GetMultipleQuotes(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should return quotes for multiple symbols", func(t *testing.T) {
		symbols := []string{"AAPL", "GOOGL", "MSFT"}
		quotes, err := service.GetMultipleQuotes(ctx, symbols)
		require.NoError(t, err)
		
		assert.Len(t, quotes, 3)
		for _, symbol := range symbols {
			quote, exists := quotes[symbol]
			assert.True(t, exists)
			assert.Equal(t, symbol, quote.Symbol)
		}
	})

	t.Run("should handle mix of known and unknown symbols", func(t *testing.T) {
		symbols := []string{"AAPL", "UNKNOWN", "GOOGL"}
		quotes, err := service.GetMultipleQuotes(ctx, symbols)
		require.NoError(t, err)
		
		assert.Len(t, quotes, 3)
		assert.Equal(t, "AAPL", quotes["AAPL"].Symbol)
		assert.Equal(t, "UNKNOWN", quotes["UNKNOWN"].Symbol)
		assert.Equal(t, "GOOGL", quotes["GOOGL"].Symbol)
	})
}

func TestMockMarketDataService_GetOHLCV(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should return historical data for date range", func(t *testing.T) {
		from := time.Now().AddDate(0, 0, -7) // 7 days ago
		to := time.Now().AddDate(0, 0, -1)   // 1 day ago
		
		ohlcv, err := service.GetOHLCV(ctx, "AAPL", from, to, "1day")
		require.NoError(t, err)
		
		assert.NotEmpty(t, ohlcv)
		
		for _, data := range ohlcv {
			assert.True(t, data.High.GreaterThanOrEqual(data.Low))
			assert.True(t, data.High.GreaterThanOrEqual(data.Open))
			assert.True(t, data.High.GreaterThanOrEqual(data.Close))
			assert.True(t, data.Low.LessThanOrEqual(data.Open))
			assert.True(t, data.Low.LessThanOrEqual(data.Close))
			assert.True(t, data.Volume > 0)
		}
	})

	t.Run("should handle different intervals", func(t *testing.T) {
		from := time.Now().Add(-2 * time.Hour)
		to := time.Now().Add(-1 * time.Hour)
		
		ohlcv, err := service.GetOHLCV(ctx, "AAPL", from, to, "1min")
		require.NoError(t, err)
		
		assert.NotEmpty(t, ohlcv)
		// Should have approximately 60 data points for 1-hour range with 1-minute intervals
		assert.True(t, len(ohlcv) > 50)
	})
}

func TestMockMarketDataService_SetQuote(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should allow setting custom quotes", func(t *testing.T) {
		customQuote := &Quote{
			Symbol:        "TEST",
			Price:         decimal.NewFromFloat(250.50),
			Change:        decimal.NewFromFloat(5.25),
			ChangePercent: decimal.NewFromFloat(2.14),
			Volume:        1500000,
			High:          decimal.NewFromFloat(252.00),
			Low:           decimal.NewFromFloat(248.00),
			Open:          decimal.NewFromFloat(249.00),
			PreviousClose: decimal.NewFromFloat(245.25),
			Timestamp:     time.Now(),
		}

		service.SetQuote("TEST", customQuote)
		
		retrievedQuote, err := service.GetQuote(ctx, "TEST")
		require.NoError(t, err)
		
		assert.Equal(t, customQuote.Symbol, retrievedQuote.Symbol)
		assert.Equal(t, customQuote.Price, retrievedQuote.Price)
		assert.Equal(t, customQuote.Change, retrievedQuote.Change)
		assert.Equal(t, customQuote.Volume, retrievedQuote.Volume)
	})
}

func TestMockMarketDataService_AddQuote(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should add new quote with price", func(t *testing.T) {
		service.AddQuote("NEWSTOCK", 175.25)
		
		quote, err := service.GetQuote(ctx, "NEWSTOCK")
		require.NoError(t, err)
		
		assert.Equal(t, "NEWSTOCK", quote.Symbol)
		assert.Equal(t, decimal.NewFromFloat(175.25), quote.Price)
		assert.Equal(t, decimal.Zero, quote.Change)
		assert.Equal(t, decimal.Zero, quote.ChangePercent)
		assert.True(t, quote.High.GreaterThan(quote.Price))
		assert.True(t, quote.Low.LessThan(quote.Price))
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("should allow calls when closed", func(t *testing.T) {
		cb := NewCircuitBreaker(3, 1*time.Second)
		
		err := cb.Call(func() error {
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, CircuitClosed, cb.state)
	})

	t.Run("should open after max failures", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 1*time.Second)
		
		// First failure
		err := cb.Call(func() error {
			return assert.AnError
		})
		assert.Error(t, err)
		assert.Equal(t, CircuitClosed, cb.state)
		
		// Second failure - should open circuit
		err = cb.Call(func() error {
			return assert.AnError
		})
		assert.Error(t, err)
		assert.Equal(t, CircuitOpen, cb.state)
		
		// Third call should be rejected
		err = cb.Call(func() error {
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker is open")
	})

	t.Run("should reset after successful call", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 1*time.Second)
		
		// Cause one failure
		cb.Call(func() error {
			return assert.AnError
		})
		
		// Successful call should reset failures
		err := cb.Call(func() error {
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, CircuitClosed, cb.state)
		assert.Equal(t, 0, cb.failures)
	})

	t.Run("should transition to half-open after timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 10*time.Millisecond)
		
		// Cause failure to open circuit
		cb.Call(func() error {
			return assert.AnError
		})
		assert.Equal(t, CircuitOpen, cb.state)
		
		// Wait for timeout
		time.Sleep(15 * time.Millisecond)
		
		// Next call should transition to half-open
		err := cb.Call(func() error {
			return nil
		})
		
		assert.NoError(t, err)
		assert.Equal(t, CircuitClosed, cb.state)
	})
}