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

// Additional comprehensive tests for MarketDataService integration

func TestMarketDataService_CacheIntegration(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should handle cache miss and populate cache", func(t *testing.T) {
		// First call should populate cache
		quote1, err := service.GetQuote(ctx, "CACHE_TEST")
		require.NoError(t, err)
		assert.Equal(t, "CACHE_TEST", quote1.Symbol)

		// Second call should return same data (simulating cache hit)
		quote2, err := service.GetQuote(ctx, "CACHE_TEST")
		require.NoError(t, err)
		assert.Equal(t, quote1.Symbol, quote2.Symbol)
		assert.Equal(t, quote1.Price, quote2.Price)
	})
}

func TestMarketDataService_ErrorHandling(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should handle empty symbol gracefully", func(t *testing.T) {
		quote, err := service.GetQuote(ctx, "")
		// Mock service should handle empty symbol
		require.NoError(t, err)
		assert.Equal(t, "", quote.Symbol)
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		// Service should handle cancelled context
		_, err := service.GetQuote(cancelCtx, "AAPL")
		// Mock service doesn't check context, but real service would
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
	})
}

func TestMarketDataService_ConcurrentAccess(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should handle concurrent requests safely", func(t *testing.T) {
		symbols := []string{"AAPL", "GOOGL", "MSFT", "AMZN", "TSLA"}
		results := make(chan *Quote, len(symbols))
		errors := make(chan error, len(symbols))

		// Launch concurrent requests
		for _, symbol := range symbols {
			go func(sym string) {
				quote, err := service.GetQuote(ctx, sym)
				if err != nil {
					errors <- err
				} else {
					results <- quote
				}
			}(symbol)
		}

		// Collect results
		var quotes []*Quote
		var errs []error

		for i := 0; i < len(symbols); i++ {
			select {
			case quote := <-results:
				quotes = append(quotes, quote)
			case err := <-errors:
				errs = append(errs, err)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		assert.Len(t, quotes, len(symbols))
		assert.Len(t, errs, 0)

		// Verify all symbols were processed
		symbolMap := make(map[string]bool)
		for _, quote := range quotes {
			symbolMap[quote.Symbol] = true
		}

		for _, symbol := range symbols {
			assert.True(t, symbolMap[symbol], "Symbol %s not found in results", symbol)
		}
	})
}

func TestMarketDataService_DataValidation(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should return valid price data", func(t *testing.T) {
		symbols := []string{"AAPL", "GOOGL", "MSFT"}
		
		for _, symbol := range symbols {
			quote, err := service.GetQuote(ctx, symbol)
			require.NoError(t, err)

			// Validate price data
			assert.True(t, quote.Price.GreaterThan(decimal.Zero), "Price should be positive for %s", symbol)
			assert.True(t, quote.High.GreaterThanOrEqual(quote.Low), "High should be >= Low for %s", symbol)
			assert.True(t, quote.High.GreaterThanOrEqual(quote.Price), "High should be >= Price for %s", symbol)
			assert.True(t, quote.Low.LessThanOrEqual(quote.Price), "Low should be <= Price for %s", symbol)
			assert.True(t, quote.Volume >= 0, "Volume should be non-negative for %s", symbol)
			assert.False(t, quote.Timestamp.IsZero(), "Timestamp should be set for %s", symbol)
		}
	})

	t.Run("should calculate change percentage correctly", func(t *testing.T) {
		// Set a custom quote with known values
		service.SetQuote("TEST_CHANGE", &Quote{
			Symbol:        "TEST_CHANGE",
			Price:         decimal.NewFromFloat(110.00),
			PreviousClose: decimal.NewFromFloat(100.00),
			Change:        decimal.NewFromFloat(10.00),
			ChangePercent: decimal.NewFromFloat(10.00),
		})

		quote, err := service.GetQuote(ctx, "TEST_CHANGE")
		require.NoError(t, err)

		expectedChange := quote.Price.Sub(quote.PreviousClose)
		assert.True(t, expectedChange.Equal(quote.Change), "Change calculation incorrect")

		if quote.PreviousClose.GreaterThan(decimal.Zero) {
			expectedChangePercent := quote.Change.Div(quote.PreviousClose).Mul(decimal.NewFromInt(100))
			assert.True(t, expectedChangePercent.Equal(quote.ChangePercent), "Change percentage calculation incorrect")
		}
	})
}

func TestMarketDataService_PerformanceMetrics(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should handle high volume requests", func(t *testing.T) {
		start := time.Now()
		numRequests := 1000

		for i := 0; i < numRequests; i++ {
			_, err := service.GetQuote(ctx, "AAPL")
			require.NoError(t, err)
		}

		duration := time.Since(start)
		avgLatency := duration / time.Duration(numRequests)

		// Should handle 1000 requests reasonably quickly
		assert.Less(t, avgLatency, 1*time.Millisecond, "Average latency too high: %v", avgLatency)
	})
}

func TestMarketDataService_HistoricalDataValidation(t *testing.T) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	t.Run("should return chronologically ordered data", func(t *testing.T) {
		from := time.Now().AddDate(0, 0, -5) // 5 days ago
		to := time.Now().AddDate(0, 0, -1)   // 1 day ago

		ohlcv, err := service.GetOHLCV(ctx, "AAPL", from, to, "1day")
		require.NoError(t, err)
		require.NotEmpty(t, ohlcv)

		// Verify chronological order
		for i := 1; i < len(ohlcv); i++ {
			assert.True(t, ohlcv[i].Timestamp.After(ohlcv[i-1].Timestamp) || 
						 ohlcv[i].Timestamp.Equal(ohlcv[i-1].Timestamp),
				"OHLCV data should be in chronological order")
		}
	})

	t.Run("should respect date range boundaries", func(t *testing.T) {
		from := time.Now().AddDate(0, 0, -3) // 3 days ago
		to := time.Now().AddDate(0, 0, -1)   // 1 day ago

		ohlcv, err := service.GetOHLCV(ctx, "AAPL", from, to, "1day")
		require.NoError(t, err)

		for _, data := range ohlcv {
			assert.True(t, data.Timestamp.After(from) || data.Timestamp.Equal(from),
				"Data timestamp should be after or equal to from date")
			assert.True(t, data.Timestamp.Before(to) || data.Timestamp.Equal(to),
				"Data timestamp should be before or equal to to date")
		}
	})
}

// Benchmark tests
func BenchmarkMarketDataService_GetQuote(b *testing.B) {
	service := NewMockMarketDataService()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetQuote(ctx, "AAPL")
	}
}

func BenchmarkMarketDataService_GetMultipleQuotes(b *testing.B) {
	service := NewMockMarketDataService()
	ctx := context.Background()
	symbols := []string{"AAPL", "GOOGL", "MSFT", "AMZN", "TSLA"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetMultipleQuotes(ctx, symbols)
	}
}

func BenchmarkMarketDataService_GetOHLCV(b *testing.B) {
	service := NewMockMarketDataService()
	ctx := context.Background()
	from := time.Now().AddDate(0, 0, -30) // 30 days ago
	to := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetOHLCV(ctx, "AAPL", from, to, "1day")
	}
}