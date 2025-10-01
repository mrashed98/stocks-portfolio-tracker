package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"portfolio-app/internal/models"
	"portfolio-app/internal/services"
)

// PortfolioHandler handles portfolio-related HTTP requests
type PortfolioHandler struct {
	portfolioService services.PortfolioServiceInterface
	cache           *services.AllocationPreviewCache
}

// NewPortfolioHandler creates a new portfolio handler
func NewPortfolioHandler(portfolioService services.PortfolioServiceInterface) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: portfolioService,
		cache:           services.NewAllocationPreviewCache(),
	}
}

// GenerateAllocationPreview handles POST /api/portfolios/preview
func (h *PortfolioHandler) GenerateAllocationPreview(c *fiber.Ctx) error {
	var req models.AllocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Check cache first
	cacheKey := services.GenerateCacheKey(&req)
	if cachedPreview, found := h.cache.Get(cacheKey); found {
		return c.JSON(fiber.Map{
			"data": cachedPreview,
			"cached": true,
		})
	}

	// Generate new preview
	preview, err := h.portfolioService.GenerateAllocationPreview(c.Context(), &req)
	if err != nil {
		// Check if it's an allocation error for better error handling
		if allocErr := services.GetAllocationError(err); allocErr != nil {
			return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": allocErr.Message,
				"type": allocErr.Type,
				"details": allocErr.Details,
			})
		}

		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate allocation preview",
			"details": err.Error(),
		})
	}

	// Cache the result for 5 minutes (market data changes frequently)
	h.cache.Set(cacheKey, preview, 5*time.Minute)

	return c.JSON(fiber.Map{
		"data": preview,
		"cached": false,
	})
}

// GenerateAllocationPreviewWithExclusions handles POST /api/portfolios/preview/exclude
func (h *PortfolioHandler) GenerateAllocationPreviewWithExclusions(c *fiber.Ctx) error {
	var reqBody struct {
		AllocationRequest models.AllocationRequest `json:"allocation_request"`
		ExcludedStocks   []uuid.UUID             `json:"excluded_stocks"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Generate cache key including exclusions
	cacheKey := services.GenerateCacheKey(&reqBody.AllocationRequest)
	if len(reqBody.ExcludedStocks) > 0 {
		cacheKey += fmt.Sprintf("_excl_%v", reqBody.ExcludedStocks)
	}

	// Check cache first
	if cachedPreview, found := h.cache.Get(cacheKey); found {
		return c.JSON(fiber.Map{
			"data": cachedPreview,
			"cached": true,
		})
	}

	// Generate new preview with exclusions
	preview, err := h.portfolioService.GenerateAllocationPreviewWithExclusions(
		c.Context(), 
		&reqBody.AllocationRequest, 
		reqBody.ExcludedStocks,
	)
	if err != nil {
		// Check if it's an allocation error for better error handling
		if allocErr := services.GetAllocationError(err); allocErr != nil {
			return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": allocErr.Message,
				"type": allocErr.Type,
				"details": allocErr.Details,
			})
		}

		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate allocation preview with exclusions",
			"details": err.Error(),
		})
	}

	// Cache the result for 5 minutes
	h.cache.Set(cacheKey, preview, 5*time.Minute)

	return c.JSON(fiber.Map{
		"data": preview,
		"cached": false,
	})
}

// ValidateAllocationRequest handles POST /api/portfolios/validate
func (h *PortfolioHandler) ValidateAllocationRequest(c *fiber.Ctx) error {
	var req models.AllocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate the request
	if err := h.portfolioService.ValidateAllocationRequest(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"valid": false,
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"valid": true,
		"message": "Allocation request is valid",
	})
}

// CreatePortfolio handles POST /api/portfolios
func (h *PortfolioHandler) CreatePortfolio(c *fiber.Ctx) error {
	var req models.CreatePortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
			"details": err.Error(),
		})
	}

	// Create portfolio
	portfolio, err := h.portfolioService.CreatePortfolio(c.Context(), &req, userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create portfolio",
			"details": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"data": portfolio.ToResponse(),
	})
}

// GetPortfolio handles GET /api/portfolios/:id
func (h *PortfolioHandler) GetPortfolio(c *fiber.Ctx) error {
	// Parse portfolio ID
	portfolioIDStr := c.Params("id")
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
			"details": err.Error(),
		})
	}

	// Get portfolio
	portfolio, err := h.portfolioService.GetPortfolio(c.Context(), portfolioID)
	if err != nil {
		if err.Error() == "portfolio not found" {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Portfolio not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get portfolio",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": portfolio.ToResponse(),
	})
}

// GetUserPortfolios handles GET /api/portfolios
func (h *PortfolioHandler) GetUserPortfolios(c *fiber.Ctx) error {
	// Get user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
			"details": err.Error(),
		})
	}

	// Get user portfolios
	portfolios, err := h.portfolioService.GetUserPortfolios(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get portfolios",
			"details": err.Error(),
		})
	}

	// Convert to response format
	responses := make([]*models.PortfolioResponse, len(portfolios))
	for i, portfolio := range portfolios {
		responses[i] = portfolio.ToResponse()
	}

	return c.JSON(fiber.Map{
		"data": responses,
	})
}

// UpdatePortfolio handles PUT /api/portfolios/:id
func (h *PortfolioHandler) UpdatePortfolio(c *fiber.Ctx) error {
	// Parse portfolio ID
	portfolioIDStr := c.Params("id")
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
			"details": err.Error(),
		})
	}

	var req models.UpdatePortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Update portfolio
	portfolio, err := h.portfolioService.UpdatePortfolio(c.Context(), portfolioID, &req)
	if err != nil {
		if err.Error() == "portfolio not found" {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Portfolio not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update portfolio",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": portfolio.ToResponse(),
	})
}

// DeletePortfolio handles DELETE /api/portfolios/:id
func (h *PortfolioHandler) DeletePortfolio(c *fiber.Ctx) error {
	// Parse portfolio ID
	portfolioIDStr := c.Params("id")
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
			"details": err.Error(),
		})
	}

	// Delete portfolio
	if err := h.portfolioService.DeletePortfolio(c.Context(), portfolioID); err != nil {
		if err.Error() == "portfolio not found" {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Portfolio not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete portfolio",
			"details": err.Error(),
		})
	}

	return c.Status(http.StatusNoContent).Send(nil)
}

// GetPortfolioHistory handles GET /api/portfolios/:id/history
func (h *PortfolioHandler) GetPortfolioHistory(c *fiber.Ctx) error {
	// Parse portfolio ID
	portfolioIDStr := c.Params("id")
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
			"details": err.Error(),
		})
	}

	// Parse date range from query parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'from' date format. Use RFC3339 format",
				"details": err.Error(),
			})
		}
	} else {
		// Default to 30 days ago
		from = time.Now().AddDate(0, 0, -30)
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'to' date format. Use RFC3339 format",
				"details": err.Error(),
			})
		}
	} else {
		// Default to now
		to = time.Now()
	}

	// Get portfolio history
	history, err := h.portfolioService.GetPortfolioHistory(c.Context(), portfolioID, from, to)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get portfolio history",
			"details": err.Error(),
		})
	}

	// Convert to response format
	responses := make([]*models.NAVHistoryResponse, len(history))
	for i, entry := range history {
		responses[i] = entry.ToResponse()
	}

	return c.JSON(fiber.Map{
		"data": responses,
		"from": from,
		"to":   to,
	})
}

// GetPortfolioPerformance handles GET /api/portfolios/:id/performance
func (h *PortfolioHandler) GetPortfolioPerformance(c *fiber.Ctx) error {
	// Parse portfolio ID
	portfolioIDStr := c.Params("id")
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
			"details": err.Error(),
		})
	}

	// Get performance metrics
	metrics, err := h.portfolioService.GetPortfolioPerformanceMetrics(c.Context(), portfolioID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get portfolio performance",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": metrics,
	})
}

// UpdatePortfolioNAV handles POST /api/portfolios/:id/nav/update
func (h *PortfolioHandler) UpdatePortfolioNAV(c *fiber.Ctx) error {
	// Parse portfolio ID
	portfolioIDStr := c.Params("id")
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
			"details": err.Error(),
		})
	}

	// Update NAV
	navHistory, err := h.portfolioService.UpdatePortfolioNAV(c.Context(), portfolioID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update portfolio NAV",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": navHistory.ToResponse(),
	})
}

// GenerateRebalancePreview handles POST /api/portfolios/:id/rebalance/preview
func (h *PortfolioHandler) GenerateRebalancePreview(c *fiber.Ctx) error {
	// Parse portfolio ID
	portfolioIDStr := c.Params("id")
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
			"details": err.Error(),
		})
	}

	var reqBody struct {
		NewTotalInvestment decimal.Decimal `json:"new_total_investment" validate:"required,gt=0"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Generate rebalance preview
	preview, err := h.portfolioService.GenerateRebalancePreview(c.Context(), portfolioID, reqBody.NewTotalInvestment)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate rebalance preview",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": preview,
	})
}

// RebalancePortfolio handles POST /api/portfolios/:id/rebalance
func (h *PortfolioHandler) RebalancePortfolio(c *fiber.Ctx) error {
	// Parse portfolio ID
	portfolioIDStr := c.Params("id")
	portfolioID, err := uuid.Parse(portfolioIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
			"details": err.Error(),
		})
	}

	var reqBody struct {
		NewTotalInvestment decimal.Decimal `json:"new_total_investment" validate:"required,gt=0"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Rebalance portfolio
	portfolio, err := h.portfolioService.RebalancePortfolio(c.Context(), portfolioID, reqBody.NewTotalInvestment)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to rebalance portfolio",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": portfolio.ToResponse(),
	})
}

// ClearAllocationCache handles DELETE /api/portfolios/cache
func (h *PortfolioHandler) ClearAllocationCache(c *fiber.Ctx) error {
	h.cache.Clear()
	
	return c.JSON(fiber.Map{
		"message": "Allocation preview cache cleared successfully",
	})
}

// GetAllocationConstraintsSuggestions handles GET /api/portfolios/constraints/suggestions
func (h *PortfolioHandler) GetAllocationConstraintsSuggestions(c *fiber.Ctx) error {
	// Parse query parameters
	totalInvestmentStr := c.Query("total_investment")
	if totalInvestmentStr == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "total_investment query parameter is required",
		})
	}

	totalInvestment, err := parseDecimal(totalInvestmentStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid total_investment value",
			"details": err.Error(),
		})
	}

	// Parse strategy IDs if provided
	strategyIDsStr := c.Query("strategy_ids")
	var strategyIDs []uuid.UUID
	if strategyIDsStr != "" {
		if err := json.Unmarshal([]byte(strategyIDsStr), &strategyIDs); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid strategy_ids format",
				"details": err.Error(),
			})
		}
	}

	// Create default constraints for suggestions
	maxAllocation := parseDecimalOrDefault(c.Query("max_allocation"), 25.0) // 25%
	minAllocation := parseDecimalOrDefault(c.Query("min_allocation"), 100.0) // $100

	// Generate suggestions (this would typically use the constraint validator)
	suggestions := []string{
		"Consider setting maximum allocation per stock between 10-30% for diversification",
		"Set minimum allocation amount based on your broker's minimum trade size",
		"Ensure minimum allocation allows for at least 5-10 different stocks",
		"Consider your total investment amount when setting constraints",
	}

	// Add specific suggestions based on parameters
	if totalInvestment.LessThan(parseDecimalOrDefault("", 1000.0)) {
		suggestions = append(suggestions, "With smaller investment amounts, consider higher maximum allocation percentages to reduce unallocated cash")
	}

	if len(strategyIDs) > 5 {
		suggestions = append(suggestions, "With many strategies, consider lower minimum allocation amounts to allow proper distribution")
	}

	return c.JSON(fiber.Map{
		"suggestions": suggestions,
		"recommended_constraints": map[string]interface{}{
			"max_allocation_per_stock": maxAllocation,
			"min_allocation_amount":    minAllocation,
		},
		"total_investment": totalInvestment,
	})
}

// Helper functions
func parseDecimal(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

func parseDecimalOrDefault(s string, defaultVal float64) decimal.Decimal {
	if s == "" {
		return decimal.NewFromFloat(defaultVal)
	}
	
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.NewFromFloat(defaultVal)
	}
	
	return d
}

// getUserIDFromContext extracts user ID from fiber context
// This assumes that authentication middleware sets the user ID in the context
func getUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	// Try to get user ID from context locals (set by auth middleware)
	userIDValue := c.Locals("user_id")
	if userIDValue == nil {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	
	// Handle different possible types
	switch v := userIDValue.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		return uuid.Parse(v)
	default:
		return uuid.Nil, fmt.Errorf("invalid user ID type in context")
	}
}