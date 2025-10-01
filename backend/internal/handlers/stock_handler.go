package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"portfolio-app/internal/models"
	"portfolio-app/internal/services"
)

// StockHandler handles HTTP requests for stock operations
type StockHandler struct {
	stockService services.StockService
}

// NewStockHandler creates a new stock handler
func NewStockHandler(stockService services.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
	}
}

// CreateStock handles POST /stocks
func (h *StockHandler) CreateStock(c *fiber.Ctx) error {
	var req models.CreateStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate request
	if err := models.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	stock, err := h.stockService.CreateStock(c.Context(), &req)
	if err != nil {
		// Check if it's a validation error
		if validationErr, ok := err.(*models.ValidationError); ok {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": validationErr.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to create stock",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(stock.ToResponse())
}

// GetStocks handles GET /stocks
func (h *StockHandler) GetStocks(c *fiber.Ctx) error {
	// Parse query parameters
	search := c.Query("search", "")
	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")
	includeSignals := c.Query("include_signals", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var stocks []*models.Stock
	if includeSignals == "true" {
		stocks, err = h.stockService.GetStocksWithSignals(c.Context(), search, limit, offset)
	} else {
		stocks, err = h.stockService.GetStocks(c.Context(), search, limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get stocks",
			"details": err.Error(),
		})
	}

	// Convert to response format
	responses := make([]*models.StockResponse, len(stocks))
	for i, stock := range stocks {
		responses[i] = stock.ToResponse()
	}

	return c.JSON(fiber.Map{
		"stocks": responses,
		"meta": fiber.Map{
			"limit":  limit,
			"offset": offset,
			"count":  len(responses),
		},
	})
}

// GetStock handles GET /stocks/:id
func (h *StockHandler) GetStock(c *fiber.Ctx) error {
	stockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stock ID",
		})
	}

	stock, err := h.stockService.GetStock(c.Context(), stockID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get stock",
			"details": err.Error(),
		})
	}

	return c.JSON(stock.ToResponse())
}

// GetStockByTicker handles GET /stocks/ticker/:ticker
func (h *StockHandler) GetStockByTicker(c *fiber.Ctx) error {
	ticker := c.Params("ticker")
	if ticker == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ticker symbol is required",
		})
	}

	stock, err := h.stockService.GetStockByTicker(c.Context(), ticker)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get stock",
			"details": err.Error(),
		})
	}

	return c.JSON(stock.ToResponse())
}

// UpdateStock handles PUT /stocks/:id
func (h *StockHandler) UpdateStock(c *fiber.Ctx) error {
	stockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stock ID",
		})
	}

	var req models.UpdateStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate request
	if err := models.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	stock, err := h.stockService.UpdateStock(c.Context(), stockID, &req)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update stock",
			"details": err.Error(),
		})
	}

	return c.JSON(stock.ToResponse())
}

// DeleteStock handles DELETE /stocks/:id
func (h *StockHandler) DeleteStock(c *fiber.Ctx) error {
	stockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stock ID",
		})
	}

	err = h.stockService.DeleteStock(c.Context(), stockID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to delete stock",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// UpdateStockSignal handles PUT /stocks/:id/signal
func (h *StockHandler) UpdateStockSignal(c *fiber.Ctx) error {
	stockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stock ID",
		})
	}

	var req models.UpdateSignalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate request
	if err := models.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	signal, err := h.stockService.UpdateStockSignal(c.Context(), stockID, req.Signal)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update stock signal",
			"details": err.Error(),
		})
	}

	return c.JSON(signal.ToResponse())
}

// GetStockSignalHistory handles GET /stocks/:id/signals
func (h *StockHandler) GetStockSignalHistory(c *fiber.Ctx) error {
	stockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stock ID",
		})
	}

	// Parse date range parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time
	if fromStr != "" {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'from' date format. Use YYYY-MM-DD",
			})
		}
	} else {
		// Default to 30 days ago
		from = time.Now().AddDate(0, 0, -30)
	}

	if toStr != "" {
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid 'to' date format. Use YYYY-MM-DD",
			})
		}
	} else {
		// Default to today
		to = time.Now()
	}

	signals, err := h.stockService.GetStockSignalHistory(c.Context(), stockID, from, to)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get signal history",
			"details": err.Error(),
		})
	}

	// Convert to response format
	responses := make([]*models.SignalResponse, len(signals))
	for i, signal := range signals {
		responses[i] = signal.ToResponse()
	}

	return c.JSON(fiber.Map{
		"signals": responses,
		"meta": fiber.Map{
			"from":  from.Format("2006-01-02"),
			"to":    to.Format("2006-01-02"),
			"count": len(responses),
		},
	})
}

// AddStockToStrategy handles POST /stocks/:id/strategies/:strategyId
func (h *StockHandler) AddStockToStrategy(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	userID := uuid.New()

	stockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stock ID",
		})
	}

	strategyID, err := uuid.Parse(c.Params("strategyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	err = h.stockService.AddStockToStrategy(c.Context(), strategyID, stockID, userID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock or strategy not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to add stock to strategy",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stock added to strategy successfully",
	})
}

// RemoveStockFromStrategy handles DELETE /stocks/:id/strategies/:strategyId
func (h *StockHandler) RemoveStockFromStrategy(c *fiber.Ctx) error {
	// TODO: Extract user ID from JWT token when authentication is implemented
	userID := uuid.New()

	stockID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stock ID",
		})
	}

	strategyID, err := uuid.Parse(c.Params("strategyId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid strategy ID",
		})
	}

	err = h.stockService.RemoveStockFromStrategy(c.Context(), strategyID, stockID, userID)
	if err != nil {
		if _, ok := err.(*models.NotFoundError); ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stock or strategy not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to remove stock from strategy",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}