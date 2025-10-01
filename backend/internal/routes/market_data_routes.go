package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
	"portfolio-app/internal/middleware"
	"portfolio-app/internal/repositories"
	"portfolio-app/internal/services"
)

// SetupMarketDataRoutes sets up all market data related routes
func SetupMarketDataRoutes(app fiber.Router, marketDataHandler *handlers.MarketDataHandler, authService *services.AuthService, userRepo repositories.UserRepository) {
	// Protected market data endpoints (require authentication)
	quotes := app.Group("/quotes", middleware.AuthMiddleware(authService, userRepo), middleware.RateLimitMiddleware())
	quotes.Get("/:ticker", marketDataHandler.GetQuote)
	quotes.Get("/", marketDataHandler.GetMultipleQuotes)

	// Protected historical data endpoint
	protected := app.Group("", middleware.AuthMiddleware(authService, userRepo), middleware.RateLimitMiddleware())
	protected.Get("/ohlcv/:ticker", marketDataHandler.GetOHLCV)

	// TradingView DataFeed compatible endpoints (optional auth for chart functionality)
	// These use optional auth middleware so charts can work for both authenticated and unauthenticated users
	tradingview := app.Group("/tradingview", middleware.OptionalAuthMiddleware(authService, userRepo), middleware.RateLimitMiddleware())
	tradingview.Get("/config", marketDataHandler.TradingViewConfig)
	tradingview.Get("/symbols", marketDataHandler.TradingViewSymbolSearch)
	tradingview.Get("/history", marketDataHandler.TradingViewHistory)
	tradingview.Get("/quotes", marketDataHandler.TradingViewQuotes)
	tradingview.Get("/symbol_info", marketDataHandler.TradingViewSymbolInfo)
	tradingview.Get("/marks", marketDataHandler.TradingViewMarks)
	tradingview.Get("/time", marketDataHandler.TradingViewTime)
}