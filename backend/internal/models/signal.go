package models

import (
	"time"

	"github.com/google/uuid"
)

// SignalType represents the type of signal for a stock
type SignalType string

const (
	SignalBuy  SignalType = "Buy"
	SignalHold SignalType = "Hold"
)

// Signal represents a trading signal for a stock
type Signal struct {
	StockID   uuid.UUID  `json:"stock_id" db:"stock_id"`
	Signal    SignalType `json:"signal" db:"signal" validate:"required,oneof=Buy Hold"`
	Date      time.Time  `json:"date" db:"date" validate:"required"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	
	// Related data (not stored in database)
	Stock *Stock `json:"stock,omitempty"`
}

// CreateSignalRequest represents the request to create a new signal
type CreateSignalRequest struct {
	StockID uuid.UUID  `json:"stock_id" validate:"required"`
	Signal  SignalType `json:"signal" validate:"required,oneof=Buy Hold"`
	Date    time.Time  `json:"date" validate:"required"`
}

// UpdateSignalRequest represents the request to update a signal
type UpdateSignalRequest struct {
	Signal SignalType `json:"signal" validate:"required,oneof=Buy Hold"`
}

// SignalResponse represents the signal data returned in API responses
type SignalResponse struct {
	StockID   uuid.UUID  `json:"stock_id"`
	Signal    SignalType `json:"signal"`
	Date      time.Time  `json:"date"`
	CreatedAt time.Time  `json:"created_at"`
	Stock     *Stock     `json:"stock,omitempty"`
}

// ToResponse converts a Signal to SignalResponse
func (s *Signal) ToResponse() *SignalResponse {
	return &SignalResponse{
		StockID:   s.StockID,
		Signal:    s.Signal,
		Date:      s.Date,
		CreatedAt: s.CreatedAt,
		Stock:     s.Stock,
	}
}

// FromCreateRequest creates a Signal from CreateSignalRequest
func (s *Signal) FromCreateRequest(req *CreateSignalRequest) {
	s.StockID = req.StockID
	s.Signal = req.Signal
	s.Date = req.Date
	s.CreatedAt = time.Now()
}