package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
)

// SetupMarketDataRoutes sets up all market data related routes
func SetupMarketDataRoutes(app fiber.Router, marketDataHandler *handlers.MarketDataHandler) {
	// Market data endpoints
	quotes := app.Group("/quotes")
	quotes.Get("/:ticker", marketDataHandler.GetQuote)
	quotes.Get("/", marketDataHandler.GetMultipleQuotes)

	// Historical data endpoint
	app.Get("/ohlcv/:ticker", marketDataHandler.GetOHLCV)

	// TradingView DataFeed compatible endpoints
	tradingview := app.Group("/tradingview")
	tradingview.Get("/config", marketDataHandler.TradingViewConfig)
	tradingview.Get("/symbols", marketDataHandler.TradingViewSymbolSearch)
	tradingview.Get("/history", marketDataHandler.TradingViewHistory)
	tradingview.Get("/quotes", marketDataHandler.TradingViewQuotes)
	tradingview.Get("/symbol_info", marketDataHandler.TradingViewSymbolInfo)
	tradingview.Get("/marks", marketDataHandler.TradingViewMarks)
	tradingview.Get("/time", marketDataHandler.TradingViewTime)
}