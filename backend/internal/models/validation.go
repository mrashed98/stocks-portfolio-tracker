package models

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// Validator is the global validator instance
var Validator *validator.Validate

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	return ve.Message
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string `json:"resource"`
}

// Error implements the error interface
func (nfe *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", nfe.Resource)
}

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve.Errors {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// init initializes the validator with custom validation rules
func init() {
	Validator = validator.New()
	
	// Register custom validation tags
	Validator.RegisterValidation("uppercase", validateUppercase)
	
	// Register custom type for decimal.Decimal
	Validator.RegisterCustomTypeFunc(validateDecimal, decimal.Decimal{})
}

// validateUppercase validates that a string is uppercase
func validateUppercase(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return strings.ToUpper(value) == value
}

// validateDecimal converts decimal.Decimal to float64 for validation
func validateDecimal(field reflect.Value) interface{} {
	if field.Type() == reflect.TypeOf(decimal.Decimal{}) {
		d := field.Interface().(decimal.Decimal)
		f, _ := d.Float64()
		return f
	}
	return nil
}

// ValidateStruct validates a struct and returns formatted validation errors
func ValidateStruct(s interface{}) error {
	err := Validator.Struct(s)
	if err == nil {
		return nil
	}
	
	var validationErrors []ValidationError
	
	for _, err := range err.(validator.ValidationErrors) {
		validationError := ValidationError{
			Field: err.Field(),
			Tag:   err.Tag(),
			Value: fmt.Sprintf("%v", err.Value()),
		}
		
		// Generate human-readable error messages
		switch err.Tag() {
		case "required":
			validationError.Message = fmt.Sprintf("%s is required", err.Field())
		case "email":
			validationError.Message = fmt.Sprintf("%s must be a valid email address", err.Field())
		case "min":
			validationError.Message = fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
		case "max":
			validationError.Message = fmt.Sprintf("%s must be at most %s characters long", err.Field(), err.Param())
		case "gt":
			validationError.Message = fmt.Sprintf("%s must be greater than %s", err.Field(), err.Param())
		case "gte":
			validationError.Message = fmt.Sprintf("%s must be greater than or equal to %s", err.Field(), err.Param())
		case "lt":
			validationError.Message = fmt.Sprintf("%s must be less than %s", err.Field(), err.Param())
		case "lte":
			validationError.Message = fmt.Sprintf("%s must be less than or equal to %s", err.Field(), err.Param())
		case "oneof":
			validationError.Message = fmt.Sprintf("%s must be one of: %s", err.Field(), err.Param())
		case "uppercase":
			validationError.Message = fmt.Sprintf("%s must be uppercase", err.Field())
		default:
			validationError.Message = fmt.Sprintf("%s failed validation for tag '%s'", err.Field(), err.Tag())
		}
		
		validationErrors = append(validationErrors, validationError)
	}
	
	return ValidationErrors{Errors: validationErrors}
}

// ValidateStrategyWeights validates that percentage-mode strategies don't exceed 100% total
func ValidateStrategyWeights(strategies []Strategy) error {
	var totalPercentage float64
	
	for _, strategy := range strategies {
		if strategy.WeightMode == WeightModePercent {
			weight, _ := strategy.WeightValue.Float64()
			totalPercentage += weight
		}
	}
	
	if totalPercentage > 100 {
		return fmt.Errorf("total percentage weights cannot exceed 100%%, current total: %.2f%%", totalPercentage)
	}
	
	return nil
}

// ValidateAllocationConstraints validates allocation constraints
func ValidateAllocationConstraints(constraints AllocationConstraints) error {
	if constraints.MaxAllocationPerStock.LessThanOrEqual(decimal.Zero) || 
	   constraints.MaxAllocationPerStock.GreaterThan(decimal.NewFromInt(100)) {
		return fmt.Errorf("max allocation per stock must be between 0 and 100 percent")
	}
	
	if constraints.MinAllocationAmount.LessThan(decimal.Zero) {
		return fmt.Errorf("min allocation amount must be non-negative")
	}
	
	return nil
}