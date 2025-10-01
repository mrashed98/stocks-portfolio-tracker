package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
	"portfolio-app/internal/middleware"
	"portfolio-app/internal/repositories"
	"portfolio-app/internal/services"
)

// SetupNAVSchedulerRoutes sets up NAV scheduler routes
func SetupNAVSchedulerRoutes(router fiber.Router, handler *handlers.NAVSchedulerHandler, authService *services.AuthService, userRepo repositories.UserRepository) {
	navGroup := router.Group("/nav-scheduler")
	
	// Apply authentication middleware and rate limiting to all NAV scheduler routes
	protected := navGroup.Group("", middleware.AuthMiddleware(authService, userRepo), middleware.RateLimitMiddleware())
	
	// Get scheduler status
	protected.Get("/status", handler.GetStatus)
	
	// Start scheduler
	protected.Post("/start", handler.Start)
	
	// Stop scheduler
	protected.Post("/stop", handler.Stop)
	
	// Force update all portfolios
	protected.Post("/update", handler.ForceUpdate)
	
	// Update specific portfolio
	protected.Post("/update/:id", handler.UpdateSinglePortfolio)
}