package services

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// AllocationError represents an error during allocation calculation
type AllocationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *AllocationError) Error() string {
	return e.Message
}

// NewAllocationError creates a new allocation error
func NewAllocationError(errorType, message string, details map[string]interface{}) *AllocationError {
	return &AllocationError{
		Type:    errorType,
		Message: message,
		Details: details,
	}
}

// Specific error types
var (
	ErrInvalidStrategyWeights = func(totalPercentage decimal.Decimal) *AllocationError {
		return NewAllocationError(
			"INVALID_STRATEGY_WEIGHTS",
			fmt.Sprintf("Total strategy weights (%s%%) exceed 100%%", totalPercentage.String()),
			map[string]interface{}{
				"total_percentage": totalPercentage,
				"max_allowed":      100,
			},
		)
	}

	ErrBudgetExceedsInvestment = func(totalBudget, totalInvestment decimal.Decimal) *AllocationError {
		return NewAllocationError(
			"BUDGET_EXCEEDS_INVESTMENT",
			fmt.Sprintf("Total budget strategies (%s) exceed total investment (%s)", totalBudget.String(), totalInvestment.String()),
			map[string]interface{}{
				"total_budget":     totalBudget,
				"total_investment": totalInvestment,
				"excess_amount":    totalBudget.Sub(totalInvestment),
			},
		)
	}

	ErrNoEligibleStocks = func(strategyName string) *AllocationError {
		return NewAllocationError(
			"NO_ELIGIBLE_STOCKS",
			fmt.Sprintf("Strategy '%s' has no eligible stocks with 'Buy' signals", strategyName),
			map[string]interface{}{
				"strategy_name": strategyName,
				"suggestion":    "Ensure stocks in this strategy have 'Buy' signals and are marked as eligible",
			},
		)
	}

	ErrConstraintViolation = func(violations []ConstraintViolation) *AllocationError {
		return NewAllocationError(
			"CONSTRAINT_VIOLATION",
			fmt.Sprintf("Allocation constraints violated: %d violations found", len(violations)),
			map[string]interface{}{
				"violations": violations,
			},
		)
	}

	ErrInsufficientAllocation = func(totalAllocated, totalInvestment decimal.Decimal) *AllocationError {
		unallocated := totalInvestment.Sub(totalAllocated)
		percentage := unallocated.Div(totalInvestment).Mul(decimal.NewFromInt(100))
		
		return NewAllocationError(
			"INSUFFICIENT_ALLOCATION",
			fmt.Sprintf("Only %s of %s (%s%%) could be allocated due to constraints", 
				totalAllocated.String(), totalInvestment.String(), percentage.StringFixed(1)),
			map[string]interface{}{
				"total_allocated":   totalAllocated,
				"total_investment":  totalInvestment,
				"unallocated_cash":  unallocated,
				"unallocated_percentage": percentage,
			},
		)
	}
)

// IsAllocationError checks if an error is an AllocationError
func IsAllocationError(err error) bool {
	_, ok := err.(*AllocationError)
	return ok
}

// GetAllocationError extracts AllocationError from error
func GetAllocationError(err error) *AllocationError {
	if allocErr, ok := err.(*AllocationError); ok {
		return allocErr
	}
	return nil
}