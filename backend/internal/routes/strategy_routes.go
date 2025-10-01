package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
)

// SetupStrategyRoutes sets up all strategy-related routes
func SetupStrategyRoutes(app fiber.Router, strategyHandler *handlers.StrategyHandler) {
	strategies := app.Group("/strategies")

	// Strategy CRUD operations
	strategies.Post("/", strategyHandler.CreateStrategy)
	strategies.Get("/", strategyHandler.GetStrategies)
	strategies.Get("/:id", strategyHandler.GetStrategy)
	strategies.Put("/:id", strategyHandler.UpdateStrategy)
	strategies.Delete("/:id", strategyHandler.DeleteStrategy)

	// Strategy weight management
	strategies.Put("/:id/weight", strategyHandler.UpdateStrategyWeight)

	// Stock assignment and eligibility management
	strategies.Put("/:id/stocks/:stockId", strategyHandler.UpdateStockEligibility)
}