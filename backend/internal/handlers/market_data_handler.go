package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/services"
)

// MarketDataHandler handles HTTP requests for market data operations
type MarketDataHandler struct {
	marketDataService services.MarketDataService
}

// NewMarketDataHandler creates a new market data handler
func NewMarketDataHandler(marketDataService services.MarketDataService) *MarketDataHandler {
	return &MarketDataHandler{
		marketDataService: marketDataService,
	}
}

// GetQuote handles GET /quotes/:ticker
func (h *MarketDataHandler) GetQuote(c *fiber.Ctx) error {
	ticker := c.Params("ticker")
	if ticker == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ticker symbol is required",
		})
	}
	ticker = strings.ToUpper(ticker)

	quote, err := h.marketDataService.GetQuote(c.Context(), ticker)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get quote",
			"details": err.Error(),
		})
	}

	return c.JSON(quote)
}

// GetMultipleQuotes handles GET /quotes with batch processing
func (h *MarketDataHandler) GetMultipleQuotes(c *fiber.Ctx) error {
	symbolsParam := c.Query("symbols")
	if symbolsParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "symbols parameter is required (comma-separated list)",
		})
	}

	// Parse symbols from comma-separated string
	symbols := strings.Split(symbolsParam, ",")
	for i, symbol := range symbols {
		symbols[i] = strings.ToUpper(strings.TrimSpace(symbol))
	}

	if len(symbols) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one symbol is required",
		})
	}

	// Limit batch size to prevent abuse
	if len(symbols) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Maximum 50 symbols allowed per request",
		})
	}

	quotes, err := h.marketDataService.GetMultipleQuotes(c.Context(), symbols)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get quotes",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"quotes": quotes,
		"meta": fiber.Map{
			"count":    len(quotes),
			"symbols":  symbols,
			"requested": len(symbols),
		},
	})
}

// GetOHLCV handles GET /ohlcv/:ticker for historical data
func (h *MarketDataHandler) GetOHLCV(c *fiber.Ctx) error {
	ticker := c.Params("ticker")
	if ticker == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ticker symbol is required",
		})
	}
	ticker = strings.ToUpper(ticker)

	// Parse query parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")
	interval := c.Query("interval", "1day")

	// Default date range (30 days)
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	var err error
	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'from' date format. Use YYYY-MM-DD",
			})
		}
	}

	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'to' date format. Use YYYY-MM-DD",
			})
		}
	}

	// Validate date range
	if from.After(to) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "'from' date must be before 'to' date",
		})
	}

	// Validate interval
	validIntervals := []string{"1min", "5min", "15min", "30min", "1h", "4h", "1day", "1week", "1month"}
	isValidInterval := false
	for _, valid := range validIntervals {
		if interval == valid {
			isValidInterval = true
			break
		}
	}
	if !isValidInterval {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid interval. Valid values: " + strings.Join(validIntervals, ", "),
		})
	}

	ohlcv, err := h.marketDataService.GetOHLCV(c.Context(), ticker, from, to, interval)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get historical data",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"symbol":   ticker,
		"interval": interval,
		"from":     from.Format("2006-01-02"),
		"to":       to.Format("2006-01-02"),
		"data":     ohlcv,
		"meta": fiber.Map{
			"count": len(ohlcv),
		},
	})
}

// TradingViewSymbolSearch handles GET /symbols for TradingView DataFeed
func (h *MarketDataHandler) TradingViewSymbolSearch(c *fiber.Ctx) error {
	query := c.Query("query", "")
	typeParam := c.Query("type", "")
	exchange := c.Query("exchange", "")
	limitStr := c.Query("limit", "30")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	// Mock symbol search results for TradingView DataFeed
	// In a real implementation, this would search a symbol database
	symbols := []fiber.Map{}

	// Add some common symbols if query matches
	commonSymbols := map[string]fiber.Map{
		"AAPL": {
			"symbol":      "AAPL",
			"full_name":   "NASDAQ:AAPL",
			"description": "Apple Inc.",
			"exchange":    "NASDAQ",
			"ticker":      "AAPL",
			"type":        "stock",
		},
		"GOOGL": {
			"symbol":      "GOOGL",
			"full_name":   "NASDAQ:GOOGL",
			"description": "Alphabet Inc. Class A",
			"exchange":    "NASDAQ",
			"ticker":      "GOOGL",
			"type":        "stock",
		},
		"MSFT": {
			"symbol":      "MSFT",
			"full_name":   "NASDAQ:MSFT",
			"description": "Microsoft Corporation",
			"exchange":    "NASDAQ",
			"ticker":      "MSFT",
			"type":        "stock",
		},
		"TSLA": {
			"symbol":      "TSLA",
			"full_name":   "NASDAQ:TSLA",
			"description": "Tesla, Inc.",
			"exchange":    "NASDAQ",
			"ticker":      "TSLA",
			"type":        "stock",
		},
		"NVDA": {
			"symbol":      "NVDA",
			"full_name":   "NASDAQ:NVDA",
			"description": "NVIDIA Corporation",
			"exchange":    "NASDAQ",
			"ticker":      "NVDA",
			"type":        "stock",
		},
	}

	queryUpper := strings.ToUpper(query)
	for symbol, data := range commonSymbols {
		if query == "" || strings.Contains(symbol, queryUpper) || strings.Contains(strings.ToUpper(data["description"].(string)), queryUpper) {
			// Filter by type if specified
			if typeParam != "" && data["type"] != typeParam {
				continue
			}
			// Filter by exchange if specified
			if exchange != "" && data["exchange"] != exchange {
				continue
			}
			symbols = append(symbols, data)
			if len(symbols) >= limit {
				break
			}
		}
	}

	return c.JSON(symbols)
}

// TradingViewHistory handles GET /history for TradingView DataFeed
func (h *MarketDataHandler) TradingViewHistory(c *fiber.Ctx) error {
	symbol := c.Query("symbol")
	resolution := c.Query("resolution", "D")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"s": "error",
			"errmsg": "symbol parameter is required",
		})
	}

	if fromStr == "" || toStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"s": "error",
			"errmsg": "from and to parameters are required",
		})
	}

	// Parse timestamps
	fromTimestamp, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"s": "error",
			"errmsg": "invalid from timestamp",
		})
	}

	toTimestamp, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"s": "error",
			"errmsg": "invalid to timestamp",
		})
	}

	from := time.Unix(fromTimestamp, 0)
	to := time.Unix(toTimestamp, 0)

	// Convert TradingView resolution to internal interval format
	interval := h.convertTradingViewResolution(resolution)

	// Extract ticker from symbol (remove exchange prefix if present)
	ticker := symbol
	if strings.Contains(symbol, ":") {
		parts := strings.Split(symbol, ":")
		ticker = parts[len(parts)-1]
	}

	ohlcv, err := h.marketDataService.GetOHLCV(c.Context(), ticker, from, to, interval)
	if err != nil {
		// Check if it's a "not supported" error from TradingView service
		if strings.Contains(err.Error(), "not supported") {
			return c.JSON(fiber.Map{
				"s": "no_data",
				"nextTime": to.Unix() + 86400, // Suggest checking again tomorrow
			})
		}
		return c.JSON(fiber.Map{
			"s": "error",
			"errmsg": "Failed to get historical data: " + err.Error(),
		})
	}

	if len(ohlcv) == 0 {
		return c.JSON(fiber.Map{
			"s": "no_data",
			"nextTime": to.Unix() + 86400, // Suggest checking again tomorrow
		})
	}

	// Convert to TradingView format
	timestamps := make([]int64, len(ohlcv))
	opens := make([]float64, len(ohlcv))
	highs := make([]float64, len(ohlcv))
	lows := make([]float64, len(ohlcv))
	closes := make([]float64, len(ohlcv))
	volumes := make([]int64, len(ohlcv))

	for i, bar := range ohlcv {
		timestamps[i] = bar.Timestamp.Unix()
		opens[i], _ = bar.Open.Float64()
		highs[i], _ = bar.High.Float64()
		lows[i], _ = bar.Low.Float64()
		closes[i], _ = bar.Close.Float64()
		volumes[i] = bar.Volume
	}

	return c.JSON(fiber.Map{
		"s": "ok",
		"t": timestamps,
		"o": opens,
		"h": highs,
		"l": lows,
		"c": closes,
		"v": volumes,
	})
}

// TradingViewConfig handles GET /config for TradingView DataFeed configuration
func (h *MarketDataHandler) TradingViewConfig(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"supports_search":                true,
		"supports_group_request":         true,
		"supports_marks":                 true,
		"supports_timescale_marks":       false,
		"supports_time":                  true,
		"supports_quotes":                true,
		"supports_symbol_info":           true,
		"currency_codes":                 []string{"USD"},
		"exchanges": []fiber.Map{
			{
				"value": "NASDAQ",
				"name":  "NASDAQ",
				"desc":  "NASDAQ Stock Exchange",
			},
			{
				"value": "NYSE",
				"name":  "NYSE",
				"desc":  "New York Stock Exchange",
			},
		},
		"symbols_types": []fiber.Map{
			{
				"name":  "All types",
				"value": "",
			},
			{
				"name":  "Stock",
				"value": "stock",
			},
		},
		"supported_resolutions": []string{"1", "5", "15", "30", "60", "240", "D", "W", "M"},
	})
}

// TradingViewQuotes handles GET /quotes for TradingView DataFeed real-time quotes
func (h *MarketDataHandler) TradingViewQuotes(c *fiber.Ctx) error {
	symbols := c.Query("symbols")
	if symbols == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"s": "error",
			"errmsg": "symbols parameter is required",
		})
	}

	// Parse symbols (comma-separated)
	symbolList := strings.Split(symbols, ",")
	for i, symbol := range symbolList {
		symbolList[i] = strings.TrimSpace(symbol)
	}

	// Get quotes for all symbols
	quotes, err := h.marketDataService.GetMultipleQuotes(c.Context(), symbolList)
	if err != nil {
		return c.JSON(fiber.Map{
			"s": "error",
			"errmsg": "Failed to get quotes: " + err.Error(),
		})
	}

	// Convert to TradingView format
	result := make(map[string]fiber.Map)
	for symbol, quote := range quotes {
		if quote != nil {
			price, _ := quote.Price.Float64()
			change, _ := quote.Change.Float64()
			changePercent, _ := quote.ChangePercent.Float64()
			
			result[symbol] = fiber.Map{
				"n":  symbol,                    // name
				"s":  "ok",                      // status
				"v":  fiber.Map{                 // values
					"lp":  price,                // last price
					"ch":  change,               // change
					"chp": changePercent,        // change percent
					"vol": quote.Volume,         // volume
				},
			}
		} else {
			result[symbol] = fiber.Map{
				"n": symbol,
				"s": "no_data",
			}
		}
	}

	return c.JSON(result)
}

// TradingViewSymbolInfo handles GET /symbol_info for TradingView DataFeed
func (h *MarketDataHandler) TradingViewSymbolInfo(c *fiber.Ctx) error {
	group := c.Query("group")
	if group == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"s": "error",
			"errmsg": "group parameter is required",
		})
	}

	// Mock symbol info - in production, this would come from a symbol database
	symbols := []fiber.Map{
		{
			"symbol":      "AAPL",
			"ticker":      "AAPL",
			"name":        "Apple Inc.",
			"full_name":   "NASDAQ:AAPL",
			"description": "Apple Inc.",
			"type":        "stock",
			"session":     "0930-1600",
			"exchange":    "NASDAQ",
			"listed_exchange": "NASDAQ",
			"timezone":    "America/New_York",
			"minmov":      1,
			"minmov2":     0,
			"pointvalue":  1,
			"pricescale":  100,
			"has_intraday": true,
			"has_no_volume": false,
			"volume_precision": 0,
			"data_status": "streaming",
		},
		{
			"symbol":      "GOOGL",
			"ticker":      "GOOGL",
			"name":        "Alphabet Inc. Class A",
			"full_name":   "NASDAQ:GOOGL",
			"description": "Alphabet Inc. Class A",
			"type":        "stock",
			"session":     "0930-1600",
			"exchange":    "NASDAQ",
			"listed_exchange": "NASDAQ",
			"timezone":    "America/New_York",
			"minmov":      1,
			"minmov2":     0,
			"pointvalue":  1,
			"pricescale":  100,
			"has_intraday": true,
			"has_no_volume": false,
			"volume_precision": 0,
			"data_status": "streaming",
		},
		{
			"symbol":      "MSFT",
			"ticker":      "MSFT",
			"name":        "Microsoft Corporation",
			"full_name":   "NASDAQ:MSFT",
			"description": "Microsoft Corporation",
			"type":        "stock",
			"session":     "0930-1600",
			"exchange":    "NASDAQ",
			"listed_exchange": "NASDAQ",
			"timezone":    "America/New_York",
			"minmov":      1,
			"minmov2":     0,
			"pointvalue":  1,
			"pricescale":  100,
			"has_intraday": true,
			"has_no_volume": false,
			"volume_precision": 0,
			"data_status": "streaming",
		},
	}

	return c.JSON(fiber.Map{
		"s": "ok",
		"d": symbols,
	})
}

// TradingViewMarks handles GET /marks for TradingView DataFeed (optional)
func (h *MarketDataHandler) TradingViewMarks(c *fiber.Ctx) error {
	// Return empty marks - this is optional for basic functionality
	return c.JSON(fiber.Map{
		"s": "ok",
		"d": []interface{}{},
	})
}

// TradingViewTime handles GET /time for TradingView DataFeed server time
func (h *MarketDataHandler) TradingViewTime(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"s": "ok",
		"t": time.Now().Unix(),
	})
}

// convertTradingViewResolution converts TradingView resolution to internal interval format
func (h *MarketDataHandler) convertTradingViewResolution(resolution string) string {
	switch resolution {
	case "1":
		return "1min"
	case "5":
		return "5min"
	case "15":
		return "15min"
	case "30":
		return "30min"
	case "60":
		return "1h"
	case "240":
		return "4h"
	case "D":
		return "1day"
	case "W":
		return "1week"
	case "M":
		return "1month"
	default:
		return "1day"
	}
}