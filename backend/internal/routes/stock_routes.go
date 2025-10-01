package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
)

// SetupStockRoutes sets up all stock-related routes
func SetupStockRoutes(app fiber.Router, stockHandler *handlers.StockHandler) {
	stocks := app.Group("/stocks")

	// Stock CRUD operations
	stocks.Post("/", stockHandler.CreateStock)
	stocks.Get("/", stockHandler.GetStocks)
	stocks.Get("/:id", stockHandler.GetStock)
	stocks.Put("/:id", stockHandler.UpdateStock)
	stocks.Delete("/:id", stockHandler.DeleteStock)

	// Stock lookup by ticker
	stocks.Get("/ticker/:ticker", stockHandler.GetStockByTicker)

	// Signal management
	stocks.Put("/:id/signal", stockHandler.UpdateStockSignal)
	stocks.Get("/:id/signals", stockHandler.GetStockSignalHistory)

	// Strategy assignment management
	stocks.Post("/:id/strategies/:strategyId", stockHandler.AddStockToStrategy)
	stocks.Delete("/:id/strategies/:strategyId", stockHandler.RemoveStockFromStrategy)
}