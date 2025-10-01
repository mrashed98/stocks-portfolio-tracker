package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	Max        int           // Maximum number of requests
	Expiration time.Duration // Time window for the limit
	Message    string        // Custom message when limit is exceeded
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:        100,                // 100 requests
		Expiration: 15 * time.Minute,   // per 15 minutes
		Message:    "Too many requests", // default message
	}
}

// AuthRateLimitConfig returns rate limiting configuration for auth endpoints
func AuthRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:        5,                 // 5 requests
		Expiration: 15 * time.Minute,  // per 15 minutes
		Message:    "Too many authentication attempts. Please try again later.",
	}
}

// CreateRateLimitMiddleware creates a rate limiting middleware with the given configuration
func CreateRateLimitMiddleware(config RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.Max,
		Expiration: config.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as the key for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": config.Message,
				"retry_after": fmt.Sprintf("%d seconds", int(config.Expiration.Seconds())),
			})
		},
	})
}

// RateLimitMiddleware creates a default rate limiting middleware
func RateLimitMiddleware() fiber.Handler {
	return CreateRateLimitMiddleware(DefaultRateLimitConfig())
}

// AuthRateLimitMiddleware creates a rate limiting middleware for authentication endpoints
func AuthRateLimitMiddleware() fiber.Handler {
	return CreateRateLimitMiddleware(AuthRateLimitConfig())
}