package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"portfolio-app/internal/models"
	"portfolio-app/internal/repositories"
	"portfolio-app/internal/services"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(authService *services.AuthService, userRepo repositories.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token required",
			})
		}

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Check if session exists in Redis
		session, err := authService.GetSession(c.Context(), claims.UserID)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Session not found or expired",
			})
		}

		// Get user from database to ensure they still exist
		user, err := userRepo.GetByID(c.Context(), claims.UserID)
		if err != nil || user == nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		// Set user information in context
		c.Locals("userID", claims.UserID)
		c.Locals("user", user)
		c.Locals("session", session)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware that optionally authenticates users
// If a valid token is provided, user info is set in context
// If no token or invalid token, the request continues without user info
func OptionalAuthMiddleware(authService *services.AuthService, userRepo repositories.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next()
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return c.Next()
		}

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			return c.Next()
		}

		// Check if session exists in Redis
		session, err := authService.GetSession(c.Context(), claims.UserID)
		if err != nil {
			return c.Next()
		}

		// Get user from database
		user, err := userRepo.GetByID(c.Context(), claims.UserID)
		if err != nil || user == nil {
			return c.Next()
		}

		// Set user information in context
		c.Locals("userID", claims.UserID)
		c.Locals("user", user)
		c.Locals("session", session)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// GetUserFromContext extracts user information from fiber context
func GetUserFromContext(c *fiber.Ctx) (*models.User, bool) {
	user, ok := c.Locals("user").(*models.User)
	return user, ok
}

// GetUserIDFromContext extracts user ID from fiber context
func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, bool) {
	userID, ok := c.Locals("userID").(uuid.UUID)
	return userID, ok
}

// GetSessionFromContext extracts session data from fiber context
func GetSessionFromContext(c *fiber.Ctx) (*models.SessionData, bool) {
	session, ok := c.Locals("session").(*models.SessionData)
	return session, ok
}

// GetClaimsFromContext extracts JWT claims from fiber context
func GetClaimsFromContext(c *fiber.Ctx) (*models.JWTClaims, bool) {
	claims, ok := c.Locals("claims").(*models.JWTClaims)
	return claims, ok
}

// RequireAuth is a helper function that can be used in handlers to ensure authentication
func RequireAuth(c *fiber.Ctx) (*models.User, error) {
	user, ok := GetUserFromContext(c)
	if !ok {
		return nil, c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}
	return user, nil
}