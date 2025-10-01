package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
	"portfolio-app/internal/middleware"
	"portfolio-app/internal/repositories"
	"portfolio-app/internal/services"
)

// SetupStockRoutes sets up all stock-related routes
func SetupStockRoutes(app fiber.Router, stockHandler *handlers.StockHandler, authService *services.AuthService, userRepo repositories.UserRepository) {
	stocks := app.Group("/stocks")

	// Apply authentication middleware and rate limiting to all stock routes
	protected := stocks.Group("", middleware.AuthMiddleware(authService, userRepo), middleware.RateLimitMiddleware())

	// Stock CRUD operations
	protected.Post("/", stockHandler.CreateStock)
	protected.Get("/", stockHandler.GetStocks)
	protected.Get("/:id", stockHandler.GetStock)
	protected.Put("/:id", stockHandler.UpdateStock)
	protected.Delete("/:id", stockHandler.DeleteStock)

	// Stock lookup by ticker
	protected.Get("/ticker/:ticker", stockHandler.GetStockByTicker)

	// Signal management
	protected.Put("/:id/signal", stockHandler.UpdateStockSignal)
	protected.Get("/:id/signals", stockHandler.GetStockSignalHistory)

	// Strategy assignment management
	protected.Post("/:id/strategies/:strategyId", stockHandler.AddStockToStrategy)
	protected.Delete("/:id/strategies/:strategyId", stockHandler.RemoveStockFromStrategy)
}