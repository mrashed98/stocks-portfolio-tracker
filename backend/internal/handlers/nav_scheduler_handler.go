package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// NAVSchedulerInterface defines the interface for NAV scheduler operations
type NAVSchedulerInterface interface {
	Start() error
	Stop() error
	IsRunning() bool
	GetMetrics() map[string]interface{}
	ForceUpdate() error
	UpdateSinglePortfolio(portfolioID uuid.UUID) error
}

// NAVSchedulerHandler handles NAV scheduler HTTP requests
type NAVSchedulerHandler struct {
	scheduler NAVSchedulerInterface
}

// NewNAVSchedulerHandler creates a new NAV scheduler handler
func NewNAVSchedulerHandler(scheduler NAVSchedulerInterface) *NAVSchedulerHandler {
	return &NAVSchedulerHandler{
		scheduler: scheduler,
	}
}

// GetStatus returns the current status of the NAV scheduler
func (h *NAVSchedulerHandler) GetStatus(c *fiber.Ctx) error {
	metrics := h.scheduler.GetMetrics()
	
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "NAV scheduler status retrieved",
		"data":    metrics,
	})
}

// Start starts the NAV scheduler
func (h *NAVSchedulerHandler) Start(c *fiber.Ctx) error {
	if err := h.scheduler.Start(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to start NAV scheduler",
			"error":   err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "NAV scheduler started successfully",
	})
}

// Stop stops the NAV scheduler
func (h *NAVSchedulerHandler) Stop(c *fiber.Ctx) error {
	if err := h.scheduler.Stop(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to stop NAV scheduler",
			"error":   err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "NAV scheduler stopped successfully",
	})
}

// ForceUpdate triggers an immediate NAV update for all portfolios
func (h *NAVSchedulerHandler) ForceUpdate(c *fiber.Ctx) error {
	if err := h.scheduler.ForceUpdate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to trigger NAV update",
			"error":   err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "NAV update triggered successfully",
	})
}

// UpdateSinglePortfolio updates NAV for a specific portfolio
func (h *NAVSchedulerHandler) UpdateSinglePortfolio(c *fiber.Ctx) error {
	portfolioIDStr := c.Params("id")
	if portfolioIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Portfolio ID is required",
		})
	}
	
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid portfolio ID format",
			"error":   err.Error(),
		})
	}
	
	if err := h.scheduler.UpdateSinglePortfolio(portfolioID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update portfolio NAV",
			"error":   err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Portfolio NAV updated successfully",
		"data": fiber.Map{
			"portfolio_id": portfolioID,
		},
	})
}