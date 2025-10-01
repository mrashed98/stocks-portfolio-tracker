package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
	"portfolio-app/internal/middleware"
	"portfolio-app/internal/repositories"
	"portfolio-app/internal/services"
)

// SetupStrategyRoutes sets up all strategy-related routes
func SetupStrategyRoutes(app fiber.Router, strategyHandler *handlers.StrategyHandler, authService *services.AuthService, userRepo repositories.UserRepository) {
	strategies := app.Group("/strategies")

	// Apply authentication middleware and rate limiting to all strategy routes
	protected := strategies.Group("", middleware.AuthMiddleware(authService, userRepo), middleware.RateLimitMiddleware())

	// Strategy CRUD operations
	protected.Post("/", strategyHandler.CreateStrategy)
	protected.Get("/", strategyHandler.GetStrategies)
	protected.Get("/:id", strategyHandler.GetStrategy)
	protected.Put("/:id", strategyHandler.UpdateStrategy)
	protected.Delete("/:id", strategyHandler.DeleteStrategy)

	// Strategy weight management
	protected.Put("/:id/weight", strategyHandler.UpdateStrategyWeight)

	// Stock assignment and eligibility management
	protected.Put("/:id/stocks/:stockId", strategyHandler.UpdateStockEligibility)
}