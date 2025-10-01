package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Portfolio represents a portfolio in the system
type Portfolio struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	Name            string          `json:"name" db:"name" validate:"required,min=1,max=255"`
	TotalInvestment decimal.Decimal `json:"total_investment" db:"total_investment" validate:"required,gt=0"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
	
	// Related data (not stored in database)
	Positions  []Position   `json:"positions,omitempty"`
	NAVHistory []NAVHistory `json:"nav_history,omitempty"`
}

// CreatePortfolioRequest represents the request to create a new portfolio
type CreatePortfolioRequest struct {
	Name            string                   `json:"name" validate:"required,min=1,max=255"`
	TotalInvestment decimal.Decimal          `json:"total_investment" validate:"required,gt=0"`
	Positions       []CreatePositionRequest  `json:"positions" validate:"required,min=1,dive"`
}

// UpdatePortfolioRequest represents the request to update a portfolio
type UpdatePortfolioRequest struct {
	Name            *string          `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	TotalInvestment *decimal.Decimal `json:"total_investment,omitempty" validate:"omitempty,gt=0"`
}

// PortfolioResponse represents the portfolio data returned in API responses
type PortfolioResponse struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"user_id"`
	Name            string          `json:"name"`
	TotalInvestment decimal.Decimal `json:"total_investment"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Positions       []Position      `json:"positions,omitempty"`
	NAVHistory      []NAVHistory    `json:"nav_history,omitempty"`
	CurrentNAV      *decimal.Decimal `json:"current_nav,omitempty"`
	TotalPnL        *decimal.Decimal `json:"total_pnl,omitempty"`
	MaxDrawdown     *decimal.Decimal `json:"max_drawdown,omitempty"`
}

// AllocationPreview represents a preview of portfolio allocation
type AllocationPreview struct {
	TotalInvestment decimal.Decimal      `json:"total_investment"`
	Allocations     []StockAllocation    `json:"allocations"`
	UnallocatedCash decimal.Decimal      `json:"unallocated_cash"`
	TotalAllocated  decimal.Decimal      `json:"total_allocated"`
	Constraints     AllocationConstraints `json:"constraints"`
}

// StockAllocation represents the allocation for a single stock
type StockAllocation struct {
	StockID         uuid.UUID       `json:"stock_id"`
	Ticker          string          `json:"ticker"`
	Name            string          `json:"name"`
	Weight          decimal.Decimal `json:"weight"`
	AllocationValue decimal.Decimal `json:"allocation_value"`
	Price           decimal.Decimal `json:"price"`
	Quantity        int             `json:"quantity"`
	ActualValue     decimal.Decimal `json:"actual_value"`
	StrategyContrib map[string]decimal.Decimal `json:"strategy_contrib"`
}

// AllocationConstraints represents constraints for portfolio allocation
type AllocationConstraints struct {
	MaxAllocationPerStock decimal.Decimal `json:"max_allocation_per_stock" validate:"gt=0,lte=100"`
	MinAllocationAmount   decimal.Decimal `json:"min_allocation_amount" validate:"gte=0"`
}

// AllocationRequest represents a request to generate allocation preview
type AllocationRequest struct {
	StrategyIDs     []uuid.UUID           `json:"strategy_ids" validate:"required,min=1"`
	TotalInvestment decimal.Decimal       `json:"total_investment" validate:"required,gt=0"`
	Constraints     AllocationConstraints `json:"constraints"`
	ExcludedStocks  []uuid.UUID          `json:"excluded_stocks,omitempty"`
}

// ToResponse converts a Portfolio to PortfolioResponse
func (p *Portfolio) ToResponse() *PortfolioResponse {
	response := &PortfolioResponse{
		ID:              p.ID,
		UserID:          p.UserID,
		Name:            p.Name,
		TotalInvestment: p.TotalInvestment,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Positions:       p.Positions,
		NAVHistory:      p.NAVHistory,
	}
	
	// Calculate current metrics if NAV history exists
	if len(p.NAVHistory) > 0 {
		latest := p.NAVHistory[len(p.NAVHistory)-1]
		response.CurrentNAV = &latest.NAV
		response.TotalPnL = &latest.PnL
		response.MaxDrawdown = latest.Drawdown
	}
	
	return response
}

// FromCreateRequest creates a Portfolio from CreatePortfolioRequest
func (p *Portfolio) FromCreateRequest(req *CreatePortfolioRequest, userID uuid.UUID) {
	p.ID = uuid.New()
	p.UserID = userID
	p.Name = req.Name
	p.TotalInvestment = req.TotalInvestment
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
}

// ApplyUpdate applies an UpdatePortfolioRequest to the portfolio
func (p *Portfolio) ApplyUpdate(req *UpdatePortfolioRequest) {
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.TotalInvestment != nil {
		p.TotalInvestment = *req.TotalInvestment
	}
	p.UpdatedAt = time.Now()
}