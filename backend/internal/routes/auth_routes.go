package routes

import (
	"github.com/gofiber/fiber/v2"
	"portfolio-app/internal/handlers"
	"portfolio-app/internal/middleware"
	"portfolio-app/internal/repositories"
	"portfolio-app/internal/services"
)

// SetupAuthRoutes sets up authentication-related routes
func SetupAuthRoutes(router fiber.Router, authHandler *handlers.AuthHandler, authService *services.AuthService, userRepo repositories.UserRepository) {
	auth := router.Group("/auth")

	// Public routes (no authentication required)
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	// Protected routes (authentication required)
	protected := auth.Group("", middleware.AuthMiddleware(authService, userRepo))
	protected.Post("/logout", authHandler.Logout)
	protected.Get("/profile", authHandler.GetProfile)
	protected.Put("/profile", authHandler.UpdateProfile)
}