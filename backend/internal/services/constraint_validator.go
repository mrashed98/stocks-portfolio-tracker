package services

import (
	"fmt"

	"github.com/shopspring/decimal"
	"portfolio-app/internal/models"
)

// ConstraintValidator handles validation of allocation constraints
type ConstraintValidator struct{}

// NewConstraintValidator creates a new constraint validator
func NewConstraintValidator() *ConstraintValidator {
	return &ConstraintValidator{}
}

// ConstraintViolation represents a constraint violation with suggestions
type ConstraintViolation struct {
	Type        string          `json:"type"`
	Message     string          `json:"message"`
	StockTicker string          `json:"stock_ticker,omitempty"`
	CurrentValue decimal.Decimal `json:"current_value,omitempty"`
	LimitValue   decimal.Decimal `json:"limit_value,omitempty"`
	Suggestions  []string        `json:"suggestions"`
}

// ValidationResult contains the result of constraint validation
type ValidationResult struct {
	IsValid    bool                  `json:"is_valid"`
	Violations []ConstraintViolation `json:"violations"`
}

// ValidateAllocations validates allocations against constraints and provides detailed feedback
func (cv *ConstraintValidator) ValidateAllocations(
	allocations []models.StockAllocation,
	constraints models.AllocationConstraints,
	totalInvestment decimal.Decimal,
) *ValidationResult {
	result := &ValidationResult{
		IsValid:    true,
		Violations: make([]ConstraintViolation, 0),
	}

	// Validate each allocation
	for _, allocation := range allocations {
		violations := cv.validateSingleAllocation(allocation, constraints, totalInvestment)
		result.Violations = append(result.Violations, violations...)
	}

	// Validate total constraints
	totalViolations := cv.validateTotalConstraints(allocations, constraints, totalInvestment)
	result.Violations = append(result.Violations, totalViolations...)

	result.IsValid = len(result.Violations) == 0
	return result
}

// validateSingleAllocation validates constraints for a single stock allocation
func (cv *ConstraintValidator) validateSingleAllocation(
	allocation models.StockAllocation,
	constraints models.AllocationConstraints,
	totalInvestment decimal.Decimal,
) []ConstraintViolation {
	violations := make([]ConstraintViolation, 0)

	// Check minimum allocation amount
	if allocation.AllocationValue.LessThan(constraints.MinAllocationAmount) {
		suggestions := []string{
			fmt.Sprintf("Increase allocation to at least %s", constraints.MinAllocationAmount.String()),
			"Consider removing this stock if minimum allocation cannot be met",
			"Reduce the number of stocks in your strategies to increase individual allocations",
		}

		violations = append(violations, ConstraintViolation{
			Type:         "MIN_ALLOCATION_VIOLATION",
			Message:      fmt.Sprintf("Stock %s allocation (%s) is below minimum required (%s)", allocation.Ticker, allocation.AllocationValue.String(), constraints.MinAllocationAmount.String()),
			StockTicker:  allocation.Ticker,
			CurrentValue: allocation.AllocationValue,
			LimitValue:   constraints.MinAllocationAmount,
			Suggestions:  suggestions,
		})
	}

	// Check maximum allocation percentage
	maxAllocationAmount := totalInvestment.Mul(constraints.MaxAllocationPerStock).Div(decimal.NewFromInt(100))
	if allocation.AllocationValue.GreaterThan(maxAllocationAmount) {
		suggestions := []string{
			fmt.Sprintf("Reduce allocation to maximum %s%% (%s)", constraints.MaxAllocationPerStock.String(), maxAllocationAmount.String()),
			"Add more stocks to your strategies to distribute the allocation",
			"Consider increasing your total investment amount",
			"Adjust strategy weights to reduce concentration in this stock",
		}

		violations = append(violations, ConstraintViolation{
			Type:         "MAX_ALLOCATION_VIOLATION",
			Message:      fmt.Sprintf("Stock %s allocation (%s%%) exceeds maximum allowed (%s%%)", allocation.Ticker, allocation.Weight.String(), constraints.MaxAllocationPerStock.String()),
			StockTicker:  allocation.Ticker,
			CurrentValue: allocation.Weight,
			LimitValue:   constraints.MaxAllocationPerStock,
			Suggestions:  suggestions,
		})
	}

	return violations
}

// validateTotalConstraints validates constraints that apply to the entire portfolio
func (cv *ConstraintValidator) validateTotalConstraints(
	allocations []models.StockAllocation,
	constraints models.AllocationConstraints,
	totalInvestment decimal.Decimal,
) []ConstraintViolation {
	violations := make([]ConstraintViolation, 0)

	// Calculate total allocated amount
	totalAllocated := decimal.Zero
	for _, allocation := range allocations {
		totalAllocated = totalAllocated.Add(allocation.AllocationValue)
	}

	// Check if total allocation is reasonable (not too far from total investment)
	allocationRatio := totalAllocated.Div(totalInvestment)
	minAllocationRatio := decimal.NewFromFloat(0.5) // At least 50% should be allocated
	
	if allocationRatio.LessThan(minAllocationRatio) {
		unallocatedAmount := totalInvestment.Sub(totalAllocated)
		unallocatedPercentage := unallocatedAmount.Div(totalInvestment).Mul(decimal.NewFromInt(100))
		
		suggestions := []string{
			"Consider lowering the minimum allocation amount constraint",
			"Add more stocks with 'Buy' signals to your strategies",
			"Review your strategy stock eligibility settings",
			"Consider adjusting your maximum allocation percentage to allow larger positions",
		}

		violations = append(violations, ConstraintViolation{
			Type:         "LOW_ALLOCATION_RATIO",
			Message:      fmt.Sprintf("Only %s%% of total investment is allocated, leaving %s (%s%%) unallocated", allocationRatio.Mul(decimal.NewFromInt(100)).StringFixed(1), unallocatedAmount.String(), unallocatedPercentage.StringFixed(1)),
			CurrentValue: unallocatedPercentage,
			LimitValue:   decimal.NewFromInt(50), // 50% max unallocated
			Suggestions:  suggestions,
		})
	}

	// Check for concentration risk (too few stocks)
	if len(allocations) > 0 && len(allocations) < 3 {
		suggestions := []string{
			"Consider adding more stocks to your strategies for better diversification",
			"Review your stock signals - ensure more stocks have 'Buy' signals",
			"Check strategy stock eligibility settings",
		}

		violations = append(violations, ConstraintViolation{
			Type:        "CONCENTRATION_RISK",
			Message:     fmt.Sprintf("Portfolio has only %d stocks, which may increase concentration risk", len(allocations)),
			Suggestions: suggestions,
		})
	}

	return violations
}

// ValidateConstraintsConfig validates the constraint configuration itself
func (cv *ConstraintValidator) ValidateConstraintsConfig(constraints models.AllocationConstraints, totalInvestment decimal.Decimal) *ValidationResult {
	result := &ValidationResult{
		IsValid:    true,
		Violations: make([]ConstraintViolation, 0),
	}

	// Check if max allocation percentage is reasonable
	if constraints.MaxAllocationPerStock.LessThan(decimal.NewFromInt(1)) {
		result.Violations = append(result.Violations, ConstraintViolation{
			Type:        "INVALID_MAX_ALLOCATION",
			Message:     "Maximum allocation per stock must be at least 1%",
			Suggestions: []string{"Set maximum allocation per stock to at least 1%"},
		})
	}

	if constraints.MaxAllocationPerStock.GreaterThan(decimal.NewFromInt(100)) {
		result.Violations = append(result.Violations, ConstraintViolation{
			Type:        "INVALID_MAX_ALLOCATION",
			Message:     "Maximum allocation per stock cannot exceed 100%",
			Suggestions: []string{"Set maximum allocation per stock to 100% or less"},
		})
	}

	// Check if minimum allocation amount is reasonable relative to total investment
	if totalInvestment.IsZero() {
		result.Violations = append(result.Violations, ConstraintViolation{
			Type:        "ZERO_INVESTMENT",
			Message:     "Total investment cannot be zero",
			Suggestions: []string{"Set a positive total investment amount"},
		})
		result.IsValid = false
		return result
	}
	
	minAllocationPercentage := constraints.MinAllocationAmount.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	if minAllocationPercentage.GreaterThan(constraints.MaxAllocationPerStock) {
		result.Violations = append(result.Violations, ConstraintViolation{
			Type:    "CONFLICTING_CONSTRAINTS",
			Message: fmt.Sprintf("Minimum allocation amount (%s, %s%% of total) exceeds maximum allocation percentage (%s%%)", constraints.MinAllocationAmount.String(), minAllocationPercentage.StringFixed(2), constraints.MaxAllocationPerStock.String()),
			Suggestions: []string{
				"Reduce minimum allocation amount",
				"Increase maximum allocation percentage",
				"Increase total investment amount",
			},
		})
	}

	// Check if minimum allocation is too high (would allow too few stocks)
	if !constraints.MinAllocationAmount.IsZero() {
		maxPossibleStocks := totalInvestment.Div(constraints.MinAllocationAmount)
		if maxPossibleStocks.LessThan(decimal.NewFromInt(2)) {
			result.Violations = append(result.Violations, ConstraintViolation{
				Type:    "HIGH_MIN_ALLOCATION",
				Message: fmt.Sprintf("Minimum allocation amount (%s) is too high - would allow fewer than 2 stocks in portfolio", constraints.MinAllocationAmount.String()),
				Suggestions: []string{
					"Reduce minimum allocation amount to allow more diversification",
					"Increase total investment amount",
				},
			})
		}
	}

	result.IsValid = len(result.Violations) == 0
	return result
}

// SuggestConstraintAdjustments provides suggestions for constraint adjustments based on strategy analysis
func (cv *ConstraintValidator) SuggestConstraintAdjustments(
	strategies []*models.Strategy,
	totalInvestment decimal.Decimal,
	currentConstraints models.AllocationConstraints,
) []string {
	suggestions := make([]string, 0)

	// Count total eligible stocks across all strategies (approximate)
	// This would need actual strategy stock data in a real implementation
	estimatedStockCount := len(strategies) * 5 // Rough estimate

	if estimatedStockCount > 0 {
		// Suggest reasonable minimum allocation
		suggestedMinAllocation := totalInvestment.Div(decimal.NewFromInt(int64(estimatedStockCount * 2)))
		if suggestedMinAllocation.LessThan(currentConstraints.MinAllocationAmount) {
			suggestions = append(suggestions, fmt.Sprintf("Consider reducing minimum allocation to %s to allow more diversification", suggestedMinAllocation.StringFixed(0)))
		}

		// Suggest reasonable maximum allocation
		suggestedMaxAllocation := decimal.NewFromInt(100).Div(decimal.NewFromInt(int64(estimatedStockCount / 2)))
		if suggestedMaxAllocation.GreaterThan(currentConstraints.MaxAllocationPerStock) {
			suggestions = append(suggestions, fmt.Sprintf("Consider increasing maximum allocation to %s%% to allow proper distribution", suggestedMaxAllocation.StringFixed(0)))
		}
	}

	return suggestions
}