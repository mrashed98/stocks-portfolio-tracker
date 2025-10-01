package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"portfolio-app/internal/models"
)

func TestConstraintValidator_ValidateAllocations_Success(t *testing.T) {
	validator := NewConstraintValidator()
	
	allocations := []models.StockAllocation{
		{
			StockID:         uuid.New(),
			Ticker:          "AAPL",
			AllocationValue: decimal.NewFromInt(3000),
			Weight:          decimal.NewFromInt(30),
		},
		{
			StockID:         uuid.New(),
			Ticker:          "GOOGL",
			AllocationValue: decimal.NewFromInt(2000),
			Weight:          decimal.NewFromInt(20),
		},
		{
			StockID:         uuid.New(),
			Ticker:          "MSFT",
			AllocationValue: decimal.NewFromInt(1500),
			Weight:          decimal.NewFromInt(15),
		},
	}
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(40), // 40%
		MinAllocationAmount:   decimal.NewFromInt(1000),
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateAllocations(allocations, constraints, totalInvestment)
	
	assert.True(t, result.IsValid)
	assert.Empty(t, result.Violations)
}

func TestConstraintValidator_ValidateAllocations_MinAllocationViolation(t *testing.T) {
	validator := NewConstraintValidator()
	
	allocations := []models.StockAllocation{
		{
			StockID:         uuid.New(),
			Ticker:          "AAPL",
			AllocationValue: decimal.NewFromInt(50), // Below minimum
			Weight:          decimal.NewFromFloat(0.5),
		},
	}
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(50),
		MinAllocationAmount:   decimal.NewFromInt(100),
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateAllocations(allocations, constraints, totalInvestment)
	
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 3) // Min allocation + low allocation ratio + concentration risk
	
	// Check min allocation violation
	minViolation := result.Violations[0]
	assert.Equal(t, "MIN_ALLOCATION_VIOLATION", minViolation.Type)
	assert.Equal(t, "AAPL", minViolation.StockTicker)
	assert.True(t, decimal.NewFromInt(50).Equal(minViolation.CurrentValue))
	assert.True(t, decimal.NewFromInt(100).Equal(minViolation.LimitValue))
	assert.NotEmpty(t, minViolation.Suggestions)
}

func TestConstraintValidator_ValidateAllocations_MaxAllocationViolation(t *testing.T) {
	validator := NewConstraintValidator()
	
	allocations := []models.StockAllocation{
		{
			StockID:         uuid.New(),
			Ticker:          "AAPL",
			AllocationValue: decimal.NewFromInt(6000),
			Weight:          decimal.NewFromInt(60), // Above maximum
		},
	}
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(50), // 50%
		MinAllocationAmount:   decimal.NewFromInt(100),
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateAllocations(allocations, constraints, totalInvestment)
	
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 2) // Max allocation + concentration risk
	
	violation := result.Violations[0]
	assert.Equal(t, "MAX_ALLOCATION_VIOLATION", violation.Type)
	assert.Equal(t, "AAPL", violation.StockTicker)
	assert.True(t, decimal.NewFromInt(60).Equal(violation.CurrentValue))
	assert.True(t, decimal.NewFromInt(50).Equal(violation.LimitValue))
	assert.NotEmpty(t, violation.Suggestions)
}

func TestConstraintValidator_ValidateAllocations_LowAllocationRatio(t *testing.T) {
	validator := NewConstraintValidator()
	
	// Only 30% of total investment allocated
	allocations := []models.StockAllocation{
		{
			StockID:         uuid.New(),
			Ticker:          "AAPL",
			AllocationValue: decimal.NewFromInt(1500),
			Weight:          decimal.NewFromInt(15),
		},
		{
			StockID:         uuid.New(),
			Ticker:          "GOOGL",
			AllocationValue: decimal.NewFromInt(1500),
			Weight:          decimal.NewFromInt(15),
		},
	}
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(50),
		MinAllocationAmount:   decimal.NewFromInt(100),
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateAllocations(allocations, constraints, totalInvestment)
	
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 2) // Low allocation ratio + concentration risk
	
	violation := result.Violations[0]
	assert.Equal(t, "LOW_ALLOCATION_RATIO", violation.Type)
	assert.NotEmpty(t, violation.Suggestions)
}

func TestConstraintValidator_ValidateAllocations_ConcentrationRisk(t *testing.T) {
	validator := NewConstraintValidator()
	
	// Only 1 stock - concentration risk
	allocations := []models.StockAllocation{
		{
			StockID:         uuid.New(),
			Ticker:          "AAPL",
			AllocationValue: decimal.NewFromInt(8000),
			Weight:          decimal.NewFromInt(80),
		},
	}
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(90),
		MinAllocationAmount:   decimal.NewFromInt(100),
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateAllocations(allocations, constraints, totalInvestment)
	
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 1)
	
	violation := result.Violations[0]
	assert.Equal(t, "CONCENTRATION_RISK", violation.Type)
	assert.NotEmpty(t, violation.Suggestions)
}

func TestConstraintValidator_ValidateConstraintsConfig_Success(t *testing.T) {
	validator := NewConstraintValidator()
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(25), // 25%
		MinAllocationAmount:   decimal.NewFromInt(500), // 5% of 10000
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateConstraintsConfig(constraints, totalInvestment)
	
	assert.True(t, result.IsValid)
	assert.Empty(t, result.Violations)
}

func TestConstraintValidator_ValidateConstraintsConfig_InvalidMaxAllocation(t *testing.T) {
	validator := NewConstraintValidator()
	
	// Test too low max allocation
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromFloat(0.5), // 0.5%
		MinAllocationAmount:   decimal.NewFromInt(100),
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateConstraintsConfig(constraints, totalInvestment)
	
	assert.False(t, result.IsValid)
	assert.GreaterOrEqual(t, len(result.Violations), 1) // At least invalid max allocation
	assert.Equal(t, "INVALID_MAX_ALLOCATION", result.Violations[0].Type)
	
	// Test too high max allocation
	constraints.MaxAllocationPerStock = decimal.NewFromInt(150) // 150%
	constraints.MinAllocationAmount = decimal.NewFromInt(10) // Lower to avoid conflicts
	
	result = validator.ValidateConstraintsConfig(constraints, totalInvestment)
	
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 1) // Only invalid max allocation
	assert.Equal(t, "INVALID_MAX_ALLOCATION", result.Violations[0].Type)
}

func TestConstraintValidator_ValidateConstraintsConfig_ConflictingConstraints(t *testing.T) {
	validator := NewConstraintValidator()
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(10), // 10%
		MinAllocationAmount:   decimal.NewFromInt(2000), // 20% of 10000
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateConstraintsConfig(constraints, totalInvestment)
	
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 1)
	assert.Equal(t, "CONFLICTING_CONSTRAINTS", result.Violations[0].Type)
	assert.NotEmpty(t, result.Violations[0].Suggestions)
}

func TestConstraintValidator_ValidateConstraintsConfig_HighMinAllocation(t *testing.T) {
	validator := NewConstraintValidator()
	
	constraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(100),
		MinAllocationAmount:   decimal.NewFromInt(8000), // 80% of 10000 - allows < 2 stocks
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	result := validator.ValidateConstraintsConfig(constraints, totalInvestment)
	
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 1)
	assert.Equal(t, "HIGH_MIN_ALLOCATION", result.Violations[0].Type)
	assert.NotEmpty(t, result.Violations[0].Suggestions)
}

func TestConstraintValidator_SuggestConstraintAdjustments(t *testing.T) {
	validator := NewConstraintValidator()
	
	strategies := []*models.Strategy{
		{ID: uuid.New(), Name: "Growth"},
		{ID: uuid.New(), Name: "Value"},
	}
	
	totalInvestment := decimal.NewFromInt(10000)
	
	currentConstraints := models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(20),
		MinAllocationAmount:   decimal.NewFromInt(2000), // High minimum
	}
	
	suggestions := validator.SuggestConstraintAdjustments(strategies, totalInvestment, currentConstraints)
	
	assert.NotEmpty(t, suggestions)
	// Should suggest reducing minimum allocation
	assert.Contains(t, suggestions[0], "reducing minimum allocation")
}

func TestConstraintValidator_EdgeCases(t *testing.T) {
	validator := NewConstraintValidator()
	
	// Test with empty allocations
	result := validator.ValidateAllocations([]models.StockAllocation{}, models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(50),
		MinAllocationAmount:   decimal.NewFromInt(100),
	}, decimal.NewFromInt(10000))
	
	assert.False(t, result.IsValid) // Should fail due to low allocation ratio
	assert.Len(t, result.Violations, 1)
	assert.Equal(t, "LOW_ALLOCATION_RATIO", result.Violations[0].Type)
	
	// Test with zero total investment
	result = validator.ValidateConstraintsConfig(models.AllocationConstraints{
		MaxAllocationPerStock: decimal.NewFromInt(50),
		MinAllocationAmount:   decimal.NewFromInt(100),
	}, decimal.Zero)
	
	// Should handle gracefully and return error
	assert.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Len(t, result.Violations, 1)
	assert.Equal(t, "ZERO_INVESTMENT", result.Violations[0].Type)
}