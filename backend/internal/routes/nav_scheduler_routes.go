package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
)

// SetupNAVSchedulerRoutes sets up NAV scheduler routes
func SetupNAVSchedulerRoutes(router fiber.Router, handler *handlers.NAVSchedulerHandler) {
	navGroup := router.Group("/nav-scheduler")
	
	// Get scheduler status
	navGroup.Get("/status", handler.GetStatus)
	
	// Start scheduler
	navGroup.Post("/start", handler.Start)
	
	// Stop scheduler
	navGroup.Post("/stop", handler.Stop)
	
	// Force update all portfolios
	navGroup.Post("/update", handler.ForceUpdate)
	
	// Update specific portfolio
	navGroup.Post("/update/:id", handler.UpdateSinglePortfolio)
}