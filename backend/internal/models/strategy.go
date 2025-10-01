package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WeightMode represents the weight mode for a strategy
type WeightMode string

const (
	WeightModePercent WeightMode = "percent"
	WeightModeBudget  WeightMode = "budget"
)

// Strategy represents an investment strategy
type Strategy struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	Name        string          `json:"name" db:"name" validate:"required,min=1,max=255"`
	WeightMode  WeightMode      `json:"weight_mode" db:"weight_mode" validate:"required,oneof=percent budget"`
	WeightValue decimal.Decimal `json:"weight_value" db:"weight_value" validate:"required,gt=0"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	
	// Related data (not stored in database)
	Stocks []StrategyStock `json:"stocks,omitempty"`
}

// CreateStrategyRequest represents the request to create a new strategy
type CreateStrategyRequest struct {
	Name        string          `json:"name" validate:"required,min=1,max=255"`
	WeightMode  WeightMode      `json:"weight_mode" validate:"required,oneof=percent budget"`
	WeightValue decimal.Decimal `json:"weight_value" validate:"required,gt=0"`
}

// UpdateStrategyRequest represents the request to update a strategy
type UpdateStrategyRequest struct {
	Name        *string          `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	WeightMode  *WeightMode      `json:"weight_mode,omitempty" validate:"omitempty,oneof=percent budget"`
	WeightValue *decimal.Decimal `json:"weight_value,omitempty" validate:"omitempty,gt=0"`
}

// StrategyResponse represents the strategy data returned in API responses
type StrategyResponse struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	Name        string          `json:"name"`
	WeightMode  WeightMode      `json:"weight_mode"`
	WeightValue decimal.Decimal `json:"weight_value"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Stocks      []StrategyStock `json:"stocks,omitempty"`
}

// ToResponse converts a Strategy to StrategyResponse
func (s *Strategy) ToResponse() *StrategyResponse {
	return &StrategyResponse{
		ID:          s.ID,
		UserID:      s.UserID,
		Name:        s.Name,
		WeightMode:  s.WeightMode,
		WeightValue: s.WeightValue,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Stocks:      s.Stocks,
	}
}

// FromCreateRequest creates a Strategy from CreateStrategyRequest
func (s *Strategy) FromCreateRequest(req *CreateStrategyRequest, userID uuid.UUID) {
	s.ID = uuid.New()
	s.UserID = userID
	s.Name = req.Name
	s.WeightMode = req.WeightMode
	s.WeightValue = req.WeightValue
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
}

// ApplyUpdate applies an UpdateStrategyRequest to the strategy
func (s *Strategy) ApplyUpdate(req *UpdateStrategyRequest) {
	if req.Name != nil {
		s.Name = *req.Name
	}
	if req.WeightMode != nil {
		s.WeightMode = *req.WeightMode
	}
	if req.WeightValue != nil {
		s.WeightValue = *req.WeightValue
	}
	s.UpdatedAt = time.Now()
}