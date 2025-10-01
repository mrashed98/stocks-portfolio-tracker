package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"portfolio-app/internal/models"
	"portfolio-app/internal/services"
)

// StrategyHandler handles HTTP requests for strategy operations
type StrategyHandler struct {
	strategyService services.StrategyService
}

// NewStrategyHandler creates a new strategy handler
func NewStrategyHandler(strategyService services.StrategyService) *StrategyHandler {
	return &StrategyHandler{
		strategyService: strategyService,
	}
}

// CreateStrategy handles POST /strategies
func (h *StrategyHandler) CreateStrategy(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	// For now, using a placeholder user ID
	userID := uuid.New()

	var req models.CreateStrategyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate request
	if err := models.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}

	strategy, err := h.strategyService.CreateStrategy(c.Context(), &req, userID)
	if err != nil {
		// Check if it's a validation error (weight constraint violation)
		if validationErr, ok := err.(*models.ValidationError); ok {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "Validation failed",
				"details": validationErr.Error(),
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create strategy",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(strategy.ToResponse())
}

// GetStrategies handles GET /strategies
func (h *StrategyHandler) GetStrategies(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	userID := uuid.New()

	strategies, err := h.strategyService.GetUserStrategies(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get strategies",
			"details": err.Error(),
		})
	}

	// Convert to response format
	responses := make([]*models.StrategyResponse, len(strategies))
	for i, strategy := range strategies {
		responses[i] = strategy.ToResponse()
	}

	return c.JSON(fiber.Map{
		"strategies": responses,
	})
}

// GetStrategy handles GET /strategies/:id
func (h *StrategyHandler) GetStrategy(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	userID := uuid.New()

	strategyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	strategy, err := h.strategyService.GetStrategy(c.Context(), strategyID, userID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Strategy not found",
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get strategy",
			"details": err.Error(),
		})
	}

	return c.JSON(strategy.ToResponse())
}

// UpdateStrategy handles PUT /strategies/:id
func (h *StrategyHandler) UpdateStrategy(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	userID := uuid.New()

	strategyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	var req models.UpdateStrategyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate request
	if err := models.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}

	strategy, err := h.strategyService.UpdateStrategy(c.Context(), strategyID, &req, userID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Strategy not found",
			})
		}
		
		// Check if it's a validation error (weight constraint violation)
		if validationErr, ok := err.(*models.ValidationError); ok {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "Validation failed",
				"details": validationErr.Error(),
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update strategy",
			"details": err.Error(),
		})
	}

	return c.JSON(strategy.ToResponse())
}

// DeleteStrategy handles DELETE /strategies/:id
func (h *StrategyHandler) DeleteStrategy(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	userID := uuid.New()

	strategyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	err = h.strategyService.DeleteStrategy(c.Context(), strategyID, userID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Strategy not found",
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete strategy",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// UpdateStrategyWeight handles PUT /strategies/:id/weight
func (h *StrategyHandler) UpdateStrategyWeight(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	userID := uuid.New()

	strategyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	var req struct {
		WeightValue string `json:"weight_value" validate:"required"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Parse weight value as decimal
	weightValue, err := strconv.ParseFloat(req.WeightValue, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid weight value",
		})
	}

	if weightValue <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Weight value must be greater than 0",
		})
	}

	// Create update request with only weight value
	weightDecimal := decimal.NewFromFloat(weightValue)
	updateReq := &models.UpdateStrategyRequest{
		WeightValue: &weightDecimal,
	}

	strategy, err := h.strategyService.UpdateStrategy(c.Context(), strategyID, updateReq, userID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Strategy not found",
			})
		}
		
		// Check if it's a validation error (weight constraint violation)
		if validationErr, ok := err.(*models.ValidationError); ok {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "Validation failed",
				"details": validationErr.Error(),
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update strategy weight",
			"details": err.Error(),
		})
	}

	return c.JSON(strategy.ToResponse())
}

// UpdateStockEligibility handles PUT /strategies/:id/stocks/:stockId
func (h *StrategyHandler) UpdateStockEligibility(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	userID := uuid.New()

	strategyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	stockID, err := uuid.Parse(c.Params("stockId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stock ID",
		})
	}

	var req models.UpdateStockEligibilityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate request
	if err := models.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}

	err = h.strategyService.UpdateStockEligibility(c.Context(), strategyID, stockID, req.Eligible, userID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Strategy or stock not found",
			})
		}
		
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update stock eligibility",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stock eligibility updated successfully",
	})
}