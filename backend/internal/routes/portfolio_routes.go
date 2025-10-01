package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
	"portfolio-app/internal/middleware"
	"portfolio-app/internal/repositories"
	"portfolio-app/internal/services"
)

// SetupPortfolioRoutes sets up portfolio routes
func SetupPortfolioRoutes(router fiber.Router, handler *handlers.PortfolioHandler, authService *services.AuthService, userRepo repositories.UserRepository) {
	portfolioGroup := router.Group("/portfolios")
	
	// Apply authentication middleware and rate limiting to all portfolio routes
	protected := portfolioGroup.Group("", middleware.AuthMiddleware(authService, userRepo), middleware.RateLimitMiddleware())
	
	// Allocation preview
	protected.Post("/preview", handler.GenerateAllocationPreview)
	protected.Post("/preview/exclusions", handler.GenerateAllocationPreviewWithExclusions)
	
	// Portfolio CRUD
	protected.Post("/", handler.CreatePortfolio)
	protected.Get("/", handler.GetUserPortfolios)
	protected.Get("/:id", handler.GetPortfolio)
	protected.Put("/:id", handler.UpdatePortfolio)
	protected.Delete("/:id", handler.DeletePortfolio)
	
	// Portfolio performance
	protected.Get("/:id/history", handler.GetPortfolioHistory)
	protected.Get("/:id/performance", handler.GetPortfolioPerformance)
	protected.Post("/:id/nav/update", handler.UpdatePortfolioNAV)
	
	// Portfolio rebalancing
	protected.Post("/:id/rebalance/preview", handler.GenerateRebalancePreview)
	protected.Post("/:id/rebalance", handler.RebalancePortfolio)
}