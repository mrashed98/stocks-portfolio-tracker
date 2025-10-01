package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// NAVHistory represents the net asset value history for a portfolio
type NAVHistory struct {
	PortfolioID uuid.UUID        `json:"portfolio_id" db:"portfolio_id"`
	Timestamp   time.Time        `json:"timestamp" db:"timestamp"`
	NAV         decimal.Decimal  `json:"nav" db:"nav" validate:"required,gte=0"`
	PnL         decimal.Decimal  `json:"pnl" db:"pnl"`
	Drawdown    *decimal.Decimal `json:"drawdown" db:"drawdown"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	
	// Related data (not stored in database)
	Portfolio *Portfolio `json:"portfolio,omitempty"`
}

// CreateNAVHistoryRequest represents the request to create a new NAV history entry
type CreateNAVHistoryRequest struct {
	PortfolioID uuid.UUID        `json:"portfolio_id" validate:"required"`
	Timestamp   time.Time        `json:"timestamp" validate:"required"`
	NAV         decimal.Decimal  `json:"nav" validate:"required,gte=0"`
	PnL         decimal.Decimal  `json:"pnl"`
	Drawdown    *decimal.Decimal `json:"drawdown,omitempty"`
}

// NAVHistoryResponse represents the NAV history data returned in API responses
type NAVHistoryResponse struct {
	PortfolioID uuid.UUID        `json:"portfolio_id"`
	Timestamp   time.Time        `json:"timestamp"`
	NAV         decimal.Decimal  `json:"nav"`
	PnL         decimal.Decimal  `json:"pnl"`
	Drawdown    *decimal.Decimal `json:"drawdown"`
	CreatedAt   time.Time        `json:"created_at"`
	Portfolio   *Portfolio       `json:"portfolio,omitempty"`
}

// PerformanceMetrics represents calculated performance metrics for a portfolio
type PerformanceMetrics struct {
	TotalReturn       decimal.Decimal  `json:"total_return"`
	TotalReturnPct    decimal.Decimal  `json:"total_return_pct"`
	AnnualizedReturn  *decimal.Decimal `json:"annualized_return,omitempty"`
	MaxDrawdown       *decimal.Decimal `json:"max_drawdown"`
	CurrentDrawdown   *decimal.Decimal `json:"current_drawdown"`
	VolatilityPct     *decimal.Decimal `json:"volatility_pct,omitempty"`
	SharpeRatio       *decimal.Decimal `json:"sharpe_ratio,omitempty"`
	DaysActive        int              `json:"days_active"`
	HighWaterMark     decimal.Decimal  `json:"high_water_mark"`
}

// ToResponse converts a NAVHistory to NAVHistoryResponse
func (n *NAVHistory) ToResponse() *NAVHistoryResponse {
	return &NAVHistoryResponse{
		PortfolioID: n.PortfolioID,
		Timestamp:   n.Timestamp,
		NAV:         n.NAV,
		PnL:         n.PnL,
		Drawdown:    n.Drawdown,
		CreatedAt:   n.CreatedAt,
		Portfolio:   n.Portfolio,
	}
}

// FromCreateRequest creates a NAVHistory from CreateNAVHistoryRequest
func (n *NAVHistory) FromCreateRequest(req *CreateNAVHistoryRequest) {
	n.PortfolioID = req.PortfolioID
	n.Timestamp = req.Timestamp
	n.NAV = req.NAV
	n.PnL = req.PnL
	n.Drawdown = req.Drawdown
	n.CreatedAt = time.Now()
}

// CalculateDrawdown calculates the drawdown from a high water mark
func (n *NAVHistory) CalculateDrawdown(highWaterMark decimal.Decimal) {
	if highWaterMark.GreaterThan(decimal.Zero) && n.NAV.LessThan(highWaterMark) {
		drawdown := n.NAV.Sub(highWaterMark).Div(highWaterMark).Mul(decimal.NewFromInt(100))
		n.Drawdown = &drawdown
	} else {
		// No drawdown if at or above high water mark
		zero := decimal.Zero
		n.Drawdown = &zero
	}
}

// CalculatePerformanceMetrics calculates performance metrics from NAV history
func CalculatePerformanceMetrics(history []NAVHistory, initialInvestment decimal.Decimal) *PerformanceMetrics {
	if len(history) == 0 {
		return &PerformanceMetrics{}
	}
	
	latest := history[len(history)-1]
	
	// Calculate total return
	totalReturn := latest.NAV.Sub(initialInvestment)
	totalReturnPct := decimal.Zero
	if initialInvestment.GreaterThan(decimal.Zero) {
		totalReturnPct = totalReturn.Div(initialInvestment).Mul(decimal.NewFromInt(100))
	}
	
	// Find high water mark and max drawdown
	highWaterMark := initialInvestment
	var maxDrawdown *decimal.Decimal
	
	for _, entry := range history {
		if entry.NAV.GreaterThan(highWaterMark) {
			highWaterMark = entry.NAV
		}
		
		if entry.Drawdown != nil {
			if maxDrawdown == nil || entry.Drawdown.LessThan(*maxDrawdown) {
				maxDrawdown = entry.Drawdown
			}
		}
	}
	
	// Calculate days active
	daysActive := 0
	if len(history) > 1 {
		firstEntry := history[0]
		lastEntry := history[len(history)-1]
		daysActive = int(lastEntry.Timestamp.Sub(firstEntry.Timestamp).Hours() / 24)
	}
	
	metrics := &PerformanceMetrics{
		TotalReturn:    totalReturn,
		TotalReturnPct: totalReturnPct,
		MaxDrawdown:    maxDrawdown,
		DaysActive:     daysActive,
		HighWaterMark:  highWaterMark,
	}
	
	// Calculate current drawdown
	if highWaterMark.GreaterThan(decimal.Zero) && latest.NAV.LessThan(highWaterMark) {
		currentDrawdown := latest.NAV.Sub(highWaterMark).Div(highWaterMark).Mul(decimal.NewFromInt(100))
		metrics.CurrentDrawdown = &currentDrawdown
	}
	
	// Calculate annualized return if we have enough data
	if daysActive > 0 {
		years := decimal.NewFromFloat(float64(daysActive) / 365.25)
		if years.GreaterThan(decimal.Zero) && initialInvestment.GreaterThan(decimal.Zero) {
			// Annualized return = (Final Value / Initial Value)^(1/years) - 1
			finalRatio := latest.NAV.Div(initialInvestment)
			if finalRatio.GreaterThan(decimal.Zero) {
				// Simplified calculation - for more accurate results, would need proper power function
				annualizedReturn := finalRatio.Sub(decimal.NewFromInt(1)).Div(years).Mul(decimal.NewFromInt(100))
				metrics.AnnualizedReturn = &annualizedReturn
			}
		}
	}
	
	return metrics
}