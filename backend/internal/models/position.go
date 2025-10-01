package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Position represents a position in a portfolio
type Position struct {
	PortfolioID     uuid.UUID       `json:"portfolio_id" db:"portfolio_id"`
	StockID         uuid.UUID       `json:"stock_id" db:"stock_id"`
	Quantity        int             `json:"quantity" db:"quantity" validate:"required,gt=0"`
	EntryPrice      decimal.Decimal `json:"entry_price" db:"entry_price" validate:"required,gt=0"`
	AllocationValue decimal.Decimal `json:"allocation_value" db:"allocation_value" validate:"required,gt=0"`
	StrategyContrib json.RawMessage `json:"strategy_contrib" db:"strategy_contrib"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	
	// Related data (not stored in database)
	Stock               *Stock           `json:"stock,omitempty"`
	CurrentPrice        *decimal.Decimal `json:"current_price,omitempty"`
	CurrentValue        *decimal.Decimal `json:"current_value,omitempty"`
	PnL                 *decimal.Decimal `json:"pnl,omitempty"`
	PnLPercentage       *decimal.Decimal `json:"pnl_percentage,omitempty"`
	StrategyContribMap  map[string]decimal.Decimal `json:"strategy_contrib_map,omitempty"`
}

// CreatePositionRequest represents the request to create a new position
type CreatePositionRequest struct {
	StockID         uuid.UUID                   `json:"stock_id" validate:"required"`
	Quantity        int                         `json:"quantity" validate:"required,gt=0"`
	EntryPrice      decimal.Decimal             `json:"entry_price" validate:"required,gt=0"`
	AllocationValue decimal.Decimal             `json:"allocation_value" validate:"required,gt=0"`
	StrategyContrib map[string]decimal.Decimal  `json:"strategy_contrib" validate:"required"`
}

// UpdatePositionRequest represents the request to update a position
type UpdatePositionRequest struct {
	Quantity        *int             `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	EntryPrice      *decimal.Decimal `json:"entry_price,omitempty" validate:"omitempty,gt=0"`
	AllocationValue *decimal.Decimal `json:"allocation_value,omitempty" validate:"omitempty,gt=0"`
}

// PositionResponse represents the position data returned in API responses
type PositionResponse struct {
	PortfolioID         uuid.UUID                   `json:"portfolio_id"`
	StockID             uuid.UUID                   `json:"stock_id"`
	Quantity            int                         `json:"quantity"`
	EntryPrice          decimal.Decimal             `json:"entry_price"`
	AllocationValue     decimal.Decimal             `json:"allocation_value"`
	StrategyContrib     map[string]decimal.Decimal  `json:"strategy_contrib"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	Stock               *Stock                      `json:"stock,omitempty"`
	CurrentPrice        *decimal.Decimal            `json:"current_price,omitempty"`
	CurrentValue        *decimal.Decimal            `json:"current_value,omitempty"`
	PnL                 *decimal.Decimal            `json:"pnl,omitempty"`
	PnLPercentage       *decimal.Decimal            `json:"pnl_percentage,omitempty"`
}

// ToResponse converts a Position to PositionResponse
func (p *Position) ToResponse() *PositionResponse {
	response := &PositionResponse{
		PortfolioID:     p.PortfolioID,
		StockID:         p.StockID,
		Quantity:        p.Quantity,
		EntryPrice:      p.EntryPrice,
		AllocationValue: p.AllocationValue,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Stock:           p.Stock,
		CurrentPrice:    p.CurrentPrice,
		CurrentValue:    p.CurrentValue,
		PnL:             p.PnL,
		PnLPercentage:   p.PnLPercentage,
	}
	
	// Convert strategy contribution from JSON to map
	if p.StrategyContribMap != nil {
		response.StrategyContrib = p.StrategyContribMap
	} else if len(p.StrategyContrib) > 0 {
		var contribMap map[string]decimal.Decimal
		if err := json.Unmarshal(p.StrategyContrib, &contribMap); err == nil {
			response.StrategyContrib = contribMap
		}
	}
	
	return response
}

// FromCreateRequest creates a Position from CreatePositionRequest
func (p *Position) FromCreateRequest(req *CreatePositionRequest, portfolioID uuid.UUID) error {
	p.PortfolioID = portfolioID
	p.StockID = req.StockID
	p.Quantity = req.Quantity
	p.EntryPrice = req.EntryPrice
	p.AllocationValue = req.AllocationValue
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	
	// Convert strategy contribution map to JSON
	contribJSON, err := json.Marshal(req.StrategyContrib)
	if err != nil {
		return err
	}
	p.StrategyContrib = contribJSON
	p.StrategyContribMap = req.StrategyContrib
	
	return nil
}

// ApplyUpdate applies an UpdatePositionRequest to the position
func (p *Position) ApplyUpdate(req *UpdatePositionRequest) {
	if req.Quantity != nil {
		p.Quantity = *req.Quantity
	}
	if req.EntryPrice != nil {
		p.EntryPrice = *req.EntryPrice
	}
	if req.AllocationValue != nil {
		p.AllocationValue = *req.AllocationValue
	}
	p.UpdatedAt = time.Now()
}

// CalculateMetrics calculates current value, P&L, and P&L percentage
func (p *Position) CalculateMetrics(currentPrice decimal.Decimal) {
	p.CurrentPrice = &currentPrice
	
	// Calculate current value
	currentValue := currentPrice.Mul(decimal.NewFromInt(int64(p.Quantity)))
	p.CurrentValue = &currentValue
	
	// Calculate P&L
	pnl := currentValue.Sub(p.AllocationValue)
	p.PnL = &pnl
	
	// Calculate P&L percentage
	if p.AllocationValue.GreaterThan(decimal.Zero) {
		pnlPercentage := pnl.Div(p.AllocationValue).Mul(decimal.NewFromInt(100))
		p.PnLPercentage = &pnlPercentage
	}
}