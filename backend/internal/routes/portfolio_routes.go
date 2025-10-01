package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
)

// SetupPortfolioRoutes sets up portfolio routes
func SetupPortfolioRoutes(router fiber.Router, handler *handlers.PortfolioHandler) {
	portfolioGroup := router.Group("/portfolios")
	
	// Allocation preview
	portfolioGroup.Post("/preview", handler.GenerateAllocationPreview)
	portfolioGroup.Post("/preview/exclusions", handler.GenerateAllocationPreviewWithExclusions)
	
	// Portfolio CRUD
	portfolioGroup.Post("/", handler.CreatePortfolio)
	portfolioGroup.Get("/", handler.GetUserPortfolios)
	portfolioGroup.Get("/:id", handler.GetPortfolio)
	portfolioGroup.Put("/:id", handler.UpdatePortfolio)
	portfolioGroup.Delete("/:id", handler.DeletePortfolio)
	
	// Portfolio performance
	portfolioGroup.Get("/:id/history", handler.GetPortfolioHistory)
	portfolioGroup.Get("/:id/performance", handler.GetPortfolioPerformance)
	portfolioGroup.Post("/:id/nav/update", handler.UpdatePortfolioNAV)
	
	// Portfolio rebalancing
	portfolioGroup.Post("/:id/rebalance/preview", handler.GenerateRebalancePreview)
	portfolioGroup.Post("/:id/rebalance", handler.RebalancePortfolio)
}