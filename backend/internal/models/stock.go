package models

import (
	"time"

	"github.com/google/uuid"
)

// Stock represents a stock in the system
type Stock struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Ticker    string    `json:"ticker" db:"ticker" validate:"required,min=1,max=20,uppercase"`
	Name      string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	Sector    *string   `json:"sector" db:"sector" validate:"omitempty,max=100"`
	Exchange  *string   `json:"exchange" db:"exchange" validate:"omitempty,max=50"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	
	// Related data (not stored in database)
	CurrentSignal *Signal `json:"current_signal,omitempty"`
}

// CreateStockRequest represents the request to create a new stock
type CreateStockRequest struct {
	Ticker   string  `json:"ticker" validate:"required,min=1,max=20,uppercase"`
	Name     string  `json:"name" validate:"required,min=1,max=255"`
	Sector   *string `json:"sector,omitempty" validate:"omitempty,max=100"`
	Exchange *string `json:"exchange,omitempty" validate:"omitempty,max=50"`
}

// UpdateStockRequest represents the request to update a stock
type UpdateStockRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Sector   *string `json:"sector,omitempty" validate:"omitempty,max=100"`
	Exchange *string `json:"exchange,omitempty" validate:"omitempty,max=50"`
}

// StockResponse represents the stock data returned in API responses
type StockResponse struct {
	ID            uuid.UUID `json:"id"`
	Ticker        string    `json:"ticker"`
	Name          string    `json:"name"`
	Sector        *string   `json:"sector"`
	Exchange      *string   `json:"exchange"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CurrentSignal *Signal   `json:"current_signal,omitempty"`
}

// StrategyStock represents the relationship between a strategy and stock
type StrategyStock struct {
	StrategyID uuid.UUID `json:"strategy_id" db:"strategy_id"`
	StockID    uuid.UUID `json:"stock_id" db:"stock_id"`
	Eligible   bool      `json:"eligible" db:"eligible"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	
	// Related data (not stored in database)
	Stock *Stock `json:"stock,omitempty"`
}

// UpdateStockEligibilityRequest represents the request to update stock eligibility
type UpdateStockEligibilityRequest struct {
	Eligible bool `json:"eligible" validate:"required"`
}

// ToResponse converts a Stock to StockResponse
func (s *Stock) ToResponse() *StockResponse {
	return &StockResponse{
		ID:            s.ID,
		Ticker:        s.Ticker,
		Name:          s.Name,
		Sector:        s.Sector,
		Exchange:      s.Exchange,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
		CurrentSignal: s.CurrentSignal,
	}
}

// FromCreateRequest creates a Stock from CreateStockRequest
func (s *Stock) FromCreateRequest(req *CreateStockRequest) {
	s.ID = uuid.New()
	s.Ticker = req.Ticker
	s.Name = req.Name
	s.Sector = req.Sector
	s.Exchange = req.Exchange
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
}

// ApplyUpdate applies an UpdateStockRequest to the stock
func (s *Stock) ApplyUpdate(req *UpdateStockRequest) {
	if req.Name != nil {
		s.Name = *req.Name
	}
	if req.Sector != nil {
		s.Sector = req.Sector
	}
	if req.Exchange != nil {
		s.Exchange = req.Exchange
	}
	s.UpdatedAt = time.Now()
}